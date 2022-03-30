package segment_test

import (
	"encoding/json"
	"fmt"
	segmentapi "github.com/forteilgmbh/segment-apis-go/segment"
	"github.com/forteilgmbh/terraform-provider-segment/segment"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"strings"
	"testing"
)

func TestAccSegmentDestination_webhook(t *testing.T) {
	resourceName := "segment_destination.test"
	srcName := acctest.RandomWithPrefix("tf-testacc-dst-webhook")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentDestinationDestroy,
		Steps: []resource.TestStep{
			testAccSegmentDestinationStep_webhook(resourceName, srcName, true, "https://example.com/api/v1"),
			testAccSegmentDestinationStep_webhook(resourceName, srcName, false, "https://example.com/api/v1"),
			testAccSegmentDestinationStep_webhook(resourceName, srcName, false, "https://example.com/api/v2"),
			testAccSegmentDestinationStep_webhook(resourceName, srcName, true, "https://example.com/api/v1"),
		},
	})
}

func testAccSegmentDestinationStep_webhook(resourceName string, srcName string, enabled bool, endpoint string) resource.TestStep {
	var destination segmentapi.Destination
	dstName := "webhooks"
	ws := os.Getenv("SEGMENT_WORKSPACE")
	configBaseName := fmt.Sprintf("workspaces/%s/sources/%s/destinations/webhooks/config/", ws, srcName)
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
		"value": "",
		"type":  "string",
	}

	return resource.TestStep{
		Config: testAccSegmentDestinationConfig_webhook(srcName, ws, enabled, endpoint),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckDestinationExists(resourceName, &destination),
			testAccCheckDestinationAttributes_webhook(&destination, ws, srcName, enabled, endpoint),
			resource.TestCheckResourceAttr(resourceName, "source_name", srcName),
			resource.TestCheckResourceAttr(resourceName, "destination_name", dstName),
			resource.TestCheckResourceAttr(resourceName, "connection_mode", "UNSPECIFIED"),
			resource.TestCheckResourceAttr(resourceName, "enabled", fmt.Sprintf("%t", enabled)),
			resource.TestCheckResourceAttr(resourceName, "configs.#", "3"),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configs.*", globalHook),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configs.*", hooks),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configs.*", sharedSecret),
		),
	}
}

func testAccCheckSegmentDestinationDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*segmentapi.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "segment_destination" {
			continue
		}
		srcName, ok := rs.Primary.Attributes["source_name"]
		if !ok {
			return fmt.Errorf("destination %q has no attribute \"source_name\"", rs.Primary.ID)
		}

		_, err := client.GetDestination(srcName, segment.IdToName(rs.Primary.ID))

		if err == nil {
			return fmt.Errorf("destination %q still exists", rs.Primary.ID)
		}
		if strings.Contains(err.Error(), "the requested uri does not exist") {
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
		srcName, ok := rs.Primary.Attributes["source_name"]
		if !ok {
			return fmt.Errorf("destination %q has no attribute \"source_name\"", name)
		}

		client := testAccProvider.Meta().(*segmentapi.Client)

		resp, err := client.GetDestination(srcName, segment.IdToName(rs.Primary.ID))
		if err != nil {
			return err
		}
		*destination = resp

		return nil
	}
}

func testAccCheckDestinationAttributes_webhook(destination *segmentapi.Destination, ws string, srcName string, enabled bool, endpoint string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fullname := fmt.Sprintf("workspaces/%s/sources/%s/destinations/webhooks", ws, srcName)
		if destination.Name != fullname {
			return fmt.Errorf("invalid destination.Name: expected: %q, actual: %q", fullname, destination.Name)
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
			return c.Name == fullname+"/config/globalHook" && c.Type == "string" && c.Value == ""
		}) {
			return fmt.Errorf("not found correct Config (globalHook) in destination.Configs: %+v", destination.Configs)
		}
		if !anyDestinationConfigValid(destination.Configs, func(c segmentapi.DestinationConfig) bool {
			return c.Name == fullname+"/config/hooks" && c.Type == "mixed" &&
				cmp.Equal(c.Value, testAccSegmentDestination_webhookConfigsHooksValue(endpoint))
		}) {
			return fmt.Errorf("not found correct Config (hooks) in destination.Configs: %+v", destination.Configs)
		}
		if !anyDestinationConfigValid(destination.Configs, func(c segmentapi.DestinationConfig) bool {
			return c.Name == fullname+"/config/sharedSecret" && c.Type == "string" && c.Value == ""
		}) {
			return fmt.Errorf("not found correct Config (sharedSecret) in destination.Configs: %+v", destination.Configs)
		}
		return nil
	}
}

func testAccSegmentDestinationConfig_webhook(srcName string, workspaceSlug string, enabled bool, endpoint string) string {
	return configCompose(
		testAccSegmentSourceConfig_basic(srcName, "catalog/sources/net"),
		fmt.Sprintf(`
resource "segment_destination" "test" {
  source_name      = segment_source.test.source_name
  destination_name = "webhooks"
  connection_mode  = "UNSPECIFIED"
  enabled          = %[2]t

  configs {
    name  = "workspaces/%[1]s/sources/${segment_source.test.source_name}/destinations/webhooks/config/globalHook"
    value = ""
    type  = "string"
  }
  configs {
    name = "workspaces/%[1]s/sources/${segment_source.test.source_name}/destinations/webhooks/config/hooks"
    value = jsonencode([
      {
        hook = %[3]q
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
    name  = "workspaces/%[1]s/sources/${segment_source.test.source_name}/destinations/webhooks/config/sharedSecret"
    value = ""
    type  = "string"
  }
}
`, workspaceSlug, enabled, endpoint),
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
