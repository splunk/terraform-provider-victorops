package victorops

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceUserPagingPolicy() *schema.Resource {
	return &schema.Resource{
		Description: `Manages a user's paging policy in VictorOps/Splunk OnCall.

This resource manages the entire paging policy for a user, including all steps and rules.
Note: This resource fully owns the paging policy. Any steps/rules not defined in Terraform
will be removed when the policy is applied.`,
		CreateContext: resourceUserPagingPolicyCreate,
		ReadContext:   resourceUserPagingPolicyRead,
		UpdateContext: resourceUserPagingPolicyUpdate,
		DeleteContext: resourceUserPagingPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The username whose paging policy to manage.",
			},
			"step": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "The steps in the paging policy, ordered by execution sequence.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"timeout": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntInSlice([]int{0, 1, 5, 10, 15, 20, 25, 30, 45, 60}),
							Description:  "The timeout in minutes before escalating to the next step. Valid values: 0, 1, 5, 10, 15, 20, 25, 30, 45, 60.",
						},
						"rule": {
							Type:        schema.TypeList,
							Required:    true,
							MinItems:    1,
							Description: "The rules within this step.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"push", "email", "phone", "sms"}, false),
										Description:  "The type of notification: 'push', 'email', 'phone', or 'sms'.",
									},
									"contact": {
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										Description: "The contact for this rule (required for email, phone, sms types).",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"id": {
													Type:        schema.TypeInt,
													Required:    true,
													Description: "The contact ID.",
												},
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{"email", "phone"}, false),
													Description:  "The contact type: 'email' or 'phone'.",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceUserPagingPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)
	username := d.Get("username").(string)

	// First, clear any existing policy
	existingPolicy, err := apiClient.GetUserPagingPolicy(username)
	if err != nil {
		return diag.FromErr(err)
	}
	if existingPolicy != nil {
		// Delete existing steps in reverse order using Index field
		for i := len(existingPolicy.Steps) - 1; i >= 0; i-- {
			if err := apiClient.DeletePagingPolicyStep(username, existingPolicy.Steps[i].Index); err != nil {
				return diag.FromErr(fmt.Errorf("failed to delete existing step %d: %w", existingPolicy.Steps[i].Index, err))
			}
		}
	}

	// Create new steps
	steps := d.Get("step").([]interface{})
	for _, s := range steps {
		stepData := s.(map[string]interface{})
		stepReq := &PagingPolicyStepCreateRequest{
			Timeout: stepData["timeout"].(int),
		}

		step, err := apiClient.CreatePagingPolicyStep(username, stepReq)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to create step: %w", err))
		}

		// Create rules for this step
		rules := stepData["rule"].([]interface{})
		for _, r := range rules {
			ruleData := r.(map[string]interface{})
			ruleReq := &PagingPolicyRuleCreateRequest{
				Type: ruleData["type"].(string),
			}

			// Handle contact block
			if contactList, ok := ruleData["contact"].([]interface{}); ok && len(contactList) > 0 {
				contactData := contactList[0].(map[string]interface{})
				ruleReq.Contact = &Contact{
					ID:   contactData["id"].(int),
					Type: contactData["type"].(string),
				}
			}

			_, err := apiClient.CreatePagingPolicyRule(username, step.Index, ruleReq)
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to create rule: %w", err))
			}
		}
	}

	d.SetId(username)
	return resourceUserPagingPolicyRead(ctx, d, m)
}

func resourceUserPagingPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)
	username := d.Id()

	policy, err := apiClient.GetUserPagingPolicy(username)
	if err != nil {
		return diag.FromErr(err)
	}

	if policy == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("username", username); err != nil {
		return diag.FromErr(err)
	}

	// Map steps and rules
	steps := make([]map[string]interface{}, len(policy.Steps))
	for i, s := range policy.Steps {
		rules := make([]map[string]interface{}, len(s.Rules))
		for j, r := range s.Rules {
			ruleMap := map[string]interface{}{
				"type": r.Type,
			}

			// Handle contact object
			if r.Contact != nil {
				ruleMap["contact"] = []map[string]interface{}{
					{
						"id":   r.Contact.ID,
						"type": r.Contact.Type,
					},
				}
			} else {
				ruleMap["contact"] = []map[string]interface{}{}
			}

			rules[j] = ruleMap
		}
		steps[i] = map[string]interface{}{
			"timeout": s.Timeout,
			"rule":    rules,
		}
	}

	if err := d.Set("step", steps); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceUserPagingPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// For simplicity, we recreate the entire policy on update
	// This ensures proper ordering of steps and rules
	return resourceUserPagingPolicyCreate(ctx, d, m)
}

func resourceUserPagingPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)
	username := d.Id()

	policy, err := apiClient.GetUserPagingPolicy(username)
	if err != nil {
		return diag.FromErr(err)
	}

	if policy != nil {
		// Delete all steps in reverse order using Index field
		for i := len(policy.Steps) - 1; i >= 0; i-- {
			if err := apiClient.DeletePagingPolicyStep(username, policy.Steps[i].Index); err != nil {
				return diag.FromErr(fmt.Errorf("failed to delete step %d: %w", policy.Steps[i].Index, err))
			}
		}
	}

	d.SetId("")
	return diags
}
