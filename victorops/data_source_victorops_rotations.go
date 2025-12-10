package victorops

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRotations() *schema.Resource {
	return &schema.Resource{
		Description: "Get a team's rotations. Note: Rotations are read-only via the API.",
		ReadContext: dataSourceRotationsRead,

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The slug/ID of the team.",
			},
			"rotations": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The rotations for the team.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The label/name of the rotation.",
						},
						"total_members_in_rotation": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The total number of members in the rotation.",
						},
						"shifts": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The shifts in the rotation.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"label": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The label of the shift.",
									},
									"duration": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The duration of the shift in seconds.",
									},
									"shift_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The type of the shift (std, fts, cstm, pho).",
									},
									"start": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The start time of the shift (ISO8601 timestamp).",
									},
									"timezone": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The timezone for the shift.",
									},
									"shift_members": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The members of the shift.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"slug": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The unique identifier for the shift member.",
												},
												"username": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The username of the shift member.",
												},
											},
										},
									},
									"current": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The current on-call period.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"start": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Start time (ISO8601 timestamp).",
												},
												"end": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "End time (ISO8601 timestamp).",
												},
												"username": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Username of the on-call user.",
												},
											},
										},
									},
									"next": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The next on-call period.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"start": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Start time (ISO8601 timestamp).",
												},
												"end": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "End time (ISO8601 timestamp).",
												},
												"username": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Username of the on-call user.",
												},
											},
										},
									},
									"periods": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "List of on-call periods.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"start": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Start time (ISO8601 timestamp).",
												},
												"end": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "End time (ISO8601 timestamp).",
												},
												"username": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Username of the on-call user.",
												},
											},
										},
									},
									"mask": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The primary rotation mask defining days and time ranges.",
										Elem:        rotationMaskSchema(),
									},
									"mask2": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "Optional secondary rotation mask.",
										Elem:        rotationMaskSchema(),
									},
									"mask3": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "Optional tertiary rotation mask.",
										Elem:        rotationMaskSchema(),
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

// rotationMaskSchema returns the schema for a rotation mask
func rotationMaskSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"day": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Days of the week the rotation is active.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"su": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Sunday.",
						},
						"m": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Monday.",
						},
						"t": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Tuesday.",
						},
						"w": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Wednesday.",
						},
						"th": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Thursday.",
						},
						"f": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Friday.",
						},
						"sa": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Saturday.",
						},
					},
				},
			},
			"time": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Time ranges for the rotation.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Start time.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hour": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Hour (0-23).",
									},
									"minute": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Minute (0-59).",
									},
								},
							},
						},
						"end": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "End time.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hour": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Hour (0-23).",
									},
									"minute": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Minute (0-59).",
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

func dataSourceRotationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(Config)
	apiClient := NewAPIClient(config.BaseURL, config.APIId, config.APIKey)

	teamID := d.Get("team_id").(string)

	rotations, err := apiClient.GetTeamRotations(teamID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(teamID)

	rotationList := make([]map[string]interface{}, len(rotations))
	for i, r := range rotations {
		shifts := make([]map[string]interface{}, len(r.Shifts))
		for j, s := range r.Shifts {
			shift := map[string]interface{}{
				"label":      s.Label,
				"duration":   s.Duration,
				"shift_type": s.ShiftType,
				"start":      s.Start,
				"timezone":   s.Timezone,
			}

			// Map shift members
			shiftMembers := make([]map[string]interface{}, len(s.ShiftMembers))
			for k, sm := range s.ShiftMembers {
				shiftMembers[k] = map[string]interface{}{
					"slug":     sm.Slug,
					"username": sm.Username,
				}
			}
			shift["shift_members"] = shiftMembers

			// Map current on-call period
			if s.Current != nil {
				shift["current"] = []map[string]interface{}{
					{
						"start":    s.Current.Start,
						"end":      s.Current.End,
						"username": s.Current.Username,
					},
				}
			} else {
				shift["current"] = []map[string]interface{}{}
			}

			// Map next on-call period
			if s.Next != nil {
				shift["next"] = []map[string]interface{}{
					{
						"start":    s.Next.Start,
						"end":      s.Next.End,
						"username": s.Next.Username,
					},
				}
			} else {
				shift["next"] = []map[string]interface{}{}
			}

			// Map periods
			periods := make([]map[string]interface{}, len(s.Periods))
			for k, p := range s.Periods {
				periods[k] = map[string]interface{}{
					"start":    p.Start,
					"end":      p.End,
					"username": p.Username,
				}
			}
			shift["periods"] = periods

			// Map masks
			shift["mask"] = flattenRotationMask(s.Mask)
			shift["mask2"] = flattenRotationMask(s.Mask2)
			shift["mask3"] = flattenRotationMask(s.Mask3)

			shifts[j] = shift
		}
		rotationList[i] = map[string]interface{}{
			"label":                     r.Label,
			"total_members_in_rotation": r.TotalMembersInRotation,
			"shifts":                    shifts,
		}
	}

	if err := d.Set("rotations", rotationList); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// flattenRotationMask converts a RotationMask to a Terraform-compatible structure
func flattenRotationMask(mask *RotationMask) []map[string]interface{} {
	if mask == nil {
		return []map[string]interface{}{}
	}

	result := map[string]interface{}{}

	// Flatten day
	if mask.Day != nil {
		result["day"] = []map[string]interface{}{
			{
				"su": mask.Day.Su,
				"m":  mask.Day.M,
				"t":  mask.Day.T,
				"w":  mask.Day.W,
				"th": mask.Day.Th,
				"f":  mask.Day.F,
				"sa": mask.Day.Sa,
			},
		}
	} else {
		result["day"] = []map[string]interface{}{}
	}

	// Flatten time ranges
	timeRanges := make([]map[string]interface{}, len(mask.Time))
	for i, tr := range mask.Time {
		timeRange := map[string]interface{}{}

		if tr.Start != nil {
			timeRange["start"] = []map[string]interface{}{
				{
					"hour":   tr.Start.Hour,
					"minute": tr.Start.Minute,
				},
			}
		} else {
			timeRange["start"] = []map[string]interface{}{}
		}

		if tr.End != nil {
			timeRange["end"] = []map[string]interface{}{
				{
					"hour":   tr.End.Hour,
					"minute": tr.End.Minute,
				},
			}
		} else {
			timeRange["end"] = []map[string]interface{}{}
		}

		timeRanges[i] = timeRange
	}
	result["time"] = timeRanges

	return []map[string]interface{}{result}
}
