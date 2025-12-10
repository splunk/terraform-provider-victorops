package victorops

import (
	"context"
	"fmt"
	"log"
	"strings"

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
			StateContext: resourceEscalationPolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the escalation policy.",
			},
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The slug/ID of the team this policy belongs to.",
			},
			"ignore_custom_paging_policies": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Whether to ignore custom paging policies when this policy is triggered.",
			},
			"step": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "The escalation steps for this policy.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"timeout": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							ForceNew:    true,
							Description: "The timeout in minutes before escalating to the next step.",
						},
						"entries": {
							Type:        schema.TypeList,
							Required:    true,
							ForceNew:    true,
							Description: "The entries (targets) for this escalation step.",
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

	for i := range steps {
		step := steps[i].(map[string]interface{})
		entryList := []victorops.EscalationPolicyStepEntry{}

		entries := step["entries"].([]interface{})
		for j := range entries {
			e := entries[j].(map[string]interface{})
			t := e["type"].(string)

			switch t {
			case "user":
				entry := victorops.EscalationPolicyStepEntry{
					ExecutionType: "user",
					User: map[string]string{
						"username": e["username"].(string),
					},
				}
				entryList = append(entryList, entry)
			case "email":
				entry := victorops.EscalationPolicyStepEntry{
					ExecutionType: "email",
					Email: map[string]string{
						"address": e["address"].(string),
					},
				}
				entryList = append(entryList, entry)
			case "rotationGroup":
				entry := victorops.EscalationPolicyStepEntry{
					ExecutionType: "rotation_group",
					RotationGroup: map[string]string{
						"slug": e["slug"].(string),
					},
				}
				entryList = append(entryList, entry)
			case "rotationGroupNext":
				entry := victorops.EscalationPolicyStepEntry{
					ExecutionType: "rotation_group_next",
					RotationGroup: map[string]string{
						"slug": e["slug"].(string),
					},
				}
				entryList = append(entryList, entry)
			case "rotationGroupPrevious":
				entry := victorops.EscalationPolicyStepEntry{
					ExecutionType: "rotation_group_previous",
					RotationGroup: map[string]string{
						"slug": e["slug"].(string),
					},
				}
				entryList = append(entryList, entry)
			case "webhook":
				entry := victorops.EscalationPolicyStepEntry{
					ExecutionType: "webhook",
					Webhook: map[string]string{
						"slug": e["slug"].(string),
					},
				}
				entryList = append(entryList, entry)
			case "targetPolicy":
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

	ep, err := generateEscalationPolicyFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	newEscalationPolicy, requestDetails, err := config.VictorOpsClient.CreateEscalationPolicy(ep)
	if err != nil {
		log.Printf("[ERROR] Request body: %s", requestDetails.RequestBody)
		log.Printf("[ERROR] Response body: %s", requestDetails.ResponseBody)
		return diag.FromErr(err)
	}

	if requestDetails.StatusCode != 200 {
		return diag.Errorf("failed to create escalation policy (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	d.SetId(newEscalationPolicy.ID)
	return resourceEscalationPolicyRead(ctx, d, m)
}

func resourceEscalationPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	escalationPolicy, requestDetails, err := config.VictorOpsClient.GetEscalationPolicy(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if requestDetails.StatusCode == 404 {
		d.SetId("")
		return diags
	}
	if requestDetails.StatusCode != 200 {
		return diag.Errorf("failed to get escalation policy (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	if err := d.Set("name", escalationPolicy.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ignore_custom_paging_policies", escalationPolicy.IgnoreCustomPagingPolicies); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceEscalationPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	// Wait for rate limiter before making API request
	if err := WaitForRateLimitWithContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	requestDetails, err := config.VictorOpsClient.DeleteEscalationPolicy(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if requestDetails.StatusCode != 200 {
		return diag.Errorf("failed to delete escalation policy (%d): %s", requestDetails.StatusCode, requestDetails.ResponseBody)
	}

	d.SetId("")
	return diags
}

func resourceEscalationPolicyImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	idAttr := strings.SplitN(d.Id(), "/", 2)
	if len(idAttr) != 2 {
		return nil, fmt.Errorf("invalid id %q specified, should be in format \"team_id/policy_id\" for import", d.Id())
	}

	teamID := idAttr[0]
	policyID := idAttr[1]

	d.Set("team_id", teamID)
	d.SetId(policyID)

	return []*schema.ResourceData{d}, nil
}
