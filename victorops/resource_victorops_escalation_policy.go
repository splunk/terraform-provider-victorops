package victorops

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/victorops/go-victorops/victorops"
)

func resourceEscalationPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEscalationPolicyCreate,
		ReadContext:   resourceEscalationPolicyRead,
		DeleteContext: resourceEscalationPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The team slug for this escalation policy",
			},
			"ignore_custom_paging_policies": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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

		// Note: Escalation policies are IMMUTABLE via the API after creation.
		// Any changes will require recreation.
		Description: "Manages a VictorOps escalation policy. Note: The VictorOps API does not support updating escalation policies - any changes will force recreation.",
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
		entries, ok := step["entries"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("step %d: 'entries' must be a list, not a block (use 'entries = [...]' syntax)", i+1)
		}
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

func resourceEscalationPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)

	// Create the escalation policy object for the request
	ep, err := generateEscalationPolicyFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Make the request
	newEscalationPolicy, requestDetails, err := config.VictorOpsClient.CreateEscalationPolicy(ctx, ep)
	if err != nil {
		log.Printf(requestDetails.RequestBody)
		log.Printf(requestDetails.ResponseBody)
		return diag.FromErr(err)
	}

	if requestDetails.StatusCode != 200 {
		return diag.Errorf("failed to create escalation policy (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	d.SetId(newEscalationPolicy.ID)
	return resourceEscalationPolicyRead(ctx, d, m)
}

// mapExecutionTypeToSchemaType converts API execution type to schema type
func mapExecutionTypeToSchemaType(executionType string) string {
	switch executionType {
	case "user":
		return "user"
	case "email":
		return "email"
	case "rotation_group":
		return "rotationGroup"
	case "rotation_group_next":
		return "rotationGroupNext"
	case "rotation_group_previous":
		return "rotationGroupPrevious"
	case "webhook":
		return "webhook"
	case "policy_routing":
		return "targetPolicy"
	default:
		return executionType
	}
}

func resourceEscalationPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	// Make the request
	escalationPolicy, requestDetails, err := config.VictorOpsClient.GetEscalationPolicy(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if requestDetails.StatusCode == 404 {
		d.SetId("")
		return diags
	} else if requestDetails.StatusCode != 200 {
		return diag.Errorf("failed to get escalation policy (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	// Update our state with the refreshed state from the API
	if err := d.Set("name", escalationPolicy.Name); err != nil {
		return diag.FromErr(err)
	}

	// The GET single policy endpoint doesn't return team info (API limitation).
	// We need to look up the team from the policy list endpoint.
	teamID := escalationPolicy.TeamID
	if teamID == "" {
		// Look up team from policy list
		policyList, listDetails, listErr := config.VictorOpsClient.GetAllEscalationPolicies(ctx)
		if listErr != nil {
			return diag.FromErr(listErr)
		}
		if listDetails.StatusCode == 200 && policyList != nil {
			for _, p := range policyList.Policies {
				if p.Policy.Slug == d.Id() {
					teamID = p.Team.Slug
					break
				}
			}
		}
	}

	if err := d.Set("team_id", teamID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ignore_custom_paging_policies", escalationPolicy.IgnoreCustomPagingPolicies); err != nil {
		return diag.FromErr(err)
	}

	// Hydrate step blocks from API response (required for import)
	steps := make([]interface{}, len(escalationPolicy.Steps))
	for i, step := range escalationPolicy.Steps {
		entries := make([]interface{}, len(step.Entries))
		for j, entry := range step.Entries {
			entryMap := map[string]interface{}{
				"type": mapExecutionTypeToSchemaType(entry.ExecutionType),
			}

			// Map based on execution type
			switch entry.ExecutionType {
			case "user":
				if entry.User != nil {
					entryMap["username"] = entry.User["username"]
				}
			case "email":
				if entry.Email != nil {
					entryMap["address"] = entry.Email["address"]
				}
			case "rotation_group", "rotation_group_next", "rotation_group_previous":
				if entry.RotationGroup != nil {
					entryMap["slug"] = entry.RotationGroup["slug"]
				}
			case "webhook":
				if entry.Webhook != nil {
					entryMap["slug"] = entry.Webhook["slug"]
				}
			case "policy_routing":
				if entry.TargetPolicy != nil {
					entryMap["slug"] = entry.TargetPolicy["policySlug"]
				}
			}
			entries[j] = entryMap
		}
		steps[i] = map[string]interface{}{
			"timeout": step.Timeout,
			"entries": entries,
		}
	}

	if err := d.Set("step", steps); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceEscalationPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	// Make the request
	requestDetails, err := config.VictorOpsClient.DeleteEscalationPolicy(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if requestDetails.StatusCode != 200 {
		return diag.FromErr(fmt.Errorf("failed to delete escalation policy (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody))
	}

	return diags
}
