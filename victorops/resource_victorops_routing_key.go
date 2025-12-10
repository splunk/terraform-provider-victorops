package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/victorops/go-victorops/victorops"
)

func resourceRoutingKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoutingKeyCreate,
		ReadContext:   resourceRoutingKeyRead,
		DeleteContext: resourceRoutingKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the routing key.",
			},
			"targets": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "A list of escalation policy slugs to route alerts to.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceRoutingKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)

	t := d.Get("targets").([]interface{})
	targets := make([]string, len(t))
	for i := range t {
		targets[i] = t[i].(string)
	}

	routingKey := &victorops.RoutingKey{
		RoutingKey: d.Get("name").(string),
		Targets:    targets,
	}

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	newRoutingKey, requestDetails, err := config.VictorOpsClient.CreateRoutingKey(routingKey)
	if err != nil {
		return diag.FromErr(err)
	}

	if requestDetails.StatusCode != 200 {
		return diag.Errorf("failed to create routing key (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	d.SetId(newRoutingKey.RoutingKey)
	return resourceRoutingKeyRead(ctx, d, m)
}

func resourceRoutingKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	name := d.Get("name").(string)
	if name == "" {
		name = d.Id()
	}

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	rk, _, err := config.VictorOpsClient.GetRoutingKey(name)
	if err != nil {
		return diag.FromErr(err)
	}

	if rk == nil {
		d.SetId("")
		return diags
	}

	d.SetId(rk.RoutingKey)

	if err := d.Set("name", rk.RoutingKey); err != nil {
		return diag.FromErr(err)
	}

	targets := []string{}
	for _, target := range rk.Targets {
		targets = append(targets, target.PolicySlug)
	}
	if err := d.Set("targets", targets); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceRoutingKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("deleting routing keys is not supported by the VictorOps API. Please delete in the UI and remove from Terraform state using 'terraform state rm'")
}
