package ccloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/sapcc/gophercloud-billing/billing/projects"
)

func billingFlattenCostObject(co projects.CostObject) []map[string]interface{} {
	var res []map[string]interface{}

	return append(res, map[string]interface{}{
		"inherited": co.Inherited,
		"name":      co.Name,
		"type":      co.Type,
	})
}

func billingExpandCostObject(raw interface{}) projects.CostObject {
	co := projects.CostObject{}

	if raw != nil {
		if v, ok := raw.([]interface{}); ok {
			for _, v := range v {
				if v, ok := v.(map[string]interface{}); ok {
					if v, ok := v["inherited"]; ok {
						co.Inherited = v.(bool)
					}
					if v, ok := v["name"]; ok {
						co.Name = v.(string)
					}
					if v, ok := v["type"]; ok {
						co.Type = v.(string)
					}

					return co
				}
			}
		}
	}

	return co
}

// replaceEmpty is a helper function to replace empty fields with another field
func replaceEmpty(d *schema.ResourceData, field string, b string) string {
	var v interface{}
	var ok bool
	if v, ok = d.GetOkExists(field); !ok {
		return b
	}
	return v.(string)
}
