package ccloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/sapcc/gophercloud-sapcc/billing/masterdata/projects"
)

func billingProjectFlattenCostObject(co projects.CostObject) []map[string]interface{} {
	return []map[string]interface{}{{
		"inherited": co.Inherited,
		"name":      co.Name,
		"type":      co.Type,
	}}
}

func billingProjectExpandCostObject(raw interface{}) projects.CostObject {
	var co projects.CostObject

	if raw != nil {
		if v, ok := raw.([]interface{}); ok {
			for _, v := range v {
				if v, ok := v.(map[string]interface{}); ok {
					if v, ok := v["inherited"]; ok {
						co.Inherited = v.(bool)
					}
					if !co.Inherited {
						if v, ok := v["name"]; ok {
							co.Name = v.(string)
						}
						if v, ok := v["type"]; ok {
							co.Type = v.(string)
						}
					}

					return co
				}
			}
		}
	}

	return co
}

// replaceEmpty is a helper function to replace empty fields with another field.
func replaceEmpty(d *schema.ResourceData, field string, b string) string {
	var v interface{}
	var ok bool
	if v, ok = d.GetOkExists(field); !ok {
		return b
	}
	return v.(string)
}
