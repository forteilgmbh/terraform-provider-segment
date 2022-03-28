package segment_test

import (
	"github.com/forteilgmbh/terraform-provider-segment/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"strings"
	"testing"
)

func TestProvider(t *testing.T) {
	if err := segment.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("SEGMENT_WORKSPACE"); v == "" {
		t.Fatal("SEGMENT_WORKSPACE must be set for acceptance tests")
	}
	if v := os.Getenv("SEGMENT_ACCESS_TOKEN"); v == "" {
		t.Fatal("SEGMENT_ACCESS_TOKEN must be set for acceptance tests")
	}
}

var testAccProviders map[string]func() (*schema.Provider, error)
var testAccProvider = segment.Provider()

func init() {
	testAccProviders = map[string]func() (*schema.Provider, error){
		"segment": func() (*schema.Provider, error) { return testAccProvider, nil },
	}
}

// configCompose can be called to concatenate multiple strings to build test configurations
func configCompose(config ...string) string {
	var str strings.Builder

	for _, conf := range config {
		str.WriteString(conf)
	}

	return str.String()
}
