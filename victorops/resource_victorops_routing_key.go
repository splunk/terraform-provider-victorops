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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"targets": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},

		// Note: Routing keys CANNOT be deleted via the API
		Description: "Manages a VictorOps routing key. WARNING: The VictorOps API does not support deleting routing keys - destroy will fail. Use lifecycle { prevent_destroy = true } or remove manually from UI.",
	}
}

func resourceRoutingKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)

	// Convert the terraform config into an []string. There may be a better way to do this
	t := d.Get("targets").([]interface{})
	targets := make([]string, len(t))
	for i := range t {
		targets[i] = t[i].(string)
	}

	// Create the routing key object for the request
	routingKey := &victorops.RoutingKey{
		RoutingKey: d.Get("name").(string),
		Targets:    targets,
	}

	// Make the request
	newRoutingKey, requestDetails, err := config.VictorOpsClient.CreateRoutingKey(ctx, routingKey)
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

	// For import, use ID as the name if name is not set
	name := d.Get("name").(string)
	if name == "" {
		name = d.Id()
	}

	rk, _, err := config.VictorOpsClient.GetRoutingKey(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	if rk == nil {
		d.SetId("")
	} else {
		d.SetId(rk.RoutingKey)

		if err := d.Set("name", rk.RoutingKey); err != nil {
			return diag.FromErr(err)
		}

		// Convert the response targets to an array of strings that can be compared
		targets := []string{}
		for _, target := range rk.Targets {
			targets = append(targets, target.PolicySlug)
		}
		if err := d.Set("targets", targets); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceRoutingKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("deleting routing keys is not supported by the VictorOps API - please delete in the UI")
}
