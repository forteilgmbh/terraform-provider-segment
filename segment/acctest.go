package segment

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"testing"
)

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("SEGMENT_WORKSPACE"); v == "" {
		t.Fatal("SEGMENT_WORKSPACE must be set for acceptance tests")
	}
	if v := os.Getenv("SEGMENT_ACCESS_TOKEN"); v == "" {
		t.Fatal("SEGMENT_ACCESS_TOKEN must be set for acceptance tests")
	}
}

var testAccProviders map[string]func() (*schema.Provider, error)
var testAccProvider = Provider()

func init() {
	testAccProviders = map[string]func() (*schema.Provider, error){
		"segment": func() (*schema.Provider, error) { return testAccProvider, nil },
	}
}
