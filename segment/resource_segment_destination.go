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

func resourceSegmentDestination() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"slug": {
				Description: `Short name of the destination (e.g. "webhooks")`,
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
			"connection_mode": {
				Description: `Connection mode of the destination`,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"enabled": {
				Description: `Delivery enabled for the destination`,
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"configs": {
				Description: `Config of the destination`,
				Type:        schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Required: true,
			},
		},
		CreateContext: resourceSegmentDestinationCreate,
		ReadContext:   resourceSegmentDestinationRead,
		UpdateContext: resourceSegmentDestinationUpdate,
		DeleteContext: resourceSegmentDestinationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSegmentDestinationCreate(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)

	slug := r.Get("slug").(string)
	srcSlug := r.Get("source_slug").(string)
	connMode := r.Get("connection_mode").(string)
	enabled := r.Get("enabled").(bool)
	configs := r.Get("configs").(*schema.Set)

	dest, err := client.CreateDestination(srcSlug, slug, connMode, enabled, extractDestinationConfigs(configs))
	if err != nil {
		return diag.FromErr(err)
	}

	r.SetId(dest.Name)

	return resourceSegmentDestinationRead(c, r, meta)
}

func resourceSegmentDestinationRead(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)

	slug := DestinationNameToSlug(r.Id())
	srcSlug := DestinationNameToSourceSlug(r.Id())

	d, err := client.GetDestination(srcSlug, slug)
	if err != nil {
		if IsNotFoundErr(err) {
			r.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
	}

	if err := r.Set("slug", slug); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("source_slug", srcSlug); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("connection_mode", d.ConnectionMode); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("enabled", d.Enabled); err != nil {
		return diag.FromErr(err)
	}

	configs, err := flattenDestinationConfigs(d.Configs)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set("configs", configs); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSegmentDestinationUpdate(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)

	slug := r.Get("slug").(string)
	srcSlug := r.Get("source_slug").(string)
	enabled := r.Get("enabled").(bool)
	configs := r.Get("configs").(*schema.Set)

	_, err := client.UpdateDestination(srcSlug, slug, enabled, extractDestinationConfigs(configs))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSegmentDestinationRead(c, r, meta)
}

func resourceSegmentDestinationDelete(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)

	slug := DestinationNameToSlug(r.Id())
	srcSlug := DestinationNameToSourceSlug(r.Id())

	err := client.DeleteDestination(srcSlug, slug)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func extractDestinationConfigs(s *schema.Set) []segment.DestinationConfig {
	configs := make([]segment.DestinationConfig, 0)

	if s != nil {
		for _, config := range s.List() {
			c := segment.DestinationConfig{
				Name:  config.(map[string]interface{})["name"].(string),
				Type:  config.(map[string]interface{})["type"].(string),
				Value: extractDestinationConfigValue(config),
			}
			configs = append(configs, c)
		}
	}

	return configs
}

func extractDestinationConfigValue(config interface{}) interface{} {
	v := config.(map[string]interface{})["value"]

	if val, err := toJsonObject(v); err == nil {
		return val
	}
	if val, err := toJsonArray(v); err == nil {
		return val
	}
	return v
}

func flattenDestinationConfigs(dcs []segment.DestinationConfig) ([]interface{}, error) {
	if dcs != nil {
		cs := make([]interface{}, len(dcs), len(dcs))

		for i, dc := range dcs {
			c := make(map[string]interface{})

			if !IsNilOrZeroValue(dc.Value) {
				v, err := json.Marshal(dc.Value)
				if err != nil {
					return nil, fmt.Errorf("cannot flatten configs: %w", err)
				}
				c["value"] = string(v)
			}
			c["name"] = dc.Name
			c["type"] = dc.Type

			cs[i] = c
		}

		return cs, nil
	}

	return make([]interface{}, 0), nil
}

func DestinationNameToSlug(name string) string {
	return strings.Split(name, "/")[5]
}

func DestinationNameToSourceSlug(name string) string {
	return strings.Split(name, "/")[3]
}
