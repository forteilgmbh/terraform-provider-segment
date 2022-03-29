package segment

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/forteilgmbh/segment-apis-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSegmentDestination() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"source_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"destination_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"connection_mode": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"configs": {
				Type: schema.TypeSet,
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
		Create: resourceSegmentDestinationCreate,
		Read:   resourceSegmentDestinationRead,
		Update: resourceSegmentDestinationUpdate,
		Delete: resourceSegmentDestinationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSegmentDestinationImport,
		},
	}
}

func resourceSegmentDestinationCreate(r *schema.ResourceData, meta interface{}) error {
	client := meta.(*segment.Client)
	srcName := r.Get("source_name").(string)
	destName := r.Get("destination_name").(string)
	connMode := r.Get("connection_mode").(string)
	enabled := r.Get("enabled").(bool)
	configs := r.Get("configs").(*schema.Set)

	dest, err := client.CreateDestination(srcName, destName, connMode, enabled, extractConfigs(configs))
	if err != nil {
		return fmt.Errorf("ERROR Creating Destination!! Source: %q; Destination: %q; err: %v", srcName, destName, err)
	}

	r.SetId(dest.Name)

	return resourceSegmentDestinationRead(r, meta)
}

func resourceSegmentDestinationRead(r *schema.ResourceData, meta interface{}) error {
	client := meta.(*segment.Client)
	srcName := r.Get("source_name").(string)
	id := r.Id()
	destName := IdToName(id)

	d, err := client.GetDestination(srcName, destName)
	if err != nil {
		return fmt.Errorf("ERROR Reading Destination!! Source: %q; Destination: %q; err: %v", srcName, destName, err)
	}

	r.Set("enabled", d.Enabled)
	r.Set("connection_mode", d.ConnectionMode)

	configs, err := flattenConfigs(d.Configs)
	if err != nil {
		return fmt.Errorf("cannot flatten configs for destination %q; err: %v", r.Id(), err)
	}
	err = r.Set("configs", configs)
	if err != nil {
		return fmt.Errorf("cannot set configs for destination %q; err: %v", r.Id(), err)
	}

	return nil
}

func resourceSegmentDestinationUpdate(r *schema.ResourceData, meta interface{}) error {
	client := meta.(*segment.Client)
	srcName := r.Get("source_name").(string)
	configs := r.Get("configs").(*schema.Set)
	enabled := r.Get("enabled").(bool)
	id := r.Id()
	destName := IdToName(id)

	_, err := client.UpdateDestination(srcName, destName, enabled, extractConfigs(configs))
	if err != nil {
		return fmt.Errorf("ERROR Updating Destination!! Source: %q; Destination: %q; err: %v", srcName, destName, err)
	}

	return resourceSegmentDestinationRead(r, meta)
}

func resourceSegmentDestinationDelete(r *schema.ResourceData, meta interface{}) error {
	client := meta.(*segment.Client)
	srcName := r.Get("source_name").(string)
	id := r.Id()
	destName := IdToName(id)

	err := client.DeleteDestination(srcName, destName)
	if err != nil {
		return fmt.Errorf("ERROR Deleting Destination!! Source: %q; Destination: %q; err: %v", srcName, destName, err)
	}

	return nil
}

func resourceSegmentDestinationImport(r *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*segment.Client)
	s := strings.SplitN(r.Id(), "/", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf(
			"invalid destination import format: %s (expected <SOURCE-NAME>/<DESTINATION-NAME>)",
			r.Id(),
		)
	}

	srcName := s[0]
	destName := s[1]

	d, err := client.GetDestination(srcName, destName)
	if err != nil {
		return nil, fmt.Errorf("invalid destination: %q; err: %v", r.Id(), err)
	}

	r.SetId(d.Name)
	r.Set("source_name", srcName)
	r.Set("destination_name", destName)
	r.Set("enabled", d.Enabled)
	r.Set("connection_mode", d.ConnectionMode)

	configs, err := flattenConfigs(d.Configs)
	if err != nil {
		return nil, fmt.Errorf("cannot flatten configs for destination: %q; err: %v", r.Id(), err)
	}
	err = r.Set("configs", configs)
	if err != nil {
		return nil, fmt.Errorf("cannot set configs for destination %q; err: %v", r.Id(), err)
	}

	results := make([]*schema.ResourceData, 1)
	results[0] = r

	return results, nil
}

func extractConfigs(s *schema.Set) []segment.DestinationConfig {
	configs := []segment.DestinationConfig{}

	if s != nil {
		for _, config := range s.List() {
			c := segment.DestinationConfig{
				Name:  config.(map[string]interface{})["name"].(string),
				Type:  config.(map[string]interface{})["type"].(string),
				Value: extractValue(config),
			}
			configs = append(configs, c)
		}
	}

	return configs
}

func extractValue(config interface{}) interface{} {
	v := config.(map[string]interface{})["value"]

	if val, err := toJsonObject(v); err == nil {
		return val
	}
	if val, err := toJsonArray(v); err == nil {
		return val
	}
	return v
}

func toJsonObject(data interface{}) (interface{}, error) {
	val := new(map[string]interface{})
	err := json.Unmarshal([]byte(data.(string)), val)
	return val, err
}

func toJsonArray(data interface{}) (interface{}, error) {
	val := new([]interface{})
	err := json.Unmarshal([]byte(data.(string)), val)
	return val, err
}

func flattenConfigs(dcs []segment.DestinationConfig) ([]interface{}, error) {
	if dcs != nil {
		cs := make([]interface{}, len(dcs), len(dcs))

		for i, dc := range dcs {
			c := make(map[string]interface{})

			if !IsNilOrZeroValue(dc.Value) {
				v, err := json.Marshal(dc.Value)
				if err != nil {
					return nil, err
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

func IsNilOrZeroValue(v interface{}) bool {
	return v == nil || reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}
