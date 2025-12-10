package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/victorops/go-victorops/victorops"
)

// Provider returns the VictorOps Terraform provider
func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Your VictorOps API key.",
				DefaultFunc: schema.EnvDefaultFunc("VO_API_KEY", nil),
			},
			"api_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Your VictorOps API ID.",
				DefaultFunc: schema.EnvDefaultFunc("VO_API_ID", nil),
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The base url to use for api requests.",
				DefaultFunc: schema.EnvDefaultFunc("VO_BASE_URL", "https://api.victorops.com"),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"victorops_user":               resourceUser(),
			"victorops_team":               resourceTeam(),
			"victorops_team_membership":    resourceTeamMembership(),
			"victorops_contact":            resourceContact(),
			"victorops_escalation_policy":  resourceEscalationPolicy(),
			"victorops_routing_key":        resourceRoutingKey(),
			"victorops_user_contact_email": resourceUserContactEmail(),
			"victorops_user_contact_phone": resourceUserContactPhone(),
			"victorops_scheduled_override": resourceScheduledOverride(),
			"victorops_maintenance_mode":   resourceMaintenanceMode(),
			"victorops_alert_rule":         resourceAlertRule(),
			"victorops_user_paging_policy": resourceUserPagingPolicy(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"victorops_team_oncall_schedule": dataSourceTeamOncallSchedule(),
			"victorops_rotations":            dataSourceRotations(),
			"victorops_routing_keys":         dataSourceRoutingKeys(),
			"victorops_users":                dataSourceUsers(),
			"victorops_team_admins":          dataSourceTeamAdmins(),
			"victorops_user_devices":         dataSourceUserDevices(),
		},
	}

	p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return providerConfigure(ctx, d, p.TerraformVersion)
	}

	return p
}

func providerConfigure(ctx context.Context, d *schema.ResourceData, terraformVersion string) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiID := d.Get("api_id").(string)
	apiKey := d.Get("api_key").(string)
	baseURL := d.Get("base_url").(string)

	if apiID == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Missing API ID",
			Detail:   "The api_id must be set for the VictorOps provider",
		})
	}

	if apiKey == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Missing API Key",
			Detail:   "The api_key must be set for the VictorOps provider",
		})
	}

	if diags.HasError() {
		return nil, diags
	}

	victoropsClient := victorops.NewClient(apiID, apiKey, baseURL)

	config := Config{
		APIId:            apiID,
		APIKey:           apiKey,
		BaseURL:          baseURL,
		VictorOpsClient:  victoropsClient,
		TerraformVersion: terraformVersion,
	}

	return config, diags
}
