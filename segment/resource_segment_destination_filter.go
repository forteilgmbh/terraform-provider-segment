package segment

import (
	"context"
	"fmt"
	"github.com/forteilgmbh/segment-config-go/segment"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
)

func resourceSegmentDestinationFilter() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"source_slug": {
				Description: `Short name of the source (e.g. "ios")`,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"destination_slug": {
				Description: `Short name of the destination (e.g. "webhooks")`,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: `Full name of the destination filter (e.g. "workspaces/myworkspace/sources/mysource/destinations/mydestination/config/abc123/filters/df_123")`,
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"title": {
				Description: `A human-readable title for this filter`,
				Type:        schema.TypeString,
				Optional:    true,
			},
			"description": {
				Description: `A longer human-readable description of this filter`,
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enabled": {
				Description: `Whether or not this filter should be active`,
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"conditions": {
				Description: `A FQL statement that causes this filter’s action to be applied if it evaluates to true. "all" will cause the filter to be applied to all events.`,
				Type:        schema.TypeString,
				Required:    true,
			},
			"action": {
				Description: `The filtering action to take on events that match the "if" statement:
- "drop_event" will cause the event to be dropped and not sent to the destination if the "if" statement evaluates to true.
- "sample_event" will allow only a percentage of events through. It can sample randomly or, if given a path attribute, it can sample a percentage of events based on the contents of a field. This is useful for sampling all events for a percentage of users rather than a percentage of all events for all users.
- "whitelist_fields" takes a list of objects and a list of fields for each object that should be allowed, with all other fields in those objects dropped.
- "blacklist_fields" takes a list of nested objects and a list of fields for each object that should be dropped, with all other fields in those objects untouched.
`,
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 4,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description:  `The action name: one of "drop_event", "sample_event", "whitelist_fields" or "blacklist_fields"`,
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"drop_event", "sample_event", "whitelist_fields", "blacklist_fields"}, false),
						},
						"percent": {
							Description:  `Required for "sample_event". A percentage in the range [0.0, 1.0] that determines the percent of events to allow through. 0.0 will allow no events and 1.0 will allow all events.`,
							Type:         schema.TypeFloat,
							Optional:     true,
							ValidateFunc: validation.FloatBetween(0.0, 1.0),
						},
						"path": {
							Description: `Optional for "sample_event". If non-empty, events will be sampled based on the value at this path. For example, if path is userId, a percentage of users will have their events allowed through to the destination`,
							Type:        schema.TypeString,
							Optional:    true,
						},
						"fields": {
							Description: `Required for "whitelist_fields" or "blacklist_fields". Specifies which fields within the object to allow/drop.`,
							Type:        schema.TypeSet,
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"context": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Description: `One or more JSON object field names. Nested fields (i.e. dot-separated field names) are not supported.`,
											Type:        schema.TypeString,
										},
									},
									"traits": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Description: `One or more JSON object field names. Nested fields (i.e. dot-separated field names) are not supported.`,
											Type:        schema.TypeString,
										},
									},
									"properties": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Description: `One or more JSON object field names. Nested fields (i.e. dot-separated field names) are not supported.`,
											Type:        schema.TypeString,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		CreateContext: resourceSegmentDestinationFilterCreate,
		ReadContext:   resourceSegmentDestinationFilterRead,
		UpdateContext: resourceSegmentDestinationFilterUpdate,
		DeleteContext: resourceSegmentDestinationFilterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSegmentDestinationFilterImport,
		},
		CustomizeDiff: customdiff.Sequence(
			customizeDiffValidateDestinationFilterActions,
		),
	}
}

func resourceSegmentDestinationFilterCreate(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)

	srcSlug := r.Get("source_slug").(string)
	dstSlug := r.Get("destination_slug").(string)
	title := r.Get("title").(string)
	description := r.Get("description").(string)
	enabled := r.Get("enabled").(bool)
	conditions := r.Get("conditions").(string)
	actions := r.Get("action").(*schema.Set)

	filter := segment.DestinationFilter{
		Title:       title,
		Description: description,
		IsEnabled:   enabled,
		Conditions:  conditions,
		Actions:     extractDestinationFiltersActions(actions),
	}

	df, err := client.CreateDestinationFilter(srcSlug, dstSlug, filter)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := r.Set("name", df.Name); err != nil {
		return diag.FromErr(err)
	}
	r.SetId(DestinationFilterNameToId(df.Name))

	return resourceSegmentDestinationFilterRead(c, r, meta)
}

func resourceSegmentDestinationFilterRead(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)

	id := r.Id()
	srcSlug := r.Get("source_slug").(string)
	dstSlug := r.Get("destination_slug").(string)

	df, err := client.GetDestinationFilter(srcSlug, dstSlug, id)
	if err != nil {
		if IsNotFoundErr(err) || Is500ValidatePermissionsErr(err) {
			r.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
	}
	if err := r.Set("name", df.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("title", df.Title); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("description", df.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("enabled", df.IsEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("conditions", df.Conditions); err != nil {
		return diag.FromErr(err)
	}
	actions, err := flattenDestinationFilterActions(df.Actions)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("action", actions); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSegmentDestinationFilterUpdate(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)

	name := r.Get("name").(string)
	srcSlug := r.Get("source_slug").(string)
	dstSlug := r.Get("destination_slug").(string)
	title := r.Get("title").(string)
	description := r.Get("description").(string)
	enabled := r.Get("enabled").(bool)
	conditions := r.Get("conditions").(string)
	actions := r.Get("action").(*schema.Set)

	filter := segment.DestinationFilter{
		Name:        name,
		Title:       title,
		Description: description,
		IsEnabled:   enabled,
		Conditions:  conditions,
		Actions:     extractDestinationFiltersActions(actions),
	}

	_, err := client.UpdateDestinationFilter(srcSlug, dstSlug, filter)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSegmentDestinationFilterRead(c, r, meta)
}

func resourceSegmentDestinationFilterDelete(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)

	id := r.Id()
	srcSlug := r.Get("source_slug").(string)
	dstSlug := r.Get("destination_slug").(string)

	if err := client.DeleteDestinationFilter(srcSlug, dstSlug, id); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceSegmentDestinationFilterImport(c context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	srcSlug := SourceNameToSlug(d.Id())
	dstSlug := DestinationNameToSlug(d.Id())
	id := DestinationFilterNameToId(d.Id())

	if err := d.Set("source_slug", srcSlug); err != nil {
		return nil, err
	}
	if err := d.Set("destination_slug", dstSlug); err != nil {
		return nil, err
	}
	d.SetId(id)

	return []*schema.ResourceData{d}, nil
}

func extractDestinationFiltersActions(s *schema.Set) segment.DestinationFilterActions {
	actions := make(segment.DestinationFilterActions, 0)

	for _, action := range s.List() {
		a := action.(map[string]interface{})
		typ := a["type"].(string)
		switch typ {
		case "drop_event":
			actions = append(actions, segment.NewDropEventAction())
		case "sample_event":
			percent := float32(a["percent"].(float64))
			path := a["path"].(string)
			actions = append(actions, segment.NewSamplingEventAction(percent, path))
		case "whitelist_fields":
			f := a["fields"].(*schema.Set).List()[0].(map[string]interface{})
			actions = append(actions, segment.NewAllowListEventAction(
				extractDestinationFiltersActionsFields(f["properties"]),
				extractDestinationFiltersActionsFields(f["context"]),
				extractDestinationFiltersActionsFields(f["traits"]),
			))
		case "blacklist_fields":
			f := a["fields"].(*schema.Set).List()[0].(map[string]interface{})
			actions = append(actions, segment.NewBlockListEventAction(
				extractDestinationFiltersActionsFields(f["properties"]),
				extractDestinationFiltersActionsFields(f["context"]),
				extractDestinationFiltersActionsFields(f["traits"]),
			))
		}
	}

	return actions
}

func extractDestinationFiltersActionsFields(fields interface{}) []string {
	fs := make([]string, 0)
	for _, f := range fields.([]interface{}) {
		fs = append(fs, f.(string))
	}
	return fs
}

func flattenDestinationFilterActions(actions segment.DestinationFilterActions) ([]interface{}, error) {
	if actions == nil || len(actions) == 0 {
		return make([]interface{}, 0), nil
	}

	tfActions := make([]interface{}, len(actions), len(actions))

	for i, action := range actions {
		tfAction := make(map[string]interface{})
		typ := action.ActionType()

		switch typ {
		case segment.DestinationFilterActionTypeDropEvent:
			tfAction["type"] = "drop_event"
		case segment.DestinationFilterActionTypeSampling:
			tfAction["type"] = "sample_event"
			a := action.(segment.SamplingEventAction)
			tfAction["percent"] = a.Percent
			if !IsNilOrZeroValue(a.Path) {
				tfAction["path"] = a.Path
			}
		case segment.DestinationFilterActionTypeAllowList:
			tfAction["type"] = "whitelist_fields"
			tfAction["fields"] = flattenDestinationFilterActionsFields(action.(segment.FieldsListEventAction).Fields)
		case segment.DestinationFilterActionTypeBlockList:
			tfAction["type"] = "blacklist_fields"
			tfAction["fields"] = flattenDestinationFilterActionsFields(action.(segment.FieldsListEventAction).Fields)
		}

		tfActions[i] = tfAction
	}

	return tfActions, nil
}

func flattenDestinationFilterActionsFields(actionFields segment.EventDescription) []interface{} {
	fields := make(map[string]interface{})
	if !IsNilOrZeroValue(actionFields.Properties) {
		fields["properties"] = strListToInterfaceList(actionFields.Properties.Fields)
	}
	if !IsNilOrZeroValue(actionFields.Context) {
		fields["context"] = strListToInterfaceList(actionFields.Context.Fields)
	}
	if !IsNilOrZeroValue(actionFields.Traits) {
		fields["traits"] = strListToInterfaceList(actionFields.Traits.Fields)
	}
	return []interface{}{fields}
}

func DestinationFilterNameToId(name string) string {
	return strings.Split(name, "/")[9]
}

func customizeDiffValidateDestinationFilterActions(c context.Context, diff *schema.ResourceDiff, v interface{}) error {
	fieldsKeys := []string{"context", "properties", "traits"}

	var err *multierror.Error
	actions := diff.Get("action").(*schema.Set)

	for i, as := range actions.List() {
		a := as.(map[string]interface{})
		f := a["fields"].(*schema.Set).List()
		typ := a["type"].(string)
		switch typ {
		case "drop_event":
			if ok, extraKeys := allForbiddenKeysNotInMap(a, "percent", "path"); !ok {
				err = multierror.Append(err, fmt.Errorf("at %d: for %q action, attributes %v should not be set", i, typ, extraKeys))
			}
			if len(f) > 0 {
				err = multierror.Append(err, fmt.Errorf("at %d: for %q action, block \"fields\" should not be set", i, typ))
			}
		case "sample_event":
			if ok, missingKeys := allRequiredKeysInMap(a, "percent"); !ok {
				err = multierror.Append(err, fmt.Errorf("at %d: for %q action, attributes %v must be set", i, typ, missingKeys))
			}
			if len(f) > 0 {
				err = multierror.Append(err, fmt.Errorf("at %d: for %q action, block \"fields\" should not be set", i, typ))
			}
		case "whitelist_fields":
			fallthrough
		case "blacklist_fields":
			if ok, extraKeys := allForbiddenKeysNotInMap(a, "percent", "path"); !ok {
				err = multierror.Append(err, fmt.Errorf("at %d: for %q action, attributes %v should not be set", i, typ, extraKeys))
			}
			if len(f) != 1 {
				err = multierror.Append(err, fmt.Errorf("at %d: for %q action, exactly one block \"fields\" must be set (is: %d)", i, typ, len(f)))
			} else if ok := anyRequiredKeyInMap(f[0].(map[string]interface{}), fieldsKeys...); !ok {
				err = multierror.Append(err, fmt.Errorf("at %d: for %q action in \"fields\" block at least one of %v must be set", i, typ, fieldsKeys))
			}
		default:
			err = multierror.Append(err, fmt.Errorf("at %d: invalid action type: %q", i, typ))
		}
	}

	return err.ErrorOrNil()
}
