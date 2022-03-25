package segment

import (
	"fmt"
	"github.com/forteilgmbh/segment-apis-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strings"
	"testing"
)

func TestAccSegmentSource_basic(t *testing.T) {
	var sourceBefore segment.Source
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceExists("segment_source.test", &sourceBefore),
					resource.TestCheckResourceAttr("segment_source.test", "source_name", rName),
					resource.TestCheckResourceAttr("segment_source.test", "catalog_name", "catalog/sources/javascript"),
				),
			},
		},
	})
}

func testAccCheckSourceExists(name string, source *segment.Source) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("source %q not found in state", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("source %q has no ID set", name)
		}

		client := testAccProvider.Meta().(*segment.Client)

		resp, err := client.GetSource(idToName(rs.Primary.ID))
		if err != nil {
			return err
		}
		*source = resp

		return nil
	}
}

func testAccCheckSegmentSourceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*segment.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "segment_source" {
			continue
		}

		_, err := client.GetSource(idToName(rs.Primary.ID))

		if err == nil {
			return fmt.Errorf("source %q still exists", rs.Primary.ID)
		}
		if strings.Contains(err.Error(), "the requested uri does not exist") {
			return nil
		}
		return err
	}

	return nil
}

func testAccSegmentSourceConfig_basic(srcName string) string {
	return fmt.Sprintf(`
resource "segment_source" "test" {
  source_name  = %q
  catalog_name = "catalog/sources/javascript"
}
`, srcName)
}
