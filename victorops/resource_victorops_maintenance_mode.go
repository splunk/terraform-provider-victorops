package victorops

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMaintenanceMode() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMaintenanceModeCreate,
		ReadContext:   resourceMaintenanceModeRead,
		DeleteContext: resourceMaintenanceModeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"routing_keys": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "List of routing keys to put in maintenance mode. Leave empty for global maintenance mode.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"purpose": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Description of the maintenance mode purpose",
			},
			"instance_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance ID of the maintenance mode",
			},
			"is_global": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this is a global maintenance mode",
			},
			"started_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "When the maintenance mode was started",
			},
			"started_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Who started the maintenance mode",
			},
		},

		// Note: This is an ephemeral resource - only active maintenance modes can be imported
		Description: "Manages a VictorOps maintenance mode. WARNING: This is an ephemeral resource - only active maintenance modes exist and can be imported.",
	}
}

func resourceMaintenanceModeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)

	routingKeysRaw := d.Get("routing_keys").([]interface{})
	routingKeys := make([]string, len(routingKeysRaw))
	for i, v := range routingKeysRaw {
		routingKeys[i] = v.(string)
	}

	purpose := d.Get("purpose").(string)

	state, details, err := config.VictorOpsClient.StartMaintenanceMode(ctx, routingKeys, purpose)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to start maintenance mode (%d): %s", details.StatusCode, details.ResponseBody)
	}

	if len(state.ActiveInstances) == 0 {
		return diag.Errorf("maintenance mode started but no instances returned")
	}

	// Use the most recently started instance (should be the last one)
	instance := state.ActiveInstances[len(state.ActiveInstances)-1]
	d.SetId(instance.InstanceID)
	if err := d.Set("instance_id", instance.InstanceID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_global", instance.IsGlobal); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("started_at", strconv.FormatInt(instance.StartedAt, 10)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("started_by", instance.StartedBy); err != nil {
		return diag.FromErr(err)
	}

	return resourceMaintenanceModeRead(ctx, d, m)
}

func resourceMaintenanceModeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	state, details, err := config.VictorOpsClient.GetMaintenanceModeState(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to get maintenance mode state (%d): %s", details.StatusCode, details.ResponseBody)
	}

	// Look for our instance in the active instances
	found := false
	for _, instance := range state.ActiveInstances {
		if instance.InstanceID == d.Id() {
			found = true
			if err := d.Set("instance_id", instance.InstanceID); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("is_global", instance.IsGlobal); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("started_at", strconv.FormatInt(instance.StartedAt, 10)); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("started_by", instance.StartedBy); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("purpose", instance.Purpose); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("routing_keys", instance.RoutingKeys); err != nil {
				return diag.FromErr(err)
			}
			break
		}
	}

	if !found {
		// Maintenance mode has ended
		d.SetId("")
	}

	return diags
}

func resourceMaintenanceModeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	_, details, err := config.VictorOpsClient.EndMaintenanceMode(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 && details.StatusCode != 204 {
		return diag.FromErr(fmt.Errorf("failed to end maintenance mode (%d): %s", details.StatusCode, details.ResponseBody))
	}

	return diags
}
