package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRotations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRotationsRead,

		Schema: map[string]*schema.Schema{
			"team_slug": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The team slug to get rotations for",
			},
			"rotations": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of rotations for the team",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The rotation name",
						},
						"slug": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The rotation slug",
						},
						"shift_length": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Length of each shift",
						},
						"shift_length_unit": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Unit of shift length (e.g., days, hours)",
						},
						"time_zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Time zone for the rotation",
						},
					},
				},
			},
		},
	}
}

func dataSourceRotationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	teamSlug := d.Get("team_slug").(string)

	rotationsResp, details, err := config.VictorOpsClient.ListRotationsV2(ctx, teamSlug)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to get rotations (%d): %s", details.StatusCode, details.ResponseBody)
	}

	rotations := make([]map[string]interface{}, len(rotationsResp.Rotations))
	for i, rotation := range rotationsResp.Rotations {
		rotations[i] = map[string]interface{}{
			"name":              rotation.Name,
			"slug":              rotation.Slug,
			"shift_length":      rotation.ShiftLength,
			"shift_length_unit": rotation.ShiftLengthUnit,
			"time_zone":         rotation.TimeZone,
		}
	}

	d.SetId("victorops_rotations_" + teamSlug)
	if err := d.Set("rotations", rotations); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
