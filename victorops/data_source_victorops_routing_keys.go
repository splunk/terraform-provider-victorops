package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRoutingKeys() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRoutingKeysRead,

		Schema: map[string]*schema.Schema{
			"routing_keys": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of all routing keys",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The routing key name",
						},
						"targets": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of policy slugs this routing key targets",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceRoutingKeysRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	routingKeysResp, details, err := config.VictorOpsClient.GetAllRoutingKeys(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to get routing keys (%d): %s", details.StatusCode, details.ResponseBody)
	}

	routingKeys := make([]map[string]interface{}, len(routingKeysResp.RoutingKeys))
	for i, rk := range routingKeysResp.RoutingKeys {
		targets := make([]string, len(rk.Targets))
		for j, target := range rk.Targets {
			targets[j] = target.PolicySlug
		}

		routingKeys[i] = map[string]interface{}{
			"name":    rk.RoutingKey,
			"targets": targets,
		}
	}

	d.SetId("victorops_routing_keys")
	if err := d.Set("routing_keys", routingKeys); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
