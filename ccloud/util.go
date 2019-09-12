package ccloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/go-openapi/validate"
	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/structure"
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

// List of headers that contain sensitive data.
var sensitiveHeaders = map[string]struct{}{
	"x-auth-token":                    {},
	"x-auth-key":                      {},
	"x-service-token":                 {},
	"x-storage-token":                 {},
	"x-account-meta-temp-url-key":     {},
	"x-account-meta-temp-url-key-2":   {},
	"x-container-meta-temp-url-key":   {},
	"x-container-meta-temp-url-key-2": {},
	"set-cookie":                      {},
	"x-subject-token":                 {},
}

func hideSensitiveHeadersData(headers http.Header) []string {
	result := make([]string, len(headers))
	headerIdx := 0
	for header, data := range headers {
		if _, ok := sensitiveHeaders[strings.ToLower(header)]; ok {
			result[headerIdx] = fmt.Sprintf("%s: %s", header, "***")
		} else {
			result[headerIdx] = fmt.Sprintf("%s: %s", header, strings.Join(data, " "))
		}
		headerIdx++
	}

	return result
}

// formatHeaders converts standard http.Header type to a string with separated headers.
// It will hide data of sensitive headers.
func formatHeaders(headers http.Header, separator string) string {
	redactedHeaders := hideSensitiveHeadersData(headers)
	sort.Strings(redactedHeaders)

	return strings.Join(redactedHeaders, separator)
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

func normalizeJsonString(v interface{}) string {
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

func validateJsonObject(v interface{}, k string) ([]string, []error) {
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

func validateJsonArray(v interface{}, k string) ([]string, []error) {
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

func diffSuppressJsonObject(k, old, new string, d *schema.ResourceData) bool {
	if strSliceContains([]string{"{}", "null", ""}, old) &&
		strSliceContains([]string{"{}", "null", ""}, new) {
		return true
	}
	return false
}

func diffSuppressJsonArray(k, old, new string, d *schema.ResourceData) bool {
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
