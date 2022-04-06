package segment

import (
	"context"
	"fmt"
	"github.com/forteilgmbh/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func resourceSegmentSourceSchemaConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"source_slug": {
				Description: `Short name of the source (e.g. "ios")`,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"allow_unplanned_track_events": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  DefaultSourceSchemaConfig["allow_unplanned_track_events"],
			},
			"allow_unplanned_identify_traits": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  DefaultSourceSchemaConfig["allow_unplanned_identify_traits"],
			},
			"allow_unplanned_group_traits": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  DefaultSourceSchemaConfig["allow_unplanned_group_traits"],
			},
			"forwarding_blocked_events_to": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  DefaultSourceSchemaConfig["forwarding_blocked_events_to"],
			},
			"allow_unplanned_track_event_properties": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  DefaultSourceSchemaConfig["allow_unplanned_track_event_properties"],
			},
			"allow_track_event_on_violations": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  DefaultSourceSchemaConfig["allow_track_event_on_violations"],
			},
			"allow_identify_traits_on_violations": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  DefaultSourceSchemaConfig["allow_identify_traits_on_violations"],
			},
			"allow_group_traits_on_violations": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  DefaultSourceSchemaConfig["allow_group_traits_on_violations"],
			},
			"forwarding_violations_to": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  DefaultSourceSchemaConfig["forwarding_violations_to"],
			},
			"allow_track_properties_on_violations": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  DefaultSourceSchemaConfig["allow_track_properties_on_violations"],
			},
			"common_track_event_on_violations": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  DefaultSourceSchemaConfig["common_track_event_on_violations"],
			},
			"common_identify_event_on_violations": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  DefaultSourceSchemaConfig["common_identify_event_on_violations"],
			},
			"common_group_event_on_violations": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  DefaultSourceSchemaConfig["common_group_event_on_violations"],
			},
		},
		CreateContext: resourceSegmentSourceSchemaConfigCreate,
		ReadContext:   resourceSegmentSourceSchemaConfigRead,
		UpdateContext: resourceSegmentSourceSchemaConfigCreate,
		DeleteContext: resourceSegmentSourceSchemaConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSegmentSourceSchemaConfigCreate(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)

	srcSlug := r.Get("source_slug").(string)

	configBefore, err := client.GetSourceConfig(srcSlug)
	if err != nil {
		return diag.FromErr(err)
	}

	config := segment.SourceConfig{
		AllowUnplannedTrackEvents:           r.Get("allow_unplanned_track_events").(bool),
		AllowUnplannedIdentifyTraits:        r.Get("allow_unplanned_identify_traits").(bool),
		AllowUnplannedGroupTraits:           r.Get("allow_unplanned_group_traits").(bool),
		ForwardingBlockedEventsTo:           r.Get("forwarding_blocked_events_to").(string),
		AllowUnplannedTrackEventsProperties: r.Get("allow_unplanned_track_event_properties").(bool),
		AllowTrackEventOnViolations:         r.Get("allow_track_event_on_violations").(bool),
		AllowIdentifyTraitsOnViolations:     r.Get("allow_identify_traits_on_violations").(bool),
		AllowGroupTraitsOnViolations:        r.Get("allow_group_traits_on_violations").(bool),
		ForwardingViolationsTo:              r.Get("forwarding_violations_to").(string),
		AllowTrackPropertiesOnViolations:    r.Get("allow_track_properties_on_violations").(bool),
		CommonTrackEventOnViolations:        segment.CommonEventSettings(r.Get("common_track_event_on_violations").(string)),
		CommonIdentifyEventOnViolations:     segment.CommonEventSettings(r.Get("common_identify_event_on_violations").(string)),
		CommonGroupEventOnViolations:        segment.CommonEventSettings(r.Get("common_group_event_on_violations").(string)),
	}

	sourceConfig, err := client.UpdateSourceConfig(srcSlug, config)
	if err != nil {
		return diag.FromErr(err)
	}

	r.SetId(sourceConfig.Name)

	err = waitUntilSourceSchemaConfigModified(client, srcSlug, configBefore)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSegmentSourceSchemaConfigRead(c, r, meta)
}

func resourceSegmentSourceSchemaConfigRead(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)

	name := r.Id()
	srcSlug := SourceNameToSlug(name)

	s, err := client.GetSourceConfig(srcSlug)
	if err != nil {
		if IsNotFoundErr(err) || Is500NilDereferenceErr(err) {
			r.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
	}

	if err := r.Set("source_slug", srcSlug); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("allow_unplanned_track_events", s.AllowUnplannedTrackEvents); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("allow_unplanned_identify_traits", s.AllowUnplannedIdentifyTraits); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("allow_unplanned_group_traits", s.AllowUnplannedGroupTraits); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("forwarding_blocked_events_to", s.ForwardingBlockedEventsTo); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("allow_unplanned_track_event_properties", s.AllowUnplannedTrackEventsProperties); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("allow_track_event_on_violations", s.AllowTrackEventOnViolations); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("allow_identify_traits_on_violations", s.AllowIdentifyTraitsOnViolations); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("allow_group_traits_on_violations", s.AllowGroupTraitsOnViolations); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("forwarding_violations_to", s.ForwardingViolationsTo); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("allow_track_properties_on_violations", s.AllowTrackPropertiesOnViolations); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("common_track_event_on_violations", string(s.CommonTrackEventOnViolations)); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("common_identify_event_on_violations", string(s.CommonIdentifyEventOnViolations)); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("common_group_event_on_violations", string(s.CommonGroupEventOnViolations)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSegmentSourceSchemaConfigDelete(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)

	name := r.Id()
	srcSlug := SourceNameToSlug(name)
	config := DefaultSegmentSourceSchemaConfig

	_, err := client.UpdateSourceConfig(srcSlug, config)
	if err != nil && !(IsNotFoundErr(err) || Is500NilDereferenceErr(err)) {
		return diag.FromErr(err)
	}
	return nil
}

var DefaultSourceSchemaConfig = map[string]interface{}{
	"allow_unplanned_track_events":           true,
	"allow_unplanned_identify_traits":        true,
	"allow_unplanned_group_traits":           true,
	"forwarding_blocked_events_to":           "",
	"allow_unplanned_track_event_properties": true,
	"allow_track_event_on_violations":        true,
	"allow_identify_traits_on_violations":    true,
	"allow_group_traits_on_violations":       true,
	"forwarding_violations_to":               "",
	"allow_track_properties_on_violations":   true,
	"common_track_event_on_violations":       "ALLOW",
	"common_identify_event_on_violations":    "ALLOW",
	"common_group_event_on_violations":       "ALLOW",
}

var DefaultSegmentSourceSchemaConfig = segment.SourceConfig{
	AllowUnplannedTrackEvents:           DefaultSourceSchemaConfig["allow_unplanned_track_events"].(bool),
	AllowUnplannedIdentifyTraits:        DefaultSourceSchemaConfig["allow_unplanned_identify_traits"].(bool),
	AllowUnplannedGroupTraits:           DefaultSourceSchemaConfig["allow_unplanned_group_traits"].(bool),
	ForwardingBlockedEventsTo:           DefaultSourceSchemaConfig["forwarding_blocked_events_to"].(string),
	AllowUnplannedTrackEventsProperties: DefaultSourceSchemaConfig["allow_unplanned_track_event_properties"].(bool),
	AllowTrackEventOnViolations:         DefaultSourceSchemaConfig["allow_track_event_on_violations"].(bool),
	AllowIdentifyTraitsOnViolations:     DefaultSourceSchemaConfig["allow_identify_traits_on_violations"].(bool),
	AllowGroupTraitsOnViolations:        DefaultSourceSchemaConfig["allow_group_traits_on_violations"].(bool),
	ForwardingViolationsTo:              DefaultSourceSchemaConfig["forwarding_violations_to"].(string),
	AllowTrackPropertiesOnViolations:    DefaultSourceSchemaConfig["allow_track_properties_on_violations"].(bool),
	CommonTrackEventOnViolations:        segment.CommonEventSettings(DefaultSourceSchemaConfig["common_track_event_on_violations"].(string)),
	CommonIdentifyEventOnViolations:     segment.CommonEventSettings(DefaultSourceSchemaConfig["common_identify_event_on_violations"].(string)),
	CommonGroupEventOnViolations:        segment.CommonEventSettings(DefaultSourceSchemaConfig["common_group_event_on_violations"].(string)),
}

func waitUntilSourceSchemaConfigModified(client *segment.Client, srcSlug string, configBefore segment.SourceConfig) error {
	c := make(chan error, 1)
	go func() {
		for {
			s, err := client.GetSourceConfig(srcSlug)
			if err != nil {
				c <- err
				break
			} else if s != configBefore {
				c <- nil
				break
			}
			time.Sleep(1 * time.Second)
		}
	}()
	select {
	case err := <-c:
		return err
	case <-time.After(1 * time.Minute):
		return fmt.Errorf("timeout after waiting for source schema %q to be modified", srcSlug)
	}
}
