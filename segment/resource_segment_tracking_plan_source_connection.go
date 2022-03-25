package segment

import (
	"fmt"
	"strings"

	"github.com/forteilgmbh/segment-apis-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSegmentTrackingPlanSourceConnection() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"tracking_plan_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
		Create: resourceSegmentTrackingPlanSourceConnectionCreate,
		Read:   resourceSegmentTrackingPlanSourceConnectionRead,
		Delete: resourceSegmentTrackingPlanSourceConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSegmentTrackingPlanSourceConnectionImport,
		},
	}
}

func resourceSegmentTrackingPlanSourceConnectionCreate(r *schema.ResourceData, meta interface{}) error {
	client := meta.(*segment.Client)
	planId := r.Get("tracking_plan_id").(string)
	srcName := r.Get("source_name").(string)

	plan, err := client.CreateTrackingPlanSourceConnection(planId, srcName)
	if err != nil {
		return fmt.Errorf("error creating TrackingPlanSourceConnection: TrackingPlan: %q; Source: %q; err: %v", planId, srcName, err)
	}
	id := createTrackingPlanSourceConnectionId(plan.TrackingPlanID, plan.SourceName)
	r.SetId(id)

	return resourceSegmentTrackingPlanSourceConnectionRead(r, meta)
}

func resourceSegmentTrackingPlanSourceConnectionRead(r *schema.ResourceData, meta interface{}) error {
	client := meta.(*segment.Client)
	planId, srcName := splitTrackingPlanSourceConnectionId(r.Id())

	ok, err := findTrackingPlanSourceConnection(client, planId, srcName)
	if err != nil {
		return fmt.Errorf("error reading TrackingPlanSourceConnection: %w", err)
	}
	if !ok {
		r.SetId("")
		return nil
	}

	r.Set("tracking_plan_id", planId)
	r.Set("source_name", srcName)

	return nil
}

func resourceSegmentTrackingPlanSourceConnectionDelete(r *schema.ResourceData, meta interface{}) error {
	client := meta.(*segment.Client)
	id := r.Id()
	planId := r.Get("tracking_plan_id").(string)
	srcName := r.Get("source_name").(string)
	srcSlug := idToName(id)

	err := client.DeleteTrackingPlanSourceConnection(planId, srcSlug)
	if err != nil {
		return fmt.Errorf("error deleting TrackingPlanSourceConnection: TrackingPlan: %q; Source: %q; err: %v", planId, srcName, err)
	}

	return nil
}

func resourceSegmentTrackingPlanSourceConnectionImport(r *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*segment.Client)
	planId, srcName := splitTrackingPlanSourceConnectionId(r.Id())

	ok, err := findTrackingPlanSourceConnection(client, planId, srcName)
	if err != nil {
		return nil, fmt.Errorf("error importing TrackingPlanSourceConnection: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("error importing TrackingPlanSourceConnection %q: no source connection %q for tracking plan %q", r.Id(), srcName, planId)
	}

	r.SetId(r.Id())
	r.Set("tracking_plan_id", planId)
	r.Set("source_name", srcName)

	results := make([]*schema.ResourceData, 1)
	results[0] = r

	return results, nil
}

func createTrackingPlanSourceConnectionId(planId, srcName string) string {
	return fmt.Sprintf("%s|%s", planId, srcName)
}

func splitTrackingPlanSourceConnectionId(id string) (planId, srcName string) {
	s := strings.SplitN(id, "|", 2)
	return s[0], s[1]
}

func findTrackingPlanSourceConnection(client *segment.Client, planId, srcName string) (bool, error) {
	trackingPlanSourceConnections, err := client.ListTrackingPlanSourceConnections(planId)
	if err != nil {
		return false, fmt.Errorf("cannot fetch source connections for tracking plan %q: %w", planId, err)
	}
	for _, sc := range trackingPlanSourceConnections.Connections {
		if sc.TrackingPlanID == planId && sc.SourceName == srcName {
			return true, nil
		}
	}
	return false, nil
}
