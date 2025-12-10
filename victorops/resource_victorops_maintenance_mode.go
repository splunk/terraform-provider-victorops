package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceMaintenanceMode() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a maintenance mode in VictorOps/Splunk OnCall. Create starts maintenance mode, delete ends it. No updates are supported - all fields are ForceNew.",
		CreateContext: resourceMaintenanceModeCreate,
		ReadContext:   resourceMaintenanceModeRead,
		DeleteContext: resourceMaintenanceModeDelete,

		Schema: map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "RoutingKeys",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"RoutingKeys"}, false),
				Description:  "The type of maintenance mode. Currently only 'RoutingKeys' is supported.",
			},
			"routing_keys": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "The routing keys to put into maintenance mode. An empty list indicates global maintenance mode.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"purpose": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The reason for the maintenance mode.",
			},
			"instance_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance ID of the maintenance mode.",
			},
			"started_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp when maintenance mode was started (milliseconds from epoch).",
			},
			"started_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The username of the user who started maintenance mode.",
			},
			"is_global": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this is a global maintenance mode.",
			},
		},
	}
}

func resourceMaintenanceModeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	routingKeysRaw := d.Get("routing_keys").([]interface{})
	routingKeys := make([]string, len(routingKeysRaw))
	for i, v := range routingKeysRaw {
		routingKeys[i] = v.(string)
	}

	req := &StartMaintenanceModeRequest{
		Type:    d.Get("type").(string),
		Names:   routingKeys,
		Purpose: d.Get("purpose").(string),
	}

	state, err := apiClient.StartMaintenanceMode(req)
	if err != nil {
		return diag.FromErr(err)
	}

	// Find the newly created instance (it should be the last one in the list)
	if len(state.ActiveInstances) == 0 {
		return diag.Errorf("maintenance mode was started but no active instance found")
	}

	// Find instance by matching routing keys (or global flag)
	var newInstance *ActiveMaintenanceMode
	isGlobal := len(routingKeys) == 0

	for i := range state.ActiveInstances {
		inst := &state.ActiveInstances[i]
		if isGlobal && inst.IsGlobal {
			newInstance = inst
			break
		}
		if !isGlobal {
			// Check if targets match by collecting all names from targets
			matched := true
			targetMap := make(map[string]bool)
			for _, t := range inst.Targets {
				for _, name := range t.Names {
					targetMap[name] = true
				}
			}
			if len(targetMap) == len(routingKeys) {
				for _, rk := range routingKeys {
					if !targetMap[rk] {
						matched = false
						break
					}
				}
				if matched {
					newInstance = inst
					break
				}
			}
		}
	}

	// Fallback to the last instance
	if newInstance == nil {
		newInstance = &state.ActiveInstances[len(state.ActiveInstances)-1]
	}

	d.SetId(newInstance.InstanceID)
	if err := d.Set("instance_id", newInstance.InstanceID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("started_at", newInstance.StartedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("started_by", newInstance.StartedBy); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_global", newInstance.IsGlobal); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMaintenanceModeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	instance, err := apiClient.FindMaintenanceModeByID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if instance == nil {
		// Maintenance mode no longer exists
		d.SetId("")
		return diags
	}

	if err := d.Set("instance_id", instance.InstanceID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("started_at", instance.StartedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("started_by", instance.StartedBy); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_global", instance.IsGlobal); err != nil {
		return diag.FromErr(err)
	}

	// Map routing keys from targets (only if not global)
	if !instance.IsGlobal && len(instance.Targets) > 0 {
		var routingKeys []string
		for _, t := range instance.Targets {
			routingKeys = append(routingKeys, t.Names...)
		}
		if err := d.Set("routing_keys", routingKeys); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceMaintenanceModeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	_, err := apiClient.EndMaintenanceMode(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
