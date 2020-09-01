package victorops

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/victorops/go-victorops/victorops"
)

// Provider defines the VO provider
func Provider() terraform.ResourceProvider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
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
			"victorops_user":              resourceUser(),
			"victorops_team":              resourceTeam(),
			"victorops_team_membership":   resourceTeamMembership(),
			"victorops_contact":           resourceContact(),
			"victorops_escalation_policy": resourceEscalationPolicy(),
			"victorops_routing_key":       resourceRoutingKey(),
		},
	}

	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(d, terraformVersion)
	}

	return p
}

func providerConfigure(data *schema.ResourceData, terraformVersion string) (interface{}, error) {

	// Create a real victorops client from the SDK
	victoropsClient := victorops.NewClient(data.Get("api_id").(string), data.Get("api_key").(string), data.Get("base_url").(string))

	config := Config{
		APIId:           data.Get("api_id").(string),
		APIKey:          data.Get("api_key").(string),
		BaseURL:         data.Get("base_url").(string),
		VictorOpsClient: victoropsClient,
	}

	return config, nil
}
