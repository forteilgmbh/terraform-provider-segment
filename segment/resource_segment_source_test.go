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

func TestAccSegmentSource_basic(t *testing.T) {
	var source segmentapi.Source
	srcSlug := acctest.RandomWithPrefix("tf-testacc-src-basic")
	catalogName := "catalog/sources/net"
	sourceNameRegexp, _ := regexp.Compile("^workspaces/[a-z0-9._-]+/sources/[a-z0-9._-]+$")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentSourceConfig_basic(srcSlug, catalogName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceExists("segment_source.test", &source),
					resource.TestMatchResourceAttr("segment_source.test", "id", sourceNameRegexp),
					resource.TestCheckResourceAttr("segment_source.test", "slug", srcSlug),
					resource.TestCheckResourceAttr("segment_source.test", "catalog_name", catalogName),
					testAccCheckSourceAttributes_basic(&source, srcSlug, catalogName),
				),
			},
			{
				ResourceName:      "segment_source.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSegmentSource_disappears(t *testing.T) {
	var source segmentapi.Source
	srcSlug := acctest.RandomWithPrefix("tf-testacc-src-disappears")
	catalogName := "catalog/sources/net"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentSourceConfig_basic(srcSlug, catalogName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceExists("segment_source.test", &source),
					testAccCheckSourceDisappears(&source),
				),
				ExpectNonEmptyPlan: true,
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

		_, err := client.GetSource(segment.SourceNameToSlug(rs.Primary.ID))

		if err == nil {
			return fmt.Errorf("source %q still exists", rs.Primary.ID)
		}
		if segment.IsNotFoundErr(err) {
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

		resp, err := client.GetSource(segment.SourceNameToSlug(rs.Primary.ID))
		if err != nil {
			return err
		}
		*source = resp

		return nil
	}
}

func testAccCheckSourceDisappears(source *segmentapi.Source) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*segmentapi.Client)
		err := client.DeleteSource(segment.SourceNameToSlug(source.Name))
		return err
	}
}

func testAccCheckSourceAttributes_basic(source *segmentapi.Source, srcSlug string, catalogName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*segmentapi.Client)

		if source.Name != segment.SourceSlugToName(client.Workspace, srcSlug) {
			return fmt.Errorf("invalid source.Name: expected: %q, actual: %q", segment.SourceSlugToName(client.Workspace, srcSlug), source.Name)
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
  slug         = %q
  catalog_name = %q
}
`, srcName, catalogName)
}
