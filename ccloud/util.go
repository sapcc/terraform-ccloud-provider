package ccloud

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-openapi/validate"
	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/sapcc/kubernikus/pkg/api/models"
)

// CheckDeleted checks the error to see if it's a 404 (Not Found) and, if so,
// sets the resource ID to the empty string instead of throwing an error.
func CheckDeleted(d *schema.ResourceData, err error, msg string) error {
	if _, ok := err.(gophercloud.ErrDefault404); ok {
		d.SetId("")
		return nil
	}

	return fmt.Errorf("%s %s: %s", msg, d.Id(), err)
}

// GetRegion returns the region that was specified in the resource. If a
// region was not set, the provider-level region is checked. The provider-level
// region can either be set by the region argument or by OS_REGION_NAME.
func GetRegion(d *schema.ResourceData, config *Config) string {
	if v, ok := d.GetOk("region"); ok {
		return v.(string)
	}

	return config.Region
}

// IsSliceContainsStr returns true if the string exists in given slice, ignore case.
func strSliceContains(sl []string, str string) bool {
	str = strings.ToLower(str)
	for _, s := range sl {
		if strings.ToLower(s) == str {
			return true
		}
	}
	return false
}

func expandToMapStringString(v map[string]interface{}) map[string]string {
	m := make(map[string]string)
	for key, val := range v {
		if strVal, ok := val.(string); ok {
			m[key] = strVal
		}
		if strVal, ok := val.(bool); ok {
			m[key] = fmt.Sprintf("%t", strVal)
		}
	}

	return m
}

func expandToStringSlice(v []interface{}) []string {
	s := make([]string, len(v))
	for i, val := range v {
		if strVal, ok := val.(string); ok {
			s[i] = strVal
		}
	}

	return s
}

func expandToNodePoolConfig(v []interface{}) *models.NodePoolConfig {
	c := new(models.NodePoolConfig)
	for _, val := range v {
		if mapVal, ok := val.(map[string]interface{}); ok {
			if v, ok := mapVal["allow_reboot"].(bool); ok {
				c.AllowReboot = &v
			}
			if v, ok := mapVal["allow_replace"].(bool); ok {
				c.AllowReplace = &v
			}
		}
	}

	return c
}

func normalizeJSONString(v interface{}) string {
	json, _ := structure.NormalizeJsonString(v)
	return json
}

func validateURL(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := url.ParseRequestURI(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q URL is not valid: %s", k, err))
	}
	return
}

func validateJSONObject(v interface{}, k string) ([]string, []error) {
	if v == nil || v.(string) == "" {
		return nil, []error{fmt.Errorf("%q value must not be empty", k)}
	}

	var j map[string]interface{}
	s := v.(string)

	err := json.Unmarshal([]byte(s), &j)
	if err != nil {
		return nil, []error{fmt.Errorf("%q must be a JSON object: %s", k, err)}
	}

	return nil, nil
}

func validateJSONArray(v interface{}, k string) ([]string, []error) {
	if v == nil || v.(string) == "" {
		return nil, []error{fmt.Errorf("%q value must not be empty", k)}
	}

	var j []interface{}
	s := v.(string)

	err := json.Unmarshal([]byte(s), &j)
	if err != nil {
		return nil, []error{fmt.Errorf("%q must be a JSON array: %s", k, err)}
	}

	return nil, nil
}

func validateTimeout(v interface{}, k string) ([]string, []error) {
	if v == nil || v.(string) == "" {
		return nil, []error{fmt.Errorf("%q value must not be empty", k)}
	}

	_, err := time.ParseDuration(v.(string))
	if err != nil {
		return nil, []error{fmt.Errorf("%q: %s", k, err)}
	}

	return nil, nil
}

func diffSuppressJSONObject(k, old, new string, d *schema.ResourceData) bool {
	if strSliceContains([]string{"{}", "null", ""}, old) &&
		strSliceContains([]string{"{}", "null", ""}, new) {
		return true
	}
	return false
}

func diffSuppressJSONArray(k, old, new string, d *schema.ResourceData) bool {
	if strSliceContains([]string{"[]", "null", ""}, old) &&
		strSliceContains([]string{"[]", "null", ""}, new) {
		return true
	}
	return false
}

func validateKubernetesVersion(v interface{}, k string) ([]string, []error) {
	if err := validate.Pattern("version", "", v.(string), `^[0-9]+\.[0-9]+\.[0-9]+$`); err != nil {
		return nil, []error{fmt.Errorf("Invalid version (%s) specified for Kubernikus cluster: %s", v.(string), err)}
	}

	return nil, nil
}
