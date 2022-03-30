package segment_test

import (
	"fmt"
	segmentapi "github.com/forteilgmbh/segment-config-go/segment"
	"github.com/forteilgmbh/terraform-provider-segment/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccSegmentTrackingPlanSourceConnection_basic(t *testing.T) {
	srcName := acctest.RandomWithPrefix("tf-testacc-tpsc-basic")
	tpName := acctest.RandomWithPrefix("tf-testacc-tpsc-basic")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentTrackingPlanSourceConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentTrackingPlanSourceConnectionConfig_basic(srcName, tpName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackingPlanSourceConnectionExists("segment_tracking_plan_source_connection.test"),
				),
			},
		},
	})
}

func testAccCheckSegmentTrackingPlanSourceConnectionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*segmentapi.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "segment_tracking_plan_source_connection" {
			continue
		}

		planId, srcName := segment.SplitTrackingPlanSourceConnectionId(rs.Primary.ID)
		ok, err := segment.FindTrackingPlanSourceConnection(client, planId, srcName)
		if ok {
			return fmt.Errorf("tracking plan source connection %q still exists", rs.Primary.ID)
		}
		return err
	}

	return nil
}

func testAccCheckTrackingPlanSourceConnectionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("tracking plan source connection %q not found in state", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("tracking plan source connection %q has no ID set", name)
		}
		planId, srcName := segment.SplitTrackingPlanSourceConnectionId(rs.Primary.ID)

		client := testAccProvider.Meta().(*segmentapi.Client)
		ok, err := segment.FindTrackingPlanSourceConnection(client, planId, srcName)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("tracking plan source connection %q does not exist", name)
		}
		return nil
	}
}

func testAccSegmentTrackingPlanSourceConnectionConfig_basic(srcName, tpName string) string {
	return configCompose(
		testAccSegmentSourceConfig_basic(srcName, "catalog/sources/net"),
		testAccSegmentTrackingPlanConfig_identify(tpName, "identify-1-1.json"),
		`
resource "segment_tracking_plan_source_connection" "test" {
  tracking_plan_id = segment_tracking_plan.test.id
  source_name      = segment_source.test.id
}
`)
}
