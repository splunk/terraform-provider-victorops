package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTeamOncallSchedule() *schema.Resource {
	return &schema.Resource{
		Description: "Get a team's on-call schedule.",
		ReadContext: dataSourceTeamOncallScheduleRead,

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The slug/ID of the team.",
			},
			"days_forward": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     14,
				Description: "The number of days in the future to include in the schedule.",
			},
			"days_skip": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The number of days to skip.",
			},
			"team_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the team.",
			},
			"team_slug": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The slug of the team.",
			},
			"schedules": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The policy schedules for the team.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"policy": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The escalation policy for this schedule.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the escalation policy.",
									},
									"slug": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The slug of the escalation policy.",
									},
								},
							},
						},
						"schedule": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The on-call entries.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"oncall_user": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The username of the on-call user.",
									},
									"override_oncall_user": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The username of the override on-call user (if any).",
									},
									"oncall_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The type of on-call.",
									},
									"rotation_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the rotation.",
									},
									"shift_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the shift.",
									},
									"shift_roll": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The shift roll time (ISO8601).",
									},
									"rolls": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The on-call rolls.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"start": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Start time (ISO8601).",
												},
												"end": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "End time (ISO8601).",
												},
												"oncall_user": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Username of the on-call user.",
												},
												"is_roll": {
													Type:        schema.TypeBool,
													Computed:    true,
													Description: "Whether this is a roll.",
												},
											},
										},
									},
								},
							},
						},
						"overrides": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The on-call overrides.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"orig_oncall_user": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Original on-call user.",
									},
									"override_oncall_user": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Override on-call user.",
									},
									"start": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Start time (ISO8601).",
									},
									"end": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "End time (ISO8601).",
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

func dataSourceTeamOncallScheduleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	teamID := d.Get("team_id").(string)
	daysForward := d.Get("days_forward").(int)
	daysSkip := d.Get("days_skip").(int)

	schedule, err := apiClient.GetTeamOncallSchedule(teamID, daysForward, daysSkip)
	if err != nil {
		return diag.FromErr(err)
	}

	if schedule == nil {
		return diag.Errorf("team %s not found", teamID)
	}

	d.SetId(teamID)

	// Set team info
	if schedule.Team != nil {
		if err := d.Set("team_name", schedule.Team.Name); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("team_slug", schedule.Team.Slug); err != nil {
			return diag.FromErr(err)
		}
	}

	// Map policy schedules
	schedulesList := make([]map[string]interface{}, len(schedule.Schedules))
	for i, ps := range schedule.Schedules {
		policySchedule := map[string]interface{}{}

		// Map policy
		if ps.Policy != nil {
			policySchedule["policy"] = []map[string]interface{}{
				{
					"name": ps.Policy.Name,
					"slug": ps.Policy.Slug,
				},
			}
		} else {
			policySchedule["policy"] = []map[string]interface{}{}
		}

		// Map schedule entries
		entries := make([]map[string]interface{}, len(ps.Schedule))
		for j, entry := range ps.Schedule {
			entryMap := map[string]interface{}{
				"oncall_type":   entry.OnCallType,
				"rotation_name": entry.RotationName,
				"shift_name":    entry.ShiftName,
				"shift_roll":    entry.ShiftRoll,
			}

			// Map on-call user
			if entry.OnCallUser != nil {
				entryMap["oncall_user"] = entry.OnCallUser.Username
			} else {
				entryMap["oncall_user"] = ""
			}

			// Map override on-call user
			if entry.OverrideOnCallUser != nil {
				entryMap["override_oncall_user"] = entry.OverrideOnCallUser.Username
			} else {
				entryMap["override_oncall_user"] = ""
			}

			// Map rolls
			rolls := make([]map[string]interface{}, len(entry.Rolls))
			for k, roll := range entry.Rolls {
				rollMap := map[string]interface{}{
					"start":   roll.Start,
					"end":     roll.End,
					"is_roll": roll.IsRoll,
				}
				if roll.OnCallUser != nil {
					rollMap["oncall_user"] = roll.OnCallUser.Username
				} else {
					rollMap["oncall_user"] = ""
				}
				rolls[k] = rollMap
			}
			entryMap["rolls"] = rolls

			entries[j] = entryMap
		}
		policySchedule["schedule"] = entries

		// Map overrides
		overrides := make([]map[string]interface{}, len(ps.Overrides))
		for j, override := range ps.Overrides {
			overrideMap := map[string]interface{}{
				"start": override.Start,
				"end":   override.End,
			}
			if override.OrigOnCallUser != nil {
				overrideMap["orig_oncall_user"] = override.OrigOnCallUser.Username
			} else {
				overrideMap["orig_oncall_user"] = ""
			}
			if override.OverrideOnCallUser != nil {
				overrideMap["override_oncall_user"] = override.OverrideOnCallUser.Username
			} else {
				overrideMap["override_oncall_user"] = ""
			}
			overrides[j] = overrideMap
		}
		policySchedule["overrides"] = overrides

		schedulesList[i] = policySchedule
	}

	if err := d.Set("schedules", schedulesList); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
