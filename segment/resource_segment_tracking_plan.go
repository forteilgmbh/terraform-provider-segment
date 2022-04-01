package segment

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/forteilgmbh/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func resourceSegmentTrackingPlan() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"display_name": {
				Description: `Display name of the tracking plan`,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
			},
			"name": {
				Description: `Full name of the tracking plan (e.g. "workspaces/myworkspace/tracking-plans/rs_123")`,
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"rules_global": {
				Description: `Rules applied to all messages as JSON-encoded string`,
				Type:        schema.TypeString,
				Optional:    true,
				StateFunc:   sanitizedSegmentRule,
			},
			"rules_identify": {
				Description: `Rules applied to Identify calls as JSON-encoded string`,
				Type:        schema.TypeString,
				Optional:    true,
				StateFunc:   sanitizedSegmentRule,
			},
			"rules_group": {
				Description: `Rules applied to Group calls as JSON-encoded string`,
				Type:        schema.TypeString,
				Optional:    true,
				StateFunc:   sanitizedSegmentRule,
			},
			"rules_events": {
				Description: `Rules applied to Track calls as list of JSON-encoded strings`,
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					StateFunc: sanitizedSegmentEvent,
					Type:      schema.TypeString,
				},
			},
		},
		CreateContext: resourceSegmentTrackingPlanCreate,
		ReadContext:   resourceSegmentTrackingPlanRead,
		DeleteContext: resourceSegmentTrackingPlanDelete,
		UpdateContext: resourceSegmentTrackingPlanUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSegmentTrackingPlanCreate(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)

	displayName := r.Get("display_name").(string)

	rules := &segment.RuleSet{}
	if tfRule, ok := r.GetOk("rules_global"); ok {
		rule, err := fromTfStateToRule(tfRule)
		if err != nil {
			return diag.Errorf("invalid \"global\" rules: %s", err)
		}
		rules.Global = &rule
	}
	if tfRule, ok := r.GetOk("rules_identify"); ok {
		rule, err := fromTfStateToRule(tfRule)
		if err != nil {
			return diag.Errorf("invalid \"identify\" rules: %s", err)
		}
		rules.Identify = &rule
	}
	if tfRule, ok := r.GetOk("rules_group"); ok {
		rule, err := fromTfStateToRule(tfRule)
		if err != nil {
			return diag.Errorf("invalid \"group\" rules: %s", err)
		}
		rules.Group = &rule
	}
	if tfEvents, ok := r.GetOk("rules_events"); ok {
		events, err := fromTfStateToEvents(tfEvents)
		if err != nil {
			return diag.Errorf("invalid \"events\" rules: %s", err)
		}
		rules.Events = events
	}

	trackingPlan, err := client.CreateTrackingPlan(segment.TrackingPlan{DisplayName: displayName, Rules: *rules})
	if err != nil {
		return diag.FromErr(err)
	}

	planId := TrackingPlanNameToId(trackingPlan.Name)
	r.SetId(planId)

	return resourceSegmentTrackingPlanRead(c, r, meta)
}

func resourceSegmentTrackingPlanRead(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)
	planId := r.Id()
	names, err := getTrackingPlansNames(client)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, ok := names[planId]; !ok {
		r.SetId("")
		return nil
	}
	trackingPlan, err := client.GetTrackingPlan(planId)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := r.Set("display_name", trackingPlan.DisplayName); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("name", trackingPlan.Name); err != nil {
		return diag.FromErr(err)
	}

	if _, ok := r.GetOk("rules_global"); ok {
		if err := r.Set("rules_global", toTfState(trackingPlan.Rules.Global)); err != nil {
			return diag.FromErr(err)
		}
	}
	if _, ok := r.GetOk("rules_identify"); ok {
		if err := r.Set("rules_identify", toTfState(trackingPlan.Rules.Identify)); err != nil {
			return diag.FromErr(err)
		}
	}
	if _, ok := r.GetOk("rules_group"); ok {
		if err := r.Set("rules_group", toTfState(trackingPlan.Rules.Group)); err != nil {
			return diag.FromErr(err)
		}
	}
	if _, ok := r.GetOk("rules_events"); ok {
		events := make([]interface{}, 0, len(trackingPlan.Rules.Events))
		for _, e := range trackingPlan.Rules.Events {
			events = append(events, toTfState(e))
		}
		if err := r.Set("rules_events", events); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceSegmentTrackingPlanDelete(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)
	planId := r.Id()
	err := client.DeleteTrackingPlan(planId)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceSegmentTrackingPlanUpdate(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)
	planId := r.Id()
	displayName := r.Get("display_name").(string)

	names, err := getTrackingPlansNames(client)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, ok := names[planId]; !ok {
		return diag.Errorf("plan no longer exists")
	}
	trackingPlan, err := client.GetTrackingPlan(planId)
	if err != nil {
		return diag.FromErr(err)
	}
	rules := trackingPlan.Rules

	if tfRule, ok := r.GetOk("rules_global"); ok {
		rule, err := fromTfStateToRule(tfRule)
		if err != nil {
			return diag.Errorf("invalid \"global\" rules: %s", err)
		}
		rules.Global = &rule
	}
	if tfRule, ok := r.GetOk("rules_identify"); ok {
		rule, err := fromTfStateToRule(tfRule)
		if err != nil {
			return diag.Errorf("invalid \"identify\" rules: %s", err)
		}
		rules.Identify = &rule
	}
	if tfRule, ok := r.GetOk("rules_group"); ok {
		rule, err := fromTfStateToRule(tfRule)
		if err != nil {
			return diag.Errorf("invalid \"group\" rules: %s", err)
		}
		rules.Group = &rule
	}
	if tfEvents, ok := r.GetOk("rules_events"); ok {
		events, err := fromTfStateToEvents(tfEvents)
		if err != nil {
			return diag.Errorf("invalid \"events\" rules: %s", err)
		}
		rules.Events = events
	}

	updatedPlan := segment.TrackingPlan{
		DisplayName: displayName,
		Rules:       rules,
	}
	_, err = client.UpdateTrackingPlan(planId, updatedPlan)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceSegmentTrackingPlanRead(c, r, meta)
}

func TrackingPlanNameToId(name string) string {
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
		id := TrackingPlanNameToId(element.Name)
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

func fromTfStateToRule(v interface{}) (segment.Rules, error) {
	rule := segment.Rules{}
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
	if IsNilOrZeroValue(v) {
		return ""
	}
	j, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(j)
}
