package segment_test

import (
	"embed"
	"encoding/json"
	"fmt"
	segmentapi "github.com/forteilgmbh/segment-config-go/segment"
	"github.com/forteilgmbh/terraform-provider-segment/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strings"
	"testing"
)

// entire resource managed solely with Terraform
func TestAccSegmentTrackingPlan_authoritative(t *testing.T) {
	var tp segmentapi.TrackingPlan
	rName := acctest.RandomWithPrefix("tf-testacc-tp-auth")
	resourceName := "segment_tracking_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentTrackingPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentTrackingPlanConfig_identify(rName, "identify-1-1.json"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackingPlanExists(resourceName, &tp),
					testAccCheckTrackingPlanAttributes(&tp, rName, "identify-1-1.json", []string{}),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
					resource.TestCheckResourceAttr(resourceName, "rules_identify", ruleStringFromFile("identify-1-1.json")),
				),
			},
			{
				Config: testAccSegmentTrackingPlanConfig_identify_events(rName, "identify-1-2.json", []string{"event-1-1.json", "event-2-1.json"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackingPlanExists(resourceName, &tp),
					testAccCheckTrackingPlanAttributes(&tp, rName, "identify-1-2.json", []string{"event-1-1.json", "event-2-1.json"}),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
					resource.TestCheckResourceAttr(resourceName, "rules_identify", ruleStringFromFile("identify-1-2.json")),
					resource.TestCheckResourceAttr(resourceName, "rules_events.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rules_events.0", eventStringFromFile("event-1-1.json")),
					resource.TestCheckResourceAttr(resourceName, "rules_events.1", eventStringFromFile("event-2-1.json")),
				),
			},
			{
				Config: testAccSegmentTrackingPlanConfig_identify_events(rName, "identify-1-2.json", []string{"event-1-1.json"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackingPlanExists(resourceName, &tp),
					testAccCheckTrackingPlanAttributes(&tp, rName, "identify-1-2.json", []string{"event-1-1.json"}),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
					resource.TestCheckResourceAttr(resourceName, "rules_identify", ruleStringFromFile("identify-1-2.json")),
					resource.TestCheckResourceAttr(resourceName, "rules_events.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules_events.0", eventStringFromFile("event-1-1.json")),
				),
			},
			{
				ResourceName:            "segment_tracking_plan.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rules_global", "rules_identify", "rules_group", "rules_events"},
				// Rules cannot be imported as the resource attributes determine what kinds of rules should be managed
				// by Terraform (attribute is set) and what should be left without changes (attribute is null).
				// As a consequence, apply is required after import.
			},
		},
	})
}

// some rules managed with Terraform, some managed outside
func TestAccSegmentTrackingPlan_nonauthoritative(t *testing.T) {
	var tp segmentapi.TrackingPlan
	rName := acctest.RandomWithPrefix("tf-testacc-tp-nonauth")
	resourceName := "segment_tracking_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentTrackingPlanDestroy,
		Steps: []resource.TestStep{
			// regular setup
			{
				Config: testAccSegmentTrackingPlanConfig_identify(rName, "identify-1-1.json"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackingPlanExists(resourceName, &tp),
					testAccCheckTrackingPlanAttributes(&tp, rName, "identify-1-1.json", []string{}),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
					resource.TestCheckResourceAttr(resourceName, "rules_identify", ruleStringFromFile("identify-1-1.json")),
					testAccUpdateTrackingPlan(resourceName, &tp, []string{"event-1-1.json", "event-2-1.json"}),
				),
			},
			// modify the tracking plan: add "events" rules but there should be no changes afterwards
			{
				Config: testAccSegmentTrackingPlanConfig_identify(rName, "identify-1-1.json"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackingPlanExists(resourceName, &tp),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
					resource.TestCheckResourceAttr(resourceName, "rules_identify", ruleStringFromFile("identify-1-1.json")),
					// "events" are set in the actual tracking plan
					testAccCheckTrackingPlanAttributes(&tp, rName, "identify-1-1.json", []string{"event-1-1.json", "event-2-1.json"}),
					// but they are not managed in state
					resource.TestCheckResourceAttr(resourceName, "rules_events.#", "0"),
				),
			},
			// the unmanaged rules should not be modified while modifying the managed ones
			{
				Config: testAccSegmentTrackingPlanConfig_identify(rName, "identify-1-2.json"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackingPlanExists(resourceName, &tp),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
					// "identify" are updated by Terraform,
					resource.TestCheckResourceAttr(resourceName, "rules_identify", ruleStringFromFile("identify-1-2.json")),
					// while "events" are still, unmodified, in the actual tracking plan
					testAccCheckTrackingPlanAttributes(&tp, rName, "identify-1-2.json", []string{"event-1-1.json", "event-2-1.json"}),
					// and still not managed in state
					resource.TestCheckResourceAttr(resourceName, "rules_events.#", "0"),
				),
			},
			{
				ResourceName:            "segment_tracking_plan.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rules_global", "rules_identify", "rules_group", "rules_events"},
				// Rules cannot be imported as the resource attributes determine what kinds of rules should be managed
				// by Terraform (attribute is set) and what should be left without changes (attribute is null).
				// As a consequence, apply is required after import.
			},
		},
	})
}

func TestAccSegmentTrackingPlan_disappears(t *testing.T) {
	var tp segmentapi.TrackingPlan
	rName := acctest.RandomWithPrefix("tf-testacc-tp-disappears")
	resourceName := "segment_tracking_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSegmentTrackingPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentTrackingPlanConfig_identify(rName, "identify-1-2.json"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackingPlanExists(resourceName, &tp),
					testAccCheckTrackingPlanDisappears(&tp),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSegmentTrackingPlanDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*segmentapi.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "segment_tracking_plan" {
			continue
		}

		_, err := client.GetTrackingPlan(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("tracking plan %q still exists", rs.Primary.ID)
		}
		if strings.Contains(err.Error(), "the requested uri does not exist") {
			return nil
		}
		return err
	}

	return nil
}

func testAccCheckTrackingPlanExists(name string, tp *segmentapi.TrackingPlan) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("tracking plan %q not found in state", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("tracking plan %q has no ID set", name)
		}

		client := testAccProvider.Meta().(*segmentapi.Client)

		resp, err := client.GetTrackingPlan(rs.Primary.ID)
		if err != nil {
			return err
		}
		*tp = resp

		return nil
	}
}

func testAccCheckTrackingPlanDisappears(tp *segmentapi.TrackingPlan) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*segmentapi.Client)
		return client.DeleteTrackingPlan(segment.TrackingPlanNameToId(tp.Name))
	}
}

func testAccCheckTrackingPlanAttributes(tp *segmentapi.TrackingPlan, displayName string, identifyFile string, eventsFiles []string) resource.TestCheckFunc {
	identify := ruleStringFromFile(identifyFile)
	return func(s *terraform.State) error {
		if tp.DisplayName != displayName {
			return fmt.Errorf("invalid displayName: expected: %q, actual: %q", displayName, tp.DisplayName)
		}
		if !segment.IsNilOrZeroValue(tp.Rules.Global) {
			return fmt.Errorf("non empty Global rules: %+v", tp.Rules.Global)
		}
		if !segment.IsNilOrZeroValue(tp.Rules.Group) {
			return fmt.Errorf("non empty Group rules: %+v", tp.Rules.Group)
		}
		if identify == "" && !segment.IsNilOrZeroValue(tp.Rules.Identify) {
			return fmt.Errorf("non empty Identify rules: %+v", tp.Rules.Identify)
		}
		if identify != "" && toPrettyJsonString(tp.Rules.Identify) != identify {
			return fmt.Errorf("invalid Identify rules: expected: %s, actual: %s", identify, toPrettyJsonString(tp.Rules.Identify))
		}
		if len(eventsFiles) == 0 && !(tp.Rules.Events == nil || len(tp.Rules.Events) == 0) {
			return fmt.Errorf("non empty Events rules: %+v", tp.Rules.Events)
		}
		if len(eventsFiles) > 0 {
			if len(tp.Rules.Events) != len(eventsFiles) {
				return fmt.Errorf("invalid numver of Events rules: expected: %d, actual: %d (%+v)", len(eventsFiles), len(tp.Rules.Events), tp.Rules.Events)
			}
			for i := range eventsFiles {
				exp := eventStringFromFile(eventsFiles[i])
				act := toPrettyJsonString(tp.Rules.Events[i])
				if act != exp {
					return fmt.Errorf("invalid Event.%d rule: expected: %s, actual: %s", i, exp, act)
				}
			}
		}
		return nil
	}
}

func testAccSegmentTrackingPlanConfig_identify(rName, rulesFile string) string {
	return fmt.Sprintf(`
resource "segment_tracking_plan" "test" {
  display_name = %q
  
  rules_identify = <<-EOF
%s
EOF
}
`, rName, ruleStringFromFile(rulesFile))
}

func testAccSegmentTrackingPlanConfig_identify_events(rName, identifyFile string, eventsFiles []string) string {
	events := make([]string, 0, len(eventsFiles))
	for _, f := range eventsFiles {
		events = append(events, fmt.Sprintf("<<-EOF\n%s\nEOF\n", eventStringFromFile(f)))
	}
	return fmt.Sprintf(`
resource "segment_tracking_plan" "test" {
  display_name = %q
  
  rules_identify = <<-EOF
%s
EOF
  rules_events = [
%s
  ]
}
`, rName, ruleStringFromFile(identifyFile), strings.Join(events, ","))
}

func testAccUpdateTrackingPlan(name string, tp *segmentapi.TrackingPlan, eventsFiles []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[name]

		events := make([]segmentapi.Event, 0, len(eventsFiles))
		for _, f := range eventsFiles {
			events = append(events, eventFromFile(f))
		}

		client := testAccProvider.Meta().(*segmentapi.Client)
		tp.Rules.Events = events
		_, err := client.UpdateTrackingPlan(rs.Primary.ID, *tp)
		if err != nil {
			return fmt.Errorf("error updating tracking plan %q: %w", tp.Name, err)
		}
		return nil
	}
}

//go:embed testdata/tracking_plans
var trackingPlans embed.FS

func ruleStringFromFile(filename string) string {
	file, err := trackingPlans.ReadFile("testdata/tracking_plans/" + filename)
	if err != nil {
		panic(err)
	}
	rule := segmentapi.Rules{}
	_ = json.Unmarshal(file, &rule)
	return toPrettyJsonString(rule)
}

func eventFromFile(filename string) segmentapi.Event {
	file, err := trackingPlans.ReadFile("testdata/tracking_plans/" + filename)
	if err != nil {
		panic(err)
	}
	event := segmentapi.Event{}
	_ = json.Unmarshal(file, &event)
	return event
}

func eventStringFromFile(filename string) string {
	return toPrettyJsonString(eventFromFile(filename))
}

func toPrettyJsonString(v interface{}) string {
	s, _ := json.MarshalIndent(v, "", "  ")
	return string(s)
}
