package victorops

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAlertRule() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages an alert rule in VictorOps/Splunk OnCall. Note: The API auto-shifts ranks on create, so rank may drift if managed externally. Consider using ignore_changes lifecycle or managing all rules in Terraform.",
		CreateContext: resourceAlertRuleCreate,
		ReadContext:   resourceAlertRuleRead,
		UpdateContext: resourceAlertRuleUpdate,
		DeleteContext: resourceAlertRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"alert_field": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The field in the alert to match against (e.g., 'host_name', 'entity_id').",
			},
			"alert_value_match": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The value pattern to match (supports wildcards or regex based on match_type).",
			},
			"match_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"WILDCARD", "REGEX"}, false),
				Description:  "The type of matching: 'WILDCARD' or 'REGEX'.",
			},
			"routing_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The routing key to route matching alerts to.",
			},
			"rank": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The rank (priority) of this rule. Lower ranks are evaluated first. Note: The API may auto-shift ranks.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Suppress diff if rank is not explicitly set
					if new == "0" || new == "" {
						return true
					}
					return false
				},
			},
			"stop_flag": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If true, stop processing rules after this one matches.",
			},
			"notes": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description or notes about this alert rule.",
			},
			"annotation": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Annotations to add or transform on matching alerts.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"annotation_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"s", "i", "u"}, false),
							Description:  "The type of annotation: 's' (notes/string), 'i' (image), 'u' (url).",
						},
						"field_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The field name for the annotation.",
						},
						"field_value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The field value for the annotation.",
						},
						"flags": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Flags: 0 for annotation, 1 for transformation.",
						},
					},
				},
			},
			"last_updated": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp when the rule was last updated.",
			},
			"last_updated_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user who last updated the rule.",
			},
		},
	}
}

func resourceAlertRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	annotations := expandAlertAnnotations(d.Get("annotation").([]interface{}))
	if annotations == nil {
		annotations = []AlertAnnotation{} // API requires empty array, not null
	}

	rank := d.Get("rank").(int)
	if rank == 0 {
		rank = 1 // Default to rank 1 if not specified
	}

	req := &AlertRuleCreateRequest{
		AlertField:      d.Get("alert_field").(string),
		AlertValueMatch: d.Get("alert_value_match").(string),
		MatchType:       d.Get("match_type").(string),
		Rank:            rank,
		StopFlag:        d.Get("stop_flag").(bool),
		Notes:           d.Get("notes").(string),
		RoutingKey:      d.Get("routing_key").(string),
		Annotations:     annotations,
	}

	rule, err := apiClient.CreateAlertRule(req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(rule.ID))

	return resourceAlertRuleRead(ctx, d, m)
}

func resourceAlertRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	ruleID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	rule, err := apiClient.GetAlertRule(ruleID)
	if err != nil {
		return diag.FromErr(err)
	}

	if rule == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("alert_field", rule.AlertField); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("alert_value_match", rule.AlertValueMatch); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("match_type", rule.MatchType); err != nil {
		return diag.FromErr(err)
	}
	// Note: API response uses "routeKey" field (not "routingKey" which is used in requests)
	// Only update if the API returns a value; otherwise preserve configured value
	if rule.RouteKey != "" {
		if err := d.Set("routing_key", rule.RouteKey); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("rank", rule.Rank); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("stop_flag", rule.StopFlag); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("notes", rule.Notes); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("last_updated", rule.LastUpdated); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("last_updated_by", rule.LastUpdatedBy); err != nil {
		return diag.FromErr(err)
	}

	annotations := flattenAlertAnnotations(rule.Annotations)
	if err := d.Set("annotation", annotations); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceAlertRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	ruleID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	annotations := expandAlertAnnotations(d.Get("annotation").([]interface{}))
	if annotations == nil {
		annotations = []AlertAnnotation{} // API requires empty array, not null
	}

	rank := d.Get("rank").(int)
	if rank == 0 {
		rank = 1 // Default to rank 1 if not specified
	}

	req := &AlertRuleCreateRequest{
		AlertField:      d.Get("alert_field").(string),
		AlertValueMatch: d.Get("alert_value_match").(string),
		MatchType:       d.Get("match_type").(string),
		Rank:            rank,
		StopFlag:        d.Get("stop_flag").(bool),
		Notes:           d.Get("notes").(string),
		RoutingKey:      d.Get("routing_key").(string),
		Annotations:     annotations,
	}

	_, err = apiClient.UpdateAlertRule(ruleID, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceAlertRuleRead(ctx, d, m)
}

func resourceAlertRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	ruleID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := apiClient.DeleteAlertRule(ruleID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}

func expandAlertAnnotations(input []interface{}) []AlertAnnotation {
	if len(input) == 0 {
		return nil
	}

	annotations := make([]AlertAnnotation, len(input))
	for i, v := range input {
		m := v.(map[string]interface{})
		annotations[i] = AlertAnnotation{
			AnnotationType: m["annotation_type"].(string),
			FieldName:      m["field_name"].(string),
			FieldValue:     m["field_value"].(string),
			Flags:          m["flags"].(int),
		}
	}

	return annotations
}

func flattenAlertAnnotations(annotations []AlertAnnotation) []map[string]interface{} {
	if len(annotations) == 0 {
		return nil
	}

	result := make([]map[string]interface{}, len(annotations))
	for i, a := range annotations {
		result[i] = map[string]interface{}{
			"annotation_type": a.AnnotationType,
			"field_name":      a.FieldName,
			"field_value":     a.FieldValue,
			"flags":           a.Flags,
		}
	}

	return result
}
