package segment

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"

	"github.com/forteilgmbh/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSegmentTrackingPlanSourceConnection() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"tracking_plan_id": {
				Description: `Unique ID of the tracking plan (e.g. "rs_123")`,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"source_slug": {
				Description: `Short name of the source (e.g. "ios")`,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
		CreateContext: resourceSegmentTrackingPlanSourceConnectionCreate,
		ReadContext:   resourceSegmentTrackingPlanSourceConnectionRead,
		DeleteContext: resourceSegmentTrackingPlanSourceConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSegmentTrackingPlanSourceConnectionCreate(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)
	planId := r.Get("tracking_plan_id").(string)
	srcSlug := r.Get("source_slug").(string)

	err := client.CreateTrackingPlanSourceConnection(planId, srcSlug)
	if err != nil {
		return diag.FromErr(err)
	}
	id := createTrackingPlanSourceConnectionId(planId, srcSlug)
	r.SetId(id)

	return resourceSegmentTrackingPlanSourceConnectionRead(c, r, meta)
}

func resourceSegmentTrackingPlanSourceConnectionRead(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)
	planId, srcSlug := SplitTrackingPlanSourceConnectionId(r.Id())

	ok, err := FindTrackingPlanSourceConnection(client, planId, srcSlug)
	if err != nil {
		if IsNotFoundErr(err) {
			r.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
	}
	if !ok {
		r.SetId("")
		return nil
	}

	if err := r.Set("tracking_plan_id", planId); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("source_slug", srcSlug); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSegmentTrackingPlanSourceConnectionDelete(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)
	planId, srcSlug := SplitTrackingPlanSourceConnectionId(r.Id())

	err := client.DeleteTrackingPlanSourceConnection(planId, srcSlug)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func createTrackingPlanSourceConnectionId(planId, srcSlug string) string {
	return fmt.Sprintf("%s|%s", planId, srcSlug)
}

func SplitTrackingPlanSourceConnectionId(id string) (planId, srcName string) {
	s := strings.SplitN(id, "|", 2)
	return s[0], s[1]
}

func FindTrackingPlanSourceConnection(client *segment.Client, planId, srcSlug string) (bool, error) {
	trackingPlanSourceConnections, err := client.ListTrackingPlanSources(planId)
	if err != nil {
		return false, fmt.Errorf("cannot fetch source connections for tracking plan %q: %w", planId, err)
	}
	for _, sc := range trackingPlanSourceConnections {
		if sc.TrackingPlanId == planId && strings.HasSuffix(sc.Source, srcSlug) {
			return true, nil
		}
	}
	return false, nil
}
