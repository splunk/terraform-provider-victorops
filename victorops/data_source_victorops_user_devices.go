package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUserDevices() *schema.Resource {
	return &schema.Resource{
		Description: "Get a user's contact devices. Note: Devices can only be created via the mobile app.",
		ReadContext: dataSourceUserDevicesRead,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The username of the user.",
			},
			"devices": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The user's devices.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ext_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The external ID of the device.",
						},
						"device_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of device.",
						},
						"label": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The label of the device.",
						},
					},
				},
			},
		},
	}
}

func dataSourceUserDevicesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	username := d.Get("username").(string)

	devices, err := apiClient.GetUserDevices(username)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(username)

	deviceList := make([]map[string]interface{}, len(devices))
	for i, dev := range devices {
		deviceList[i] = map[string]interface{}{
			"ext_id":      dev.ExtID,
			"device_type": dev.DeviceType,
			"label":       dev.Label,
		}
	}

	if err := d.Set("devices", deviceList); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
