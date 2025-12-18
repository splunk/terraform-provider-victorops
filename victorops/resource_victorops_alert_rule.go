package victorops

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/victorops/go-victorops/victorops"
)

func resourceAlertRule() *schema.Resource {
	return &schema.Resource{
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
				Description: "The field in the alert to match against (e.g., 'monitoring_tool', 'entity_id')",
			},
			"alert_value_match": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The value or pattern to match in the alert field",
			},
			"match_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"WILDCARD", "REGEX"}, false),
				Description:  "The type of matching to use: WILDCARD or REGEX",
			},
			"stop_flag": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If true, stop processing further rules when this rule matches",
			},
			"rank": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The rank/priority of the alert rule (lower numbers are evaluated first)",
			},
			"routing_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The routing key to apply to matching alerts",
			},
			"notes": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description or notes about the alert rule",
			},
			"annotation": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Annotations or transformations to apply to matching alerts",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"annotation_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"i", "s", "u"}, false),
							Description:  "Type of annotation: 'i' for image, 's' for notes, 'u' for URL",
						},
						"field_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the field to annotate",
						},
						"field_value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The value for the annotation",
						},
						"flags": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "0 for annotation, 1 for transformation",
						},
					},
				},
			},
		},
	}
}

func buildAlertRulePayload(d *schema.ResourceData) *victorops.AlertRulePayload {
	payload := &victorops.AlertRulePayload{
		AlertField:      d.Get("alert_field").(string),
		AlertValueMatch: d.Get("alert_value_match").(string),
		MatchType:       d.Get("match_type").(string),
		StopFlag:        d.Get("stop_flag").(bool),
		Rank:            d.Get("rank").(int),
		RoutingKey:      d.Get("routing_key").(string),
		Notes:           d.Get("notes").(string),
		Annotations:     []victorops.AlertAnnotationPayload{}, // Always include, even if empty
	}

	// Convert annotations list
	if v, ok := d.GetOk("annotation"); ok {
		annotationsList := v.([]interface{})
		annotations := make([]victorops.AlertAnnotationPayload, len(annotationsList))
		for i, ann := range annotationsList {
			annMap := ann.(map[string]interface{})
			annotations[i] = victorops.AlertAnnotationPayload{
				AnnotationType: annMap["annotation_type"].(string),
				FieldName:      annMap["field_name"].(string),
				FieldValue:     annMap["field_value"].(string),
				Flags:          annMap["flags"].(int),
			}
		}
		payload.Annotations = annotations
	}

	return payload
}

func resourceAlertRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)

	payload := buildAlertRulePayload(d)

	rule, details, err := config.VictorOpsClient.CreateAlertRule(ctx, payload)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 && details.StatusCode != 201 {
		return diag.Errorf("failed to create alert rule (%d): %s", details.StatusCode, details.ResponseBody)
	}

	d.SetId(strconv.FormatInt(rule.ID, 10))

	return resourceAlertRuleRead(ctx, d, m)
}

func resourceAlertRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	rules, details, err := config.VictorOpsClient.ListAlertRules(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to list alert rules (%d): %s", details.StatusCode, details.ResponseBody)
	}

	// Find our rule by ID
	var found *victorops.AlertRule
	targetID, _ := strconv.ParseInt(d.Id(), 10, 64)
	for i := range rules {
		if rules[i].ID == targetID {
			found = &rules[i]
			break
		}
	}

	if found == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("alert_field", found.AlertField); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("alert_value_match", found.AlertValueMatch); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("match_type", found.MatchType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("stop_flag", found.StopFlag); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("rank", found.Rank); err != nil {
		return diag.FromErr(err)
	}
	// API returns routeKey (not routingKey)
	if err := d.Set("routing_key", found.RouteKey); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("notes", found.Notes); err != nil {
		return diag.FromErr(err)
	}

	// Convert annotations from API response
	if found.Annotations != nil && len(found.Annotations) > 0 {
		annotations := make([]interface{}, len(found.Annotations))
		for i, ann := range found.Annotations {
			annotations[i] = map[string]interface{}{
				"annotation_type": ann.AnnotationType,
				"field_name":      ann.FieldName,
				"field_value":     ann.FieldValue,
				"flags":           ann.Flags,
			}
		}
		if err := d.Set("annotation", annotations); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceAlertRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)

	payload := buildAlertRulePayload(d)

	_, details, err := config.VictorOpsClient.UpdateAlertRule(ctx, d.Id(), payload)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to update alert rule (%d): %s", details.StatusCode, details.ResponseBody)
	}

	return resourceAlertRuleRead(ctx, d, m)
}

func resourceAlertRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	_, details, err := config.VictorOpsClient.DeleteAlertRule(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 && details.StatusCode != 204 {
		return diag.FromErr(fmt.Errorf("failed to delete alert rule (%d): %s", details.StatusCode, details.ResponseBody))
	}

	return diags
}
