package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/victorops/go-victorops/victorops"
)

// Provider defines the VO provider
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
			"victorops_scheduled_override": resourceScheduledOverride(),
			"victorops_maintenance_mode":   resourceMaintenanceMode(),
			"victorops_alert_rule":         resourceAlertRule(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"victorops_user":               dataSourceUser(),
			"victorops_users":              dataSourceUsers(),
			"victorops_team":               dataSourceTeam(),
			"victorops_teams":              dataSourceTeams(),
			"victorops_escalation_policy":  dataSourceEscalationPolicy(),
			"victorops_escalation_policies": dataSourceEscalationPolicies(),
			"victorops_routing_keys":       dataSourceRoutingKeys(),
			"victorops_rotations":          dataSourceRotations(),
			"victorops_on_call":            dataSourceOnCall(),
		},
	}

	p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return providerConfigure(ctx, d)
	}

	return p
}

func providerConfigure(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Create a real victorops client from the SDK
	victoropsClient := victorops.NewClient(data.Get("api_id").(string), data.Get("api_key").(string), data.Get("base_url").(string))

	config := Config{
		APIId:           data.Get("api_id").(string),
		APIKey:          data.Get("api_key").(string),
		BaseURL:         data.Get("base_url").(string),
		VictorOpsClient: victoropsClient,
	}

	return config, diags
}
