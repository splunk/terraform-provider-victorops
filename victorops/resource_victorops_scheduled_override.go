package victorops

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/victorops/go-victorops/victorops"
)

func resourceScheduledOverride() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScheduledOverrideCreate,
		ReadContext:   resourceScheduledOverrideRead,
		DeleteContext: resourceScheduledOverrideDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The username of the user who will be on-call during the override",
			},
			"start": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The start date/time of the override in ISO 8601 format (e.g., 2025-01-01T00:00:00Z)",
			},
			"end": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The end date/time of the override in ISO 8601 format (e.g., 2025-01-02T00:00:00Z)",
			},
			"public_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public ID of the scheduled override",
			},
			"timezone": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "UTC",
				ForceNew:    true,
				Description: "The timezone of the override (default: UTC)",
			},
		},
	}
}

func resourceScheduledOverrideCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(Config)

	payload := &victorops.ScheduledOverridePayload{
		Username: d.Get("username").(string),
		Start:    d.Get("start").(string),
		End:      d.Get("end").(string),
		Timezone: d.Get("timezone").(string),
	}

	override, details, err := config.VictorOpsClient.CreateScheduledOverride(ctx, payload)
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 {
		return diag.Errorf("failed to create scheduled override (%d): %s", details.StatusCode, details.ResponseBody)
	}

	// Check if PublicID was returned (required for subsequent operations)
	if override == nil || override.PublicID == "" {
		return diag.Errorf("scheduled override created but no publicId returned. Response: %s", details.ResponseBody)
	}

	d.SetId(override.PublicID)
	if err := d.Set("public_id", override.PublicID); err != nil {
		return diag.FromErr(err)
	}

	return resourceScheduledOverrideRead(ctx, d, m)
}

func resourceScheduledOverrideRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	override, details, err := config.VictorOpsClient.GetScheduledOverride(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode == 404 {
		d.SetId("")
		return diags
	} else if details.StatusCode != 200 {
		return diag.Errorf("failed to get scheduled override (%d): %s", details.StatusCode, details.ResponseBody)
	}

	if err := d.Set("username", override.GetUsername()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("start", override.Start); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("end", override.End); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("public_id", override.PublicID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("timezone", override.Timezone); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceScheduledOverrideDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)

	details, err := config.VictorOpsClient.DeleteScheduledOverride(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if details.StatusCode != 200 && details.StatusCode != 204 {
		return diag.FromErr(fmt.Errorf("failed to delete scheduled override (%d): %s", details.StatusCode, details.ResponseBody))
	}

	return diags
}
