package victorops

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/victorops/go-victorops/victorops"
)

func resourceRoutingKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceRoutingKeyCreate,
		Read:   resourceRoutingKeyRead,
		Delete: resourceRoutingKeyDelete,

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
	}
}

func resourceRoutingKeyCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)

	// Convert the terrform config into an []string. There may be a better way to do this
	t := d.Get("targets").([]interface{})
	targets := make([]string, len(t))
	for i := range t {
		targets[i] = t[i].(string)
	}

	// Create the user object for the request
	routingKey := &victorops.RoutingKey{
		RoutingKey: d.Get("name").(string),
		Targets:    targets,
	}

	// Make the request
	newRoutingKey, requestDetails, err := config.VictorOpsClient.CreateRoutingKey(routingKey)
	if err != nil {
		return err
	}

	if requestDetails.StatusCode != 200 {
		return fmt.Errorf("failed to create routing key (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	d.SetId(newRoutingKey.RoutingKey)
	return resourceRoutingKeyRead(d, m)
}

func resourceRoutingKeyRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)

	rk, _, err := config.VictorOpsClient.GetRoutingKey(d.Get("name").(string))
	if err != nil {
		return err
	}

	if rk == nil {
		d.SetId("")
	} else {
		d.SetId(rk.RoutingKey)

		// Convert the response targets to an array of strings that can be compared
		targets := []string{}
		for _, target := range rk.Targets {
			targets = append(targets, target.PolicySlug)
		}
		d.Set("targets", targets)
	}

	return nil
}

func resourceRoutingKeyDelete(d *schema.ResourceData, m interface{}) error {
	return fmt.Errorf("deleting routing keys not yet implemented in the API, please delete in the UI")
}

// todo: Add acceptance tests
