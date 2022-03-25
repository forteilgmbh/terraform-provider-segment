package segment

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/forteilgmbh/segment-apis-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSegmentTrackingPlan() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"display_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"rules_global": {
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: sanitizedSegmentRule,
			},
			"rules_identify": {
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: sanitizedSegmentRule,
			},
			"rules_group": {
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: sanitizedSegmentRule,
			},
			"rules_events": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					StateFunc: sanitizedSegmentEvent,
					Type:      schema.TypeString,
				},
			},
		},
		Create: resourceSegmentTrackingPlanCreate,
		Read:   resourceSegmentTrackingPlanRead,
		Delete: resourceSegmentTrackingPlanDelete,
		Update: resourceSegmentTrackingPlanUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceSegmentTrackingPlanImport,
		},
	}
}

func resourceSegmentTrackingPlanCreate(r *schema.ResourceData, meta interface{}) error {
	client := meta.(*segment.Client)

	displayName := r.Get("display_name").(string)
	rules := &segment.Rules{}

	if tfRule, ok := r.GetOk("rules_global"); ok {
		rule, err := fromTfStateToRule(tfRule)
		if err != nil {
			return fmt.Errorf("error creating tracking plan %q: invalid \"global\" rules: %w", displayName, err)
		}
		rules.Global = &rule
	}
	if tfRule, ok := r.GetOk("rules_identify"); ok {
		rule, err := fromTfStateToRule(tfRule)
		if err != nil {
			return fmt.Errorf("error creating tracking plan %q: invalid \"identify\" rules: %w", displayName, err)
		}
		rules.Identify = &rule
	}
	if tfRule, ok := r.GetOk("rules_group"); ok {
		rule, err := fromTfStateToRule(tfRule)
		if err != nil {
			return fmt.Errorf("error creating tracking plan %q: invalid \"group\" rules: %w", displayName, err)
		}
		rules.Group = &rule
	}
	if tfEvents, ok := r.GetOk("rules_events"); ok {
		events, err := fromTfStateToEvents(tfEvents)
		if err != nil {
			return fmt.Errorf("error creating tracking plan %q: invalid \"events\" rules: %w", displayName, err)
		}
		rules.Events = events
	}

	trackingPlan, err := client.CreateTrackingPlan(displayName, *rules)
	if err != nil {
		return fmt.Errorf("error creating tracking plan %q: %w", displayName, err)
	}

	planName := parseNameID(trackingPlan.Name)
	r.SetId(planName)
	return resourceSegmentTrackingPlanRead(r, meta)
}

func resourceSegmentTrackingPlanRead(r *schema.ResourceData, meta interface{}) error {
	client := meta.(*segment.Client)
	planName := r.Id()
	names, err := getTrackingPlansNames(client)
	if err != nil {
		return fmt.Errorf("error reading tracking plan %q: %w", planName, err)
	}
	if _, ok := names[planName]; !ok {
		r.SetId("")
		return nil
	}
	trackingPlan, err := client.GetTrackingPlan(planName)
	if err != nil {
		return fmt.Errorf("error reading tracking plan %q: %w", planName, err)
	}

	r.Set("display_name", trackingPlan.DisplayName)
	r.Set("name", planName)

	if _, ok := r.GetOk("rules_global"); ok {
		r.Set("rules_global", toTfState(trackingPlan.Rules.Global))
	}
	if _, ok := r.GetOk("rules_identify"); ok {
		r.Set("rules_identify", toTfState(trackingPlan.Rules.Identify))
	}
	if _, ok := r.GetOk("rules_group"); ok {
		r.Set("rules_group", toTfState(trackingPlan.Rules.Group))
	}
	if _, ok := r.GetOk("rules_events"); ok {
		events := make([]interface{}, 0, len(trackingPlan.Rules.Events))
		for _, e := range trackingPlan.Rules.Events {
			events = append(events, toTfState(e))
		}
		r.Set("rules_events", events)
	}

	return nil
}

func resourceSegmentTrackingPlanDelete(r *schema.ResourceData, meta interface{}) error {
	client := meta.(*segment.Client)
	planName := r.Id()
	err := client.DeleteTrackingPlan(planName)
	if err != nil {
		return fmt.Errorf("ERROR Deleting Tracking Plan!! PlanName: %q; err: %v", planName, err)
	}

	return nil
}

func resourceSegmentTrackingPlanUpdate(r *schema.ResourceData, meta interface{}) error {
	client := meta.(*segment.Client)
	planName := r.Id()
	displayName := r.Get("display_name").(string)

	names, err := getTrackingPlansNames(client)
	if err != nil {
		return fmt.Errorf("error updating tracking plan %q: %w", planName, err)
	}
	if _, ok := names[planName]; !ok {
		return fmt.Errorf("error updating tracking plan %q: plan no longer exists", planName)
	}
	trackingPlan, err := client.GetTrackingPlan(planName)
	if err != nil {
		return fmt.Errorf("error updating tracking plan %q: %w", planName, err)
	}
	rules := trackingPlan.Rules

	if tfRule, ok := r.GetOk("rules_global"); ok {
		rule, err := fromTfStateToRule(tfRule)
		if err != nil {
			return fmt.Errorf("error updating tracking plan %q: invalid \"global\" rules: %w", displayName, err)
		}
		rules.Global = &rule
	}
	if tfRule, ok := r.GetOk("rules_identify"); ok {
		rule, err := fromTfStateToRule(tfRule)
		if err != nil {
			return fmt.Errorf("error updating tracking plan %q: invalid \"identify\" rules: %w", displayName, err)
		}
		rules.Identify = &rule
	}
	if tfRule, ok := r.GetOk("rules_group"); ok {
		rule, err := fromTfStateToRule(tfRule)
		if err != nil {
			return fmt.Errorf("error updating tracking plan %q: invalid \"group\" rules: %w", displayName, err)
		}
		rules.Group = &rule
	}
	if tfEvents, ok := r.GetOk("rules_events"); ok {
		events, err := fromTfStateToEvents(tfEvents)
		if err != nil {
			return fmt.Errorf("error updating tracking plan %q: invalid \"events\" rules: %w", displayName, err)
		}
		rules.Events = events
	}

	paths := []string{"tracking_plan.display_name", "tracking_plan.rules"}
	updatedPlan := segment.TrackingPlan{
		DisplayName: displayName,
		Rules:       rules,
	}
	_, err = client.UpdateTrackingPlan(planName, paths, updatedPlan)
	if err != nil {
		return fmt.Errorf("error updating tracking plan %q: %w", planName, err)
	}
	return resourceSegmentTrackingPlanRead(r, meta)
}

func resourceSegmentTrackingPlanImport(r *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*segment.Client)
	planName := r.Id()
	names, err := getTrackingPlansNames(client)
	if err != nil {
		return nil, fmt.Errorf("error importing tracking plan %q: %w", planName, err)
	}
	if _, ok := names[planName]; !ok {
		r.SetId("")
		return nil, fmt.Errorf("error importing tracking plan %q: plan does not exist", planName)
	}
	trackingPlan, err := client.GetTrackingPlan(planName)
	if err != nil {
		return nil, fmt.Errorf("error importing tracking plan %q: %w", planName, err)
	}

	r.Set("display_name", trackingPlan.DisplayName)
	r.Set("name", planName)

	r.Set("rules_global", toTfState(trackingPlan.Rules.Global))
	r.Set("rules_identify", toTfState(trackingPlan.Rules.Identify))
	r.Set("rules_group", toTfState(trackingPlan.Rules.Group))

	events := make([]interface{}, 0, len(trackingPlan.Rules.Events))
	for _, e := range trackingPlan.Rules.Events {
		events = append(events, toTfState(e))
	}
	r.Set("rules_events", events)

	results := make([]*schema.ResourceData, 1)
	results[0] = r

	return results, nil
}

func parseNameID(name string) string {
	nameSplit := strings.Split(name, "/")
	return nameSplit[len(nameSplit)-1]
}

func getTrackingPlansNames(client *segment.Client) (map[string]string, error) {
	plans, err := client.ListTrackingPlans()
	if err != nil {
		return nil, fmt.Errorf("cannot list tracking plans: %w", err)
	}
	names := make(map[string]string)
	for _, element := range plans.TrackingPlans {
		id := parseNameID(element.Name)
		names[id] = element.Name
	}
	return names, nil
}

func sanitizedSegmentRule(val interface{}) string {
	rule, err := fromTfStateToRule(val)
	if err != nil {
		panic(err)
	}
	return toTfState(rule)
}

func sanitizedSegmentEvent(val interface{}) string {
	event, err := fromTfStateToEvent(val)
	if err != nil {
		panic(err)
	}
	return toTfState(event)
}

func fromTfStateToRule(v interface{}) (segment.Rule, error) {
	rule := segment.Rule{}
	err := json.Unmarshal([]byte(v.(string)), &rule)
	return rule, err
}

func fromTfStateToEvents(v interface{}) ([]segment.Event, error) {
	events := make([]segment.Event, 0)
	for i, tfEvent := range v.([]interface{}) {
		event, err := fromTfStateToEvent(tfEvent)
		if err != nil {
			return nil, fmt.Errorf("at #%d: %w", i, err)
		}
		events = append(events, event)
	}
	return events, nil
}

func fromTfStateToEvent(v interface{}) (segment.Event, error) {
	event := segment.Event{}
	err := json.Unmarshal([]byte(v.(string)), &event)
	return event, err
}

func toTfState(v interface{}) string {
	if isNilOrZeroValue(v) {
		return ""
	}
	j, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(j)
}
