package segment_test

import (
	"fmt"
	segmentapi "github.com/forteilgmbh/segment-config-go/segment"
	"github.com/forteilgmbh/terraform-provider-segment/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"regexp"
	"testing"
)

func TestAccSegmentSourceSchemaConfig_basic(t *testing.T) {
	var schemaConfigBefore, schemaConfigAfter segmentapi.SourceConfig
	srcSlug := acctest.RandomWithPrefix("tf-testacc-srcschema-basic")
	srcViolationsSlug := acctest.RandomWithPrefix("tf-testacc-srcschema-basic-violations-dest")
	nameRegexp, _ := regexp.Compile("^workspaces/[a-z0-9._-]+/sources/[a-z0-9._-]+/schema-config$")
	resourceName := "segment_source_schema_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentSourceSchemaConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentSourceSchemaConfigConfig_basic(srcSlug, srcViolationsSlug, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceSchemaConfigExists(resourceName, &schemaConfigBefore),
					resource.TestMatchResourceAttr(resourceName, "id", nameRegexp),
					resource.TestCheckResourceAttr(resourceName, "source_slug", srcSlug),
					testAccCheckSourceSchemaConfigResource_basic(resourceName, srcViolationsSlug, true),
					testAccCheckSourceSchemaConfigAttributes_basic(&schemaConfigBefore, srcViolationsSlug, true),
				),
			},
			{
				Config: testAccSegmentSourceSchemaConfigConfig_basic(srcSlug, srcViolationsSlug, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceSchemaConfigExists(resourceName, &schemaConfigAfter),
					resource.TestMatchResourceAttr(resourceName, "id", nameRegexp),
					resource.TestCheckResourceAttr(resourceName, "source_slug", srcSlug),
					testAccCheckSourceSchemaConfigResource_basic(resourceName, srcViolationsSlug, false),
					testAccCheckSourceSchemaConfigAttributes_basic(&schemaConfigAfter, srcViolationsSlug, false),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSegmentSourceSchemaConfig_disappears(t *testing.T) {
	var schemaConfig segmentapi.SourceConfig
	srcSlug := acctest.RandomWithPrefix("tf-testacc-srcschema-disappears")
	srcViolationsSlug := acctest.RandomWithPrefix("tf-testacc-srcschema-disappears-violations-dest")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentSourceSchemaConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentSourceSchemaConfigConfig_basic(srcSlug, srcViolationsSlug, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceSchemaConfigExists("segment_source_schema_config.test", &schemaConfig),
					testAccCheckSourceSchemaConfigDisappears(&schemaConfig),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSegmentSourceSchemaConfigDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*segmentapi.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "segment_source_schema_config" {
			continue
		}

		c, err := client.GetSourceConfig(segment.SourceNameToSlug(rs.Primary.ID))

		if err == nil {
			if c == segment.DefaultSegmentSourceSchemaConfig {
				return nil
			} else {
				return fmt.Errorf("source schema config %q still exists", rs.Primary.ID)
			}
		} else {
			if segment.IsNotFoundErr(err) {
				return nil
			} else {
				return fmt.Errorf("source schema config %q still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckSourceSchemaConfigExists(name string, schemaConfig *segmentapi.SourceConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("source schema config %q not found in state", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("source schema config %q has no ID set", name)
		}

		client := testAccProvider.Meta().(*segmentapi.Client)

		resp, err := client.GetSourceConfig(segment.SourceNameToSlug(rs.Primary.ID))
		if err != nil {
			return err
		}
		*schemaConfig = resp

		return nil
	}
}

func testAccCheckSourceSchemaConfigDisappears(schemaConfig *segmentapi.SourceConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*segmentapi.Client)
		err := client.DeleteSource(segment.SourceNameToSlug(schemaConfig.Name)) // not a mistake - we want to check the case when entire source is deleted
		return err
	}
}

func testAccCheckSourceSchemaConfigResource_basic(resourceName, srcViolationsSlug string, flag bool) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "allow_unplanned_track_events", "false"),
		resource.TestCheckResourceAttr(resourceName, "allow_unplanned_identify_traits", fmt.Sprintf("%t", segment.DefaultSourceSchemaConfig["allow_unplanned_identify_traits"].(bool))),
		resource.TestCheckResourceAttr(resourceName, "allow_unplanned_group_traits", fmt.Sprintf("%t", segment.DefaultSourceSchemaConfig["allow_unplanned_group_traits"].(bool))),
		resource.TestCheckResourceAttr(resourceName, "forwarding_blocked_events_to", segment.DefaultSourceSchemaConfig["forwarding_blocked_events_to"].(string)),
		resource.TestCheckResourceAttr(resourceName, "allow_unplanned_track_event_properties", fmt.Sprintf("%t", segment.DefaultSourceSchemaConfig["allow_unplanned_track_event_properties"].(bool))),
		resource.TestCheckResourceAttr(resourceName, "allow_track_event_on_violations", fmt.Sprintf("%t", flag)),
		resource.TestCheckResourceAttr(resourceName, "allow_identify_traits_on_violations", fmt.Sprintf("%t", segment.DefaultSourceSchemaConfig["allow_identify_traits_on_violations"].(bool))),
		resource.TestCheckResourceAttr(resourceName, "allow_group_traits_on_violations", fmt.Sprintf("%t", !flag)),
		resource.TestCheckResourceAttr(resourceName, "forwarding_violations_to", srcViolationsSlug),
		resource.TestCheckResourceAttr(resourceName, "allow_track_properties_on_violations", fmt.Sprintf("%t", segment.DefaultSourceSchemaConfig["allow_track_properties_on_violations"].(bool))),
		resource.TestCheckResourceAttr(resourceName, "common_track_event_on_violations", "ALLOW"),
		resource.TestCheckResourceAttr(resourceName, "common_identify_event_on_violations", segment.DefaultSourceSchemaConfig["common_identify_event_on_violations"].(string)),
		resource.TestCheckResourceAttr(resourceName, "common_group_event_on_violations", segment.DefaultSourceSchemaConfig["common_group_event_on_violations"].(string)),
	)
}

func testAccCheckSourceSchemaConfigAttributes_basic(schemaConfig *segmentapi.SourceConfig, srcViolationsSlug string, flag bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if schemaConfig.AllowUnplannedTrackEvents != false {
			return fmt.Errorf("invalid AllowUnplannedTrackEvents: expected: %t, actual: %t", false, schemaConfig.AllowUnplannedTrackEvents)
		}
		if schemaConfig.AllowUnplannedIdentifyTraits != segment.DefaultSegmentSourceSchemaConfig.AllowUnplannedIdentifyTraits {
			return fmt.Errorf("invalid AllowUnplannedIdentifyTraits: expected: %t, actual: %t", segment.DefaultSegmentSourceSchemaConfig.AllowUnplannedIdentifyTraits, schemaConfig.AllowUnplannedIdentifyTraits)
		}
		if schemaConfig.AllowUnplannedGroupTraits != segment.DefaultSegmentSourceSchemaConfig.AllowUnplannedGroupTraits {
			return fmt.Errorf("invalid AllowUnplannedGroupTraits: expected: %t, actual: %t", segment.DefaultSegmentSourceSchemaConfig.AllowUnplannedGroupTraits, schemaConfig.AllowUnplannedGroupTraits)
		}
		if schemaConfig.ForwardingBlockedEventsTo != segment.DefaultSegmentSourceSchemaConfig.ForwardingBlockedEventsTo {
			return fmt.Errorf("invalid ForwardingBlockedEventsTo: expected: %s, actual: %s", segment.DefaultSegmentSourceSchemaConfig.ForwardingBlockedEventsTo, schemaConfig.ForwardingBlockedEventsTo)
		}
		if schemaConfig.AllowUnplannedTrackEventsProperties != segment.DefaultSegmentSourceSchemaConfig.AllowUnplannedTrackEventsProperties {
			return fmt.Errorf("invalid AllowUnplannedTrackEventsProperties: expected: %t, actual: %t", segment.DefaultSegmentSourceSchemaConfig.AllowUnplannedTrackEventsProperties, schemaConfig.AllowUnplannedTrackEventsProperties)
		}
		if schemaConfig.AllowTrackEventOnViolations != flag {
			return fmt.Errorf("invalid AllowTrackEventOnViolations: expected: %t, actual: %t", flag, schemaConfig.AllowTrackEventOnViolations)
		}
		if schemaConfig.AllowIdentifyTraitsOnViolations != segment.DefaultSegmentSourceSchemaConfig.AllowIdentifyTraitsOnViolations {
			return fmt.Errorf("invalid AllowIdentifyTraitsOnViolations: expected: %t, actual: %t", segment.DefaultSegmentSourceSchemaConfig.AllowIdentifyTraitsOnViolations, schemaConfig.AllowIdentifyTraitsOnViolations)
		}
		if schemaConfig.AllowGroupTraitsOnViolations != !flag {
			return fmt.Errorf("invalid AllowGroupTraitsOnViolations: expected: %t, actual: %t", !flag, schemaConfig.AllowGroupTraitsOnViolations)
		}
		if schemaConfig.ForwardingViolationsTo != srcViolationsSlug {
			return fmt.Errorf("invalid ForwardingViolationsTo: expected: %s, actual: %s", srcViolationsSlug, schemaConfig.ForwardingViolationsTo)
		}
		if schemaConfig.AllowTrackPropertiesOnViolations != segment.DefaultSegmentSourceSchemaConfig.AllowTrackPropertiesOnViolations {
			return fmt.Errorf("invalid AllowTrackPropertiesOnViolations: expected: %t, actual: %t", segment.DefaultSegmentSourceSchemaConfig.AllowTrackPropertiesOnViolations, schemaConfig.AllowTrackPropertiesOnViolations)
		}
		if schemaConfig.CommonTrackEventOnViolations != segmentapi.Allow {
			return fmt.Errorf("invalid CommonTrackEventOnViolations: expected: %s, actual: %s", segmentapi.Allow, schemaConfig.CommonTrackEventOnViolations)
		}
		if schemaConfig.CommonIdentifyEventOnViolations != segment.DefaultSegmentSourceSchemaConfig.CommonIdentifyEventOnViolations {
			return fmt.Errorf("invalid CommonIdentifyEventOnViolations: expected: %s, actual: %s", segment.DefaultSegmentSourceSchemaConfig.CommonIdentifyEventOnViolations, schemaConfig.CommonIdentifyEventOnViolations)
		}
		if schemaConfig.CommonGroupEventOnViolations != segment.DefaultSegmentSourceSchemaConfig.CommonGroupEventOnViolations {
			return fmt.Errorf("invalid CommonGroupEventOnViolations: expected: %s, actual: %s", segment.DefaultSegmentSourceSchemaConfig.CommonGroupEventOnViolations, schemaConfig.CommonGroupEventOnViolations)
		}
		return nil
	}
}

func testAccSegmentSourceSchemaConfigConfig_basic(srcSlug, srcViolationsSlug string, flag bool) string {
	return configCompose(
		testAccSegmentTrackingPlanSourceConnectionConfig_basic(srcSlug, srcSlug),
		fmt.Sprintf(`
resource "segment_source_schema_config" "test" {
  source_slug = segment_source.test.slug

  allow_unplanned_track_events     = false
  allow_track_event_on_violations  = %t
  allow_group_traits_on_violations = %t
  forwarding_violations_to         = segment_source.violations_destination.slug
  common_track_event_on_violations = "ALLOW"

  depends_on = [segment_tracking_plan_source_connection.test]
}

resource "segment_source" "violations_destination" {
  slug         = %q
  catalog_name = "catalog/sources/net"
}
`, flag, !flag, srcViolationsSlug),
	)
}
