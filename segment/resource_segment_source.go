package segment

import (
	"context"
	"fmt"
	"github.com/forteilgmbh/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func resourceSegmentSource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"slug": {
				Description: `Short name of the source (e.g. "ios")`,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"catalog_name": {
				Description: "Catalog name of the source",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
		CreateContext: resourceSegmentSourceCreate,
		ReadContext:   resourceSegmentSourceRead,
		DeleteContext: resourceSegmentSourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSegmentSourceCreate(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)

	slug := r.Get("slug").(string)
	catName := r.Get("catalog_name").(string)

	source, err := client.CreateSource(slug, catName)
	if err != nil {
		return diag.FromErr(err)
	}

	r.SetId(source.Name)

	return resourceSegmentSourceRead(c, r, meta)
}

func resourceSegmentSourceRead(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)

	name := r.Id()
	slug := SourceNameToSlug(name)

	s, err := client.GetSource(slug)
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
	if err := r.Set("catalog_name", s.CatalogName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSegmentSourceDelete(c context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*segment.Client)
	name := r.Id()

	err := client.DeleteSource(SourceNameToSlug(name))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func IdToName(id string) string {
	splitID := strings.Split(id, "/")

	return splitID[len(splitID)-1]
}

func SourceSlugToName(workspace, slug string) string {
	return fmt.Sprintf("workspaces/%s/sources/%s", workspace, slug)
}

func SourceNameToSlug(name string) string {
	return strings.SplitN(name, "/", 4)[3]
}
