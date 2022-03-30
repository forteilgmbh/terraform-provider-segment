package segment

import (
	"fmt"
	"strings"

	"github.com/forteilgmbh/segment-config-go/segment"
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
	srcSlug := IdToName(srcName)

	err := client.CreateTrackingPlanSourceConnection(planId, srcSlug)
	if err != nil {
		return fmt.Errorf("error creating TrackingPlanSourceConnection: TrackingPlan: %q; Source: %q; err: %v", planId, srcName, err)
	}
	id := createTrackingPlanSourceConnectionId(planId, srcName)
	r.SetId(id)

	return resourceSegmentTrackingPlanSourceConnectionRead(r, meta)
}

func resourceSegmentTrackingPlanSourceConnectionRead(r *schema.ResourceData, meta interface{}) error {
	client := meta.(*segment.Client)
	planId, srcName := SplitTrackingPlanSourceConnectionId(r.Id())

	ok, err := FindTrackingPlanSourceConnection(client, planId, srcName)
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
	srcSlug := IdToName(id)

	err := client.DeleteTrackingPlanSourceConnection(planId, srcSlug)
	if err != nil {
		return fmt.Errorf("error deleting TrackingPlanSourceConnection: TrackingPlan: %q; Source: %q; err: %v", planId, srcName, err)
	}

	return nil
}

func resourceSegmentTrackingPlanSourceConnectionImport(r *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*segment.Client)
	planId, srcName := SplitTrackingPlanSourceConnectionId(r.Id())

	ok, err := FindTrackingPlanSourceConnection(client, planId, srcName)
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

func SplitTrackingPlanSourceConnectionId(id string) (planId, srcName string) {
	s := strings.SplitN(id, "|", 2)
	return s[0], s[1]
}

func FindTrackingPlanSourceConnection(client *segment.Client, planId, srcName string) (bool, error) {
	trackingPlanSourceConnections, err := client.ListTrackingPlanSources(planId)
	if err != nil {
		return false, fmt.Errorf("cannot fetch source connections for tracking plan %q: %w", planId, err)
	}
	for _, sc := range trackingPlanSourceConnections {
		if sc.TrackingPlanId == planId && sc.Source == srcName {
			return true, nil
		}
	}
	return false, nil
}
