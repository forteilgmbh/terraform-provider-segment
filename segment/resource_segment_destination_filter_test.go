package segment_test

import (
	"fmt"
	segmentapi "github.com/forteilgmbh/segment-config-go/segment"
	"github.com/forteilgmbh/terraform-provider-segment/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"regexp"
	"strings"
	"testing"
)

func TestAccSegmentDestinationFilter_basic(t *testing.T) {
	var dfBefore, dfAfter, df3 segmentapi.DestinationFilter
	resourceName := "segment_destination_filter.test"
	srcSlug := acctest.RandomWithPrefix("tf-testacc-df-basic")
	dfTitle := acctest.RandomWithPrefix("tf-testacc-df-basic")
	dfIdRegexp, _ := regexp.Compile("^df_[a-zA-Z0-9]+$")
	dfNameRegexp, _ := regexp.Compile("^workspaces/[a-z0-9._-]+/sources/[a-z0-9._-]+/destinations/[a-z0-9._-]+/config/[a-f0-9]+/filters/df_[a-zA-Z0-9]+$")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentDestinationFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentDestinationFilterConfig_basic_drop(srcSlug, dfTitle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationFilterExists("segment_destination_filter.test", &dfBefore),
					testAccCheckDestinationFilterAttributes_basic(&dfBefore, resourceName, dfTitle, true, segmentapi.DestinationFilterActionTypeDropEvent),
					resource.TestMatchResourceAttr(resourceName, "id", dfIdRegexp),
					resource.TestCheckResourceAttr(resourceName, "source_slug", srcSlug),
					resource.TestCheckResourceAttr(resourceName, "destination_slug", "webhooks"),
					resource.TestMatchResourceAttr(resourceName, "name", dfNameRegexp),
					resource.TestCheckResourceAttr(resourceName, "title", dfTitle),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "conditions", "type = \"identify\""),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "action.*", map[string]string{
						"type": "drop_event",
					}),
				),
			},
			{
				Config: testAccSegmentDestinationFilterConfig_basic_sample_fieldlist(srcSlug, dfTitle, "white"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationFilterExists("segment_destination_filter.test", &dfAfter),
					testAccCheckDestinationFilterAttributes_basic(&dfAfter, resourceName, dfTitle, false, segmentapi.DestinationFilterActionTypeSampling, segmentapi.DestinationFilterActionTypeAllowList),
					resource.TestMatchResourceAttr(resourceName, "id", dfIdRegexp),
					resource.TestCheckResourceAttr(resourceName, "source_slug", srcSlug),
					resource.TestCheckResourceAttr(resourceName, "destination_slug", "webhooks"),
					resource.TestMatchResourceAttr(resourceName, "name", dfNameRegexp),
					resource.TestCheckResourceAttr(resourceName, "title", dfTitle),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "conditions", "type = \"identify\""),
					resource.TestCheckResourceAttr(resourceName, "action.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "action.*", map[string]string{
						"type":    "sample_event",
						"percent": "0.5",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "action.*", map[string]string{
						"type":                  "whitelist_fields",
						"fields.#":              "1",
						"fields.0.context.#":    "0",
						"fields.0.properties.#": "2",
						"fields.0.properties.0": "foo",
						"fields.0.properties.1": "bar",
						"fields.0.traits.#":     "1",
						"fields.0.traits.0":     "baz",
					}),
				),
			},
			{
				Config: testAccSegmentDestinationFilterConfig_basic_sample_fieldlist(srcSlug, dfTitle, "black"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationFilterExists("segment_destination_filter.test", &df3),
					testAccCheckDestinationFilterAttributes_basic(&df3, resourceName, dfTitle, false, segmentapi.DestinationFilterActionTypeSampling, segmentapi.DestinationFilterActionTypeBlockList),
					resource.TestMatchResourceAttr(resourceName, "id", dfIdRegexp),
					resource.TestCheckResourceAttr(resourceName, "source_slug", srcSlug),
					resource.TestCheckResourceAttr(resourceName, "destination_slug", "webhooks"),
					resource.TestMatchResourceAttr(resourceName, "name", dfNameRegexp),
					resource.TestCheckResourceAttr(resourceName, "title", dfTitle),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "conditions", "type = \"identify\""),
					resource.TestCheckResourceAttr(resourceName, "action.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "action.*", map[string]string{
						"type":    "sample_event",
						"percent": "0.5",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "action.*", map[string]string{
						"type":                  "blacklist_fields",
						"fields.#":              "1",
						"fields.0.context.#":    "0",
						"fields.0.properties.#": "2",
						"fields.0.properties.0": "foo",
						"fields.0.properties.1": "bar",
						"fields.0.traits.#":     "1",
						"fields.0.traits.0":     "baz",
					}),
				),
			},
			{
				ResourceName: "segment_destination_filter.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["segment_destination_filter.test"]
					if !ok {
						return "", fmt.Errorf("not found: segment_destination_filter.test")
					}
					name, ok := rs.Primary.Attributes["name"]
					if !ok {
						return "", fmt.Errorf("attribute name not set")
					}
					return name, nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSegmentDestinationFilter_disappears(t *testing.T) {
	var df segmentapi.DestinationFilter
	resourceName := "segment_destination_filter.test"
	srcSlug := acctest.RandomWithPrefix("tf-testacc-df-disappears")
	dfTitle := acctest.RandomWithPrefix("tf-testacc-df-disappears")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentDestinationFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentDestinationFilterConfig_basic_drop(srcSlug, dfTitle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationFilterExists(resourceName, &df),
					testAccCheckDestinationFilterDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSegmentDestinationFilterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*segmentapi.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "segment_destination_filter" {
			continue
		}
		id := rs.Primary.ID
		srcSlug := rs.Primary.Attributes["source_slug"]
		dstSlug := rs.Primary.Attributes["destination_slug"]

		_, err := client.GetDestinationFilter(srcSlug, dstSlug, id)

		if err == nil {
			return fmt.Errorf("destination filter %q still exists", rs.Primary.ID)
		}
		if segment.IsNotFoundErr(err) || segment.Is500ValidatePermissionsErr(err) {
			return nil
		}
		return err
	}

	return nil
}

func testAccCheckDestinationFilterExists(name string, df *segmentapi.DestinationFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("destination filter %q not found in state", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("destination filter %q has no ID set", name)
		}
		client := testAccProvider.Meta().(*segmentapi.Client)

		id := rs.Primary.ID
		srcSlug := rs.Primary.Attributes["source_slug"]
		dstSlug := rs.Primary.Attributes["destination_slug"]

		resp, err := client.GetDestinationFilter(srcSlug, dstSlug, id)
		if err != nil {
			return err
		}
		*df = *resp

		return nil
	}
}

func testAccCheckDestinationFilterDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[name]
		id := rs.Primary.ID
		srcSlug := rs.Primary.Attributes["source_slug"]
		dstSlug := rs.Primary.Attributes["destination_slug"]
		client := testAccProvider.Meta().(*segmentapi.Client)
		return client.DeleteDestinationFilter(srcSlug, dstSlug, id)
	}
}

func testAccCheckDestinationFilterAttributes_basic(df *segmentapi.DestinationFilter, name string, title string, enabled bool, actionTypes ...segmentapi.DestinationFilterActionType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[name]

		if !strings.HasSuffix(df.Name, rs.Primary.ID) {
			return fmt.Errorf("invalid ID: expected: %q, actual: %q", rs.Primary.ID, df.Name)
		}
		if df.Title != title {
			return fmt.Errorf("invalid title: expected: %q, actual: %q", title, df.Title)
		}
		if df.Description != "" {
			return fmt.Errorf("invalid description: expected: %q, actual: %q", "", df.Description)
		}
		if df.IsEnabled != enabled {
			return fmt.Errorf("invalid isEnabled: expected: %t, actual: %t", true, df.IsEnabled)
		}
		if df.Conditions != "type = \"identify\"" {
			return fmt.Errorf("invalid conditions: expected: %q, actual: %q", "type = \"identify\"", df.Conditions)
		}
		if err := testAccCheckDestinationFilterAttributesActions(df, actionTypes...); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDestinationFilterAttributesActions(df *segmentapi.DestinationFilter, actionTypes ...segmentapi.DestinationFilterActionType) error {
	if len(df.Actions) != len(actionTypes) {
		return fmt.Errorf("invalid number of actions: expected: %d, actual: %d", len(actionTypes), len(df.Actions))
	}
	for _, at := range actionTypes {
		if err := validateDestinationFilterActions(df.Actions, at); err != nil {
			return err
		}
	}
	return nil
}

func validateDestinationFilterActions(actions segmentapi.DestinationFilterActions, expectedType segmentapi.DestinationFilterActionType) error {
	for _, a := range actions {
		if a.ActionType() == expectedType {
			return validateDestinationFilterAction(a, expectedType)
		}
	}
	return fmt.Errorf("expected action type %q not found", expectedType)
}

func validateDestinationFilterAction(actual segmentapi.DestinationFilterAction, expectedType segmentapi.DestinationFilterActionType) error {
	if actual.ActionType() != expectedType {
		return fmt.Errorf("invalid action type: expected: %s, actual: %s", expectedType, actual.ActionType())
	}
	switch a := actual.(type) {
	case segmentapi.SamplingEventAction:
		if a.Percent != 0.5 {
			return fmt.Errorf("invalid percent: expected: %f, actual: %f", 0.5, a.Percent)
		}
		if a.Path != "" {
			return fmt.Errorf("invalid path: expected: %s, actual: %s", "", a.Path)
		}
	case segmentapi.FieldsListEventAction:
		if a.Fields.Context != nil && len(a.Fields.Context.Fields) > 0 {
			return fmt.Errorf("invalid context fields length: expected: %d, actual: %d", 0, len(a.Fields.Context.Fields))
		}
		if a.Fields.Properties == nil {
			return fmt.Errorf("invalid properties fields length: expected: %d, actual: %d", 2, 0)
		} else if len(a.Fields.Properties.Fields) != 2 {
			return fmt.Errorf("invalid properties fields length: expected: %d, actual: %d", 2, len(a.Fields.Properties.Fields))
		}
		if a.Fields.Traits == nil {
			return fmt.Errorf("invalid traits fields length: expected: %d, actual: %d", 1, 0)
		} else if len(a.Fields.Traits.Fields) != 1 {
			return fmt.Errorf("invalid traits fields length: expected: %d, actual: %d", 1, len(a.Fields.Traits.Fields))
		}
		if !segment.Contains("foo", a.Fields.Properties.Fields) {
			return fmt.Errorf("not found required element %q in properties fields: %v", "foo", a.Fields.Properties.Fields)
		}
		if !segment.Contains("bar", a.Fields.Properties.Fields) {
			return fmt.Errorf("not found required element %q in properties fields: %v", "bar", a.Fields.Properties.Fields)
		}
		if !segment.Contains("baz", a.Fields.Traits.Fields) {
			return fmt.Errorf("not found required element %q in properties fields: %v", "baz", a.Fields.Traits.Fields)
		}
	}
	return nil
}

func testAccSegmentDestinationFilterConfig_basic_drop(srcSlug, dfTitle string) string {
	return configCompose(
		testAccSegmentDestinationConfig_webhook(srcSlug, true, "https://example.com/api/v1"),
		fmt.Sprintf(`
resource "segment_destination_filter" "test" {
  title = %q

  source_slug      = segment_source.test.slug
  destination_slug = segment_destination.test.slug
  
  conditions = "type = \"identify\""

  action {
    type = "drop_event"
  }
}
`, dfTitle))
}

func testAccSegmentDestinationFilterConfig_basic_sample_fieldlist(srcSlug, dfTitle, action string) string {
	return configCompose(
		testAccSegmentDestinationConfig_webhook(srcSlug, true, "https://example.com/api/v1"),
		fmt.Sprintf(`
resource "segment_destination_filter" "test" {
  title = %q

  source_slug      = segment_source.test.slug
  destination_slug = segment_destination.test.slug
  
  enabled    = false
  conditions = "type = \"identify\""

  action {
    type    = "sample_event"
    percent = 0.5
  }

  action {
    type = "%slist_fields"
    
    fields {
      properties = ["foo", "bar"]
      traits     = ["baz"]
    }
  }
}
`, dfTitle, action))
}
