package victorops

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRoutingKeys() *schema.Resource {
	return &schema.Resource{
		Description: "Get all routing keys for the organization.",
		ReadContext: dataSourceRoutingKeysRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional filter pattern to match routing key names (supports * wildcard).",
			},
			"routing_keys": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The routing keys.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the routing key.",
						},
						"is_default": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this is the default routing key.",
						},
						"targets": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The policy targets for this routing key.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"policy_slug": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The slug of the target policy.",
									},
									"policy_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the target policy.",
									},
								},
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
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	filter := d.Get("filter").(string)

	routingKeys, err := apiClient.GetRoutingKeys()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("routing_keys")

	routingKeyList := make([]map[string]interface{}, 0)
	for _, rk := range routingKeys {
		// Apply filter if specified
		if filter != "" && !matchWildcard(filter, rk.RoutingKey) {
			continue
		}

		targets := make([]map[string]interface{}, len(rk.Targets))
		for j, t := range rk.Targets {
			targets[j] = map[string]interface{}{
				"policy_slug": t.PolicySlug,
				"policy_name": t.PolicyName,
			}
		}
		routingKeyList = append(routingKeyList, map[string]interface{}{
			"name":       rk.RoutingKey,
			"is_default": rk.IsDefault,
			"targets":    targets,
		})
	}

	if err := d.Set("routing_keys", routingKeyList); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// matchWildcard performs simple wildcard matching (supports * for any characters)
func matchWildcard(pattern, str string) bool {
	if pattern == "" {
		return true
	}
	if pattern == "*" {
		return true
	}

	// Handle prefix wildcard
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(str, pattern[1:len(pattern)-1])
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(str, pattern[1:])
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(str, pattern[:len(pattern)-1])
	}

	return str == pattern
}
