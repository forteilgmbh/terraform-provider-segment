package segment_test

import (
	"encoding/json"
	"fmt"
	segmentapi "github.com/forteilgmbh/segment-config-go/segment"
	"github.com/forteilgmbh/terraform-provider-segment/segment"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccSegmentDestination_webhook(t *testing.T) {
	resourceName := "segment_destination.test"
	srcSlug := acctest.RandomWithPrefix("tf-testacc-dst-webhook")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentDestinationDestroy,
		Steps: []resource.TestStep{
			testAccSegmentDestinationStep_webhook(resourceName, srcSlug, true, "https://example.com/api/v1"),
			testAccSegmentDestinationStep_webhook(resourceName, srcSlug, false, "https://example.com/api/v1"),
			testAccSegmentDestinationStep_webhook(resourceName, srcSlug, false, "https://example.com/api/v2"),
			testAccSegmentDestinationStep_webhook(resourceName, srcSlug, true, "https://example.com/api/v1"),
			{
				ResourceName:      "segment_destination.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSegmentDestinationStep_webhook(resourceName string, srcSlug string, enabled bool, endpoint string) resource.TestStep {
	var destination segmentapi.Destination
	slug := "webhooks"

	return resource.TestStep{
		Config: testAccSegmentDestinationConfig_webhook(srcSlug, enabled, endpoint),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckDestinationExists(resourceName, &destination),
			testAccCheckDestinationAttributes_webhook(resourceName, &destination, enabled, endpoint),
			resource.TestCheckResourceAttr(resourceName, "slug", slug),
			resource.TestCheckResourceAttr(resourceName, "source_slug", srcSlug),
			resource.TestCheckResourceAttr(resourceName, "connection_mode", "UNSPECIFIED"),
			resource.TestCheckResourceAttr(resourceName, "enabled", fmt.Sprintf("%t", enabled)),
			testAccCheckDestinationConfigs_webhook(resourceName, srcSlug, endpoint),
		),
	}
}

func TestAccSegmentDestination_disappears(t *testing.T) {
	var destination segmentapi.Destination
	srcSlug := acctest.RandomWithPrefix("tf-testacc-dst-disappears")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentDestinationConfig_webhook(srcSlug, true, "https://example.com/api/v1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationExists("segment_destination.test", &destination),
					testAccCheckDestinationDisappears(&destination),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSegmentDestinationDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*segmentapi.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "segment_destination" {
			continue
		}
		slug := segment.DestinationNameToSlug(rs.Primary.ID)
		srcSlug := segment.DestinationNameToSourceSlug(rs.Primary.ID)

		_, err := client.GetDestination(srcSlug, slug)

		if err == nil {
			return fmt.Errorf("destination %q still exists", rs.Primary.ID)
		}
		if segment.IsNotFoundErr(err) {
			return nil
		}
		return err
	}

	return nil
}

func testAccCheckDestinationExists(name string, destination *segmentapi.Destination) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("destination %q not found in state", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("destination %q has no ID set", name)
		}
		client := testAccProvider.Meta().(*segmentapi.Client)

		slug := segment.DestinationNameToSlug(rs.Primary.ID)
		srcSlug := segment.DestinationNameToSourceSlug(rs.Primary.ID)

		resp, err := client.GetDestination(srcSlug, slug)
		if err != nil {
			return err
		}
		*destination = resp

		return nil
	}
}

func testAccCheckDestinationDisappears(destination *segmentapi.Destination) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*segmentapi.Client)

		slug := segment.DestinationNameToSlug(destination.Name)
		srcSlug := segment.DestinationNameToSourceSlug(destination.Name)

		return client.DeleteDestination(srcSlug, slug)
	}
}

func testAccCheckDestinationAttributes_webhook(name string, destination *segmentapi.Destination, enabled bool, endpoint string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[name]

		if destination.Name != rs.Primary.ID {
			return fmt.Errorf("invalid destination.Name: expected: %q, actual: %q", rs.Primary.ID, destination.Name)
		}
		if destination.ConnectionMode != "UNSPECIFIED" {
			return fmt.Errorf("invalid destination.Name: expected: %q, actual: %q", "UNSPECIFIED", destination.ConnectionMode)
		}
		if destination.Enabled != enabled {
			return fmt.Errorf("invalid destination.Enabled: expected: %v actual: %v", enabled, destination.Enabled)
		}
		if len(destination.Configs) != 3 {
			return fmt.Errorf("invalid size of destination.Configs: expected: %d actual: %d", 3, len(destination.Configs))
		}
		if !anyDestinationConfigValid(destination.Configs, func(c segmentapi.DestinationConfig) bool {
			return c.Name == rs.Primary.ID+"/config/globalHook" && c.Type == "string" && c.Value == ""
		}) {
			return fmt.Errorf("not found correct Config (globalHook) in destination.Configs: %+v", destination.Configs)
		}
		if !anyDestinationConfigValid(destination.Configs, func(c segmentapi.DestinationConfig) bool {
			return c.Name == rs.Primary.ID+"/config/hooks" && c.Type == "mixed" &&
				cmp.Equal(c.Value, testAccSegmentDestination_webhookConfigsHooksValue(endpoint))
		}) {
			return fmt.Errorf("not found correct Config (hooks) in destination.Configs: %+v", destination.Configs)
		}
		if !anyDestinationConfigValid(destination.Configs, func(c segmentapi.DestinationConfig) bool {
			return c.Name == rs.Primary.ID+"/config/sharedSecret" && c.Type == "string" && c.Value == "secretValue"
		}) {
			return fmt.Errorf("not found correct Config (sharedSecret) in destination.Configs: %+v", destination.Configs)
		}
		return nil
	}
}

func testAccCheckDestinationConfigs_webhook(resourceName, srcSlug, endpoint string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*segmentapi.Client)

		configBaseName := fmt.Sprintf("workspaces/%s/sources/%s/destinations/webhooks/config/", client.Workspace, srcSlug)
		globalHook := map[string]string{
			"name":  configBaseName + "globalHook",
			"value": "",
			"type":  "string",
		}
		hooks := map[string]string{
			"name":  configBaseName + "hooks",
			"value": toJsonString(testAccSegmentDestination_webhookConfigsHooksValue(endpoint)),
			"type":  "mixed",
		}
		sharedSecret := map[string]string{
			"name":  configBaseName + "sharedSecret",
			"value": "secretValue",
			"type":  "string",
		}

		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "configs.#", "3"),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configs.*", globalHook),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configs.*", hooks),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configs.*", sharedSecret),
		)(s)
	}
}

func testAccSegmentDestinationConfig_webhook(srcSlug string, enabled bool, endpoint string) string {
	return configCompose(
		testAccSegmentSourceConfig_basic(srcSlug, "catalog/sources/net"),
		fmt.Sprintf(`
resource "segment_destination" "test" {
  slug             = "webhooks"  
  source_slug      = segment_source.test.slug
  connection_mode  = "UNSPECIFIED"
  enabled          = %t

  configs {
    name  = "${segment_source.test.id}/destinations/webhooks/config/globalHook"
    value = ""
    type  = "string"
  }
  configs {
    name = "${segment_source.test.id}/destinations/webhooks/config/hooks"
    value = jsonencode([
      {
        hook = %q
        headers = [
          {
            "key"   = "Authorization"
            "value" = "Basic d2h5OmFyZXlvdWV2ZW5kZWNvZGluZ3RoaXM/Pz8="
          }
        ]
      }
    ])
    type = "mixed"
  }
  configs {
    name  = "${segment_source.test.id}/destinations/webhooks/config/sharedSecret"
    value = "secretValue"
    type  = "string"
  }
}
`, enabled, endpoint),
	)
}

func testAccSegmentDestination_webhookConfigsHooksValue(endpoint string) interface{} {
	h := map[string]interface{}{
		"key":   "Authorization",
		"value": "Basic d2h5OmFyZXlvdWV2ZW5kZWNvZGluZ3RoaXM/Pz8=",
	}
	hs := []interface{}{h}
	m := map[string]interface{}{
		"hook":    endpoint,
		"headers": hs,
	}
	return []interface{}{m}
}

func anyDestinationConfigValid(configs []segmentapi.DestinationConfig, predicate func(c segmentapi.DestinationConfig) bool) bool {
	for _, c := range configs {
		if predicate(c) {
			return true
		}
	}
	return false
}

func toJsonString(v interface{}) string {
	s, _ := json.Marshal(v)
	return string(s)
}
