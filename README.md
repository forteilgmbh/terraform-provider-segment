# terraform-provider-segment

A [Terraform](https://www.terraform.io/) provider for [Segment](https://www.segment.com)

## Usage

### Sources

Create and manage Segment [sources](https://segment.com/docs/sources/)
```
resource "segment_source" "test" {
  slug         = "your-source"
  catalog_name = "catalog/sources/javascript"
}
```
#### Attributes

- `id`: full Source name, e.g. `workspaces/your-workspace/sources/your-source`

#### Import

Through ID.

### Source schema configs

```
resource "segment_source_schema_config" "test" {
  source_slug = segment_source.test.slug
  
  allow_unplanned_track_events           = true
  allow_unplanned_identify_traits        = true
  allow_unplanned_group_traits           = true
  forwarding_blocked_events_to           = ""
  allow_unplanned_track_event_properties = true
  allow_track_event_on_violations        = true
  allow_identify_traits_on_violations    = true
  allow_group_traits_on_violations       = true
  forwarding_violations_to               = ""
  allow_track_properties_on_violations   = true
  common_track_event_on_violations       = "ALLOW"
  common_identify_event_on_violations    = "ALLOW"
  common_group_event_on_violations       = "ALLOW"
  
  depends_on = [segment_tracking_plan_source_connection.test]
}
```
#### Attributes

- `id`: full Source Schema Config name, e.g. `workspaces/your-workspace/sources/your-source/schema-config`

#### Import

Through ID.

### Destinations

Create and manage Segment [destinations](https://segment.com/docs/destinations/)
```
resource "segment_destination" "test" {
  slug            = "google-analytics"
  source_slug     = segment_source.test.slug
  connection_mode = "CLOUD"
  enabled         = false
  
  configs {
      name = "${segment_source.test.id}/destinations/google-analytics/config/trackingId"
      type = "string"
      value = "your-tracking-id"
  }
}
```

#### Attributes

- `id`: full Destination name, e.g. `workspaces/your-workspace/sources/your-source/destinations/google-analytics`

#### Import

Through ID.

### Destination Filters

```
resource "segment_destination_filter" "test" {
  title            = "my-destination-filter"
  
  source_slug      = segment_source.test.slug
  destination_slug = segment_destination.test.slug
  
  enabled    = false
  conditions = "type = \"identify\""
  
  action {
    type = "drop_event"
  }
  
  action {
    type    = "sample_event"
    percent = 0.5 # optional
  }
  
  action {
    type = "whitelist_fields" # or "blacklist_fields"
    
    fields {
      # at least one required
      context    = ["baa"]
      properties = ["foo", "bar"]
      traits     = ["baz"]
    }
  }
}
```

#### Attributes

- `id`: Destination Filter ID, e.g. `df_xyz987`
- `name`: full Destination Filter name, e.g. `workspaces/your-workspace/sources/your-source/destinations/google-analytics/config/abc123/filters/df_xyz987`

#### Import

Through `name`.

### Tracking Plans

```
resource "segment_tracking_plan" "test" {
  display_name = "my-tracking-plan"
  
  rules_global   = "json-schema-here"
  rules_group    = "json-schema-here"
  rules_identify = "json-schema-here"
  rules_events = [
    "json-schema-here",
  ]
}
```
All rules (global, group, identify, events) are managed independently from each other. 
This means that a tracking plan can be managed partially with Terraform and partially manually.

#### Attributes

- `id`: Tracking Plan ID, e.g. `rs_xyz987`
- `name`: full Tracking Plan name, e.g. `workspaces/your-workspace/tracking-plans/rs_xyz987`

#### Import

Through `name`.

### Tracking Plans Source Connections

```
resource "segment_tracking_plan_source_connection" "test" {
  tracking_plan_id = segment_tracking_plan.test.id
  source_slug      = segment_source.test.slug
}
```

#### Attributes

- `id`: artificial ID in `<tracking-plan-id>|<source-name>` format, e.g. `rs_xyz987|workspaces/your-workspace/sources/your-source`

#### Import

Through `id`.

