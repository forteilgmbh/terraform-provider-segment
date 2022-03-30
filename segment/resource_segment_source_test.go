package segment_test

import (
	"fmt"
	segmentapi "github.com/forteilgmbh/segment-apis-go/segment"
	"github.com/forteilgmbh/terraform-provider-segment/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"strings"
	"testing"
)

func TestAccSegmentSource_basic(t *testing.T) {
	var source segmentapi.Source
	srcName := acctest.RandomWithPrefix("tf-testacc-src-basic")
	catalogName := "catalog/sources/net"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentSourceConfig_basic(srcName, catalogName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceExists("segment_source.test", &source),
					resource.TestCheckResourceAttr("segment_source.test", "source_name", srcName),
					resource.TestCheckResourceAttr("segment_source.test", "catalog_name", catalogName),
					testAccCheckSourceAttributes_basic(&source, os.Getenv("SEGMENT_WORKSPACE"), srcName, catalogName),
				),
			},
		},
	})
}

func testAccCheckSegmentSourceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*segmentapi.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "segment_source" {
			continue
		}

		_, err := client.GetSource(segment.IdToName(rs.Primary.ID))

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

func testAccCheckSourceExists(name string, source *segmentapi.Source) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("source %q not found in state", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("source %q has no ID set", name)
		}

		client := testAccProvider.Meta().(*segmentapi.Client)

		resp, err := client.GetSource(segment.IdToName(rs.Primary.ID))
		if err != nil {
			return err
		}
		*source = resp

		return nil
	}
}

func testAccCheckSourceAttributes_basic(source *segmentapi.Source, ws string, srcName string, catalogName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fullname := fmt.Sprintf("workspaces/%s/sources/%s", ws, srcName)
		if source.Name != fullname {
			return fmt.Errorf("invalid source.Name: expected: %q, actual: %q", fullname, source.Name)
		}
		if source.CatalogName != catalogName {
			return fmt.Errorf("invalid source.CatalogName: expected: %q, actual: %q", catalogName, source.CatalogName)
		}
		return nil
	}
}

func testAccSegmentSourceConfig_basic(srcName, catalogName string) string {
	return fmt.Sprintf(`
resource "segment_source" "test" {
  source_name  = %q
  catalog_name = %q
}
`, srcName, catalogName)
}
