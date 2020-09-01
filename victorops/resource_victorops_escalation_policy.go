package victorops

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/victorops/go-victorops/victorops"
)

func resourceEscalationPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceEscalationPolicyCreate,
		Read:   resourceEscalationPolicyRead,
		Delete: resourceEscalationPolicyDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ignore_custom_paging_policies": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  "false",
				ForceNew: true,
			},
			"step": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"timeout": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
							ForceNew: true,
						},
						"entries": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeMap,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									Required:     true,
									ForceNew:     true,
									ValidateFunc: validation.StringInSlice([]string{"user", "email", "rotationGroup", "rotationGroupNext", "rotationGroupPrevious", "webhook", "targetPolicy"}, true),
								},
							},
						},
					},
				},
			},
		},
	}
}

func generateEscalationPolicyFromResourceData(d *schema.ResourceData) (*victorops.EscalationPolicy, error) {
	epsList := []victorops.EscalationPolicySteps{}
	steps := d.Get("step").([]interface{})

	// Crawl through each of the steps and add them to the escalation policy step list
	for i := range steps {
		step := steps[i].(map[string]interface{})
		entryList := []victorops.EscalationPolicyStepEntry{}

		// Crawl through all of the entries in this step and add them to the entries list
		entries := step["entries"].([]interface{})
		for i := range entries {
			e := entries[i].(map[string]interface{})
			t := e["type"].(string)

			if t == "user" {
				entry := victorops.EscalationPolicyStepEntry{
					ExecutionType: "user",
					User: map[string]string{
						"username": e["username"].(string),
					},
				}
				entryList = append(entryList, entry)
			} else if t == "email" {
				entry := victorops.EscalationPolicyStepEntry{
					ExecutionType: "email",
					Email: map[string]string{
						"address": e["address"].(string),
					},
				}
				entryList = append(entryList, entry)
			} else if t == "rotationGroup" {
				entry := victorops.EscalationPolicyStepEntry{
					ExecutionType: "rotation_group",
					RotationGroup: map[string]string{
						"slug": e["slug"].(string),
					},
				}
				entryList = append(entryList, entry)
			} else if t == "rotationGroupNext" {
				entry := victorops.EscalationPolicyStepEntry{
					ExecutionType: "rotation_group_next",
					RotationGroup: map[string]string{
						"slug": e["slug"].(string),
					},
				}
				entryList = append(entryList, entry)
			} else if t == "rotationGroupPrevious" {
				entry := victorops.EscalationPolicyStepEntry{
					ExecutionType: "rotation_group_previous",
					RotationGroup: map[string]string{
						"slug": e["slug"].(string),
					},
				}
				entryList = append(entryList, entry)
			} else if t == "webhook" {
				entry := victorops.EscalationPolicyStepEntry{
					ExecutionType: "webhook",
					Webhook: map[string]string{
						"slug": e["slug"].(string),
					},
				}
				entryList = append(entryList, entry)
			} else if t == "targetPolicy" {
				entry := victorops.EscalationPolicyStepEntry{
					ExecutionType: "policy_routing",
					TargetPolicy: map[string]string{
						"policySlug": e["slug"].(string),
					},
				}
				entryList = append(entryList, entry)
			}
		}

		eps := victorops.EscalationPolicySteps{
			Timeout: step["timeout"].(int),
			Entries: entryList,
		}

		epsList = append(epsList, eps)
	}

	return &victorops.EscalationPolicy{
		Name:                       d.Get("name").(string),
		TeamID:                     d.Get("team_id").(string),
		IgnoreCustomPagingPolicies: d.Get("ignore_custom_paging_policies").(bool),
		Steps:                      epsList,
	}, nil
}

func resourceEscalationPolicyCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)

	// Create the user object for the request
	ep, err := generateEscalationPolicyFromResourceData(d)
	if err != nil {
		return err
	}

	// Make the request
	newEscalationPolicy, requestDetails, err := config.VictorOpsClient.CreateEscalationPolicy(ep)
	if err != nil {
		log.Printf(requestDetails.RequestBody)
		log.Printf(requestDetails.ResponseBody)
		return err
	}

	if requestDetails.StatusCode != 200 {
		return fmt.Errorf("failed to create escaltion policy (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	d.SetId(newEscalationPolicy.ID)
	return resourceEscalationPolicyRead(d, m)
}

// TODO: Implement Read Escalation Policy w/ indepth comparison
func resourceEscalationPolicyRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)

	// Make the request
	escalationPolicy, requestDetails, err := config.VictorOpsClient.GetEscalationPolicy(d.Id())
	if err != nil {
		return err
	}

	if requestDetails.StatusCode == 404 {
		d.SetId("")
		return nil
	} else if requestDetails.StatusCode != 200 {
		return fmt.Errorf("failed to get escaltion policy (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	// Update our state with the refreshed state from the API
	err = d.Set("name", escalationPolicy.Name)
	if err != nil {
		return err
	}

	err = d.Set("ignore_custom_paging_policies", escalationPolicy.IgnoreCustomPagingPolicies)
	if err != nil {
		return err
	}

	return nil
}

func resourceEscalationPolicyDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)

	// Make the request
	requestDetails, err := config.VictorOpsClient.DeleteEscalationPolicy(d.Id())
	if err != nil {
		return err
	}

	if requestDetails.StatusCode != 200 {
		return fmt.Errorf("failed to delete escalation policy (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	return nil
}
