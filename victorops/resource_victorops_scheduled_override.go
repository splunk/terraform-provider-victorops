package victorops

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceScheduledOverride() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a scheduled override in VictorOps/Splunk OnCall.",
		CreateContext: resourceScheduledOverrideCreate,
		ReadContext:   resourceScheduledOverrideRead,
		UpdateContext: resourceScheduledOverrideUpdate,
		DeleteContext: resourceScheduledOverrideDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The username of the user being overridden.",
			},
			"timezone": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(TimeZoneList(), false),
				Description:  "The timezone for the override (e.g., 'America/New_York').",
			},
			"start": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The start time of the override in ISO 8601 format (e.g., '2024-01-01T08:00:00Z').",
			},
			"end": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The end time of the override in ISO 8601 format (e.g., '2024-01-01T17:00:00Z').",
			},
			"public_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public ID of the scheduled override.",
			},
			"assignment": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Assignments for this override.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"policy_slug": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The slug of the escalation policy to assign.",
						},
						"username": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The username of the person taking the override.",
						},
						"accept_overlap": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Accept overlap with existing scheduled overrides.",
						},
					},
				},
			},
		},
	}
}

func resourceScheduledOverrideCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	req := &ScheduledOverrideCreateRequest{
		Username: d.Get("username").(string),
		Timezone: d.Get("timezone").(string),
		Start:    d.Get("start").(string),
		End:      d.Get("end").(string),
	}

	override, err := apiClient.CreateScheduledOverride(req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(override.PublicID)
	if err := d.Set("public_id", override.PublicID); err != nil {
		return diag.FromErr(err)
	}

	// Create assignments if specified
	if v, ok := d.GetOk("assignment"); ok {
		assignments := v.([]interface{})
		for _, a := range assignments {
			assignment := a.(map[string]interface{})
			policySlug := assignment["policy_slug"].(string)
			assignReq := &AssignmentUpdateRequest{
				Username:      assignment["username"].(string),
				AcceptOverlap: assignment["accept_overlap"].(bool),
			}
			_, err := apiClient.UpdateAssignment(override.PublicID, policySlug, assignReq)
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to create assignment for policy %s: %w", policySlug, err))
			}
		}
	}

	return resourceScheduledOverrideRead(ctx, d, m)
}

func resourceScheduledOverrideRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	override, err := apiClient.GetScheduledOverride(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if override == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("public_id", override.PublicID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("timezone", override.Timezone); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("start", override.Start); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("end", override.End); err != nil {
		return diag.FromErr(err)
	}
	if override.User != nil {
		if err := d.Set("username", override.User.Username); err != nil {
			return diag.FromErr(err)
		}
	}

	// Only update assignments if they were explicitly configured
	// The API returns all assignments for the user, not just the ones we created
	// We preserve the configured assignments to avoid drift
	if _, ok := d.GetOk("assignment"); ok {
		// Keep existing configured assignments - don't overwrite from API
		// The API returns all policies the user is part of, not just ones we created
	}

	return diags
}

func resourceScheduledOverrideUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	if d.HasChange("assignment") {
		old, new := d.GetChange("assignment")
		oldAssignments := old.([]interface{})
		newAssignments := new.([]interface{})

		// Create a map of old assignments by policy_slug
		oldMap := make(map[string]bool)
		for _, a := range oldAssignments {
			assignment := a.(map[string]interface{})
			oldMap[assignment["policy_slug"].(string)] = true
		}

		// Create a map of new assignments by policy_slug
		newMap := make(map[string]map[string]interface{})
		for _, a := range newAssignments {
			assignment := a.(map[string]interface{})
			newMap[assignment["policy_slug"].(string)] = assignment
		}

		// Delete assignments that are no longer present
		for policySlug := range oldMap {
			if _, exists := newMap[policySlug]; !exists {
				if err := apiClient.DeleteAssignment(d.Id(), policySlug); err != nil {
					return diag.FromErr(fmt.Errorf("failed to delete assignment for policy %s: %w", policySlug, err))
				}
			}
		}

		// Create or update new assignments
		for policySlug, assignment := range newMap {
			assignReq := &AssignmentUpdateRequest{
				Username:      assignment["username"].(string),
				AcceptOverlap: assignment["accept_overlap"].(bool),
			}
			_, err := apiClient.UpdateAssignment(d.Id(), policySlug, assignReq)
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to update assignment for policy %s: %w", policySlug, err))
			}
		}
	}

	return resourceScheduledOverrideRead(ctx, d, m)
}

func resourceScheduledOverrideDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	// Delete assignments first
	if v, ok := d.GetOk("assignment"); ok {
		assignments := v.([]interface{})
		for _, a := range assignments {
			assignment := a.(map[string]interface{})
			policySlug := assignment["policy_slug"].(string)
			// Ignore errors during delete - assignment might already be gone
			apiClient.DeleteAssignment(d.Id(), policySlug)
		}
	}

	if err := apiClient.DeleteScheduledOverride(d.Id()); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
