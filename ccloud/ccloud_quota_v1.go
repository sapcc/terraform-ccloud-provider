package ccloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/gophercloud/gophercloud"
	"github.com/sapcc/gophercloud-limes/resources/v1/projects"
	"github.com/sapcc/limes"
)

var (
	SERVICES = map[string]map[string]limes.Unit{
		"database": {
			"cfm_share_capacity": limes.UnitBytes,
		},
		"compute": {
			"cores":     limes.UnitNone,
			"instances": limes.UnitNone,
			"ram":       limes.UnitMebibytes,
		},
		"volumev2": {
			"capacity":  limes.UnitGibibytes,
			"snapshots": limes.UnitNone,
			"volumes":   limes.UnitNone,
		},
		"network": {
			"floating_ips":         limes.UnitNone,
			"networks":             limes.UnitNone,
			"ports":                limes.UnitNone,
			"rbac_policies":        limes.UnitNone,
			"routers":              limes.UnitNone,
			"security_group_rules": limes.UnitNone,
			"security_groups":      limes.UnitNone,
			"subnet_pools":         limes.UnitNone,
			"subnets":              limes.UnitNone,
			"healthmonitors":       limes.UnitNone,
			"l7policies":           limes.UnitNone,
			"listeners":            limes.UnitNone,
			"loadbalancers":        limes.UnitNone,
			"pools":                limes.UnitNone,
			"pool_members":         limes.UnitNone,
		},
		"dns": {
			"zones":      limes.UnitNone,
			"recordsets": limes.UnitNone,
		},
		"sharev2": {
			"share_networks":    limes.UnitNone,
			"share_capacity":    limes.UnitGibibytes,
			"shares":            limes.UnitNone,
			"snapshot_capacity": limes.UnitGibibytes,
			"share_snapshots":   limes.UnitNone,
		},
		"object-store": {
			"capacity": limes.UnitBytes,
		},
	}
)

func toString(r interface{}) string {
	switch v := r.(type) {
	case *limes.ProjectResourceReport:
		return limes.ValueWithUnit{v.Quota, v.Unit}.String()
	case *limes.DomainResourceReport:
		return limes.ValueWithUnit{v.DomainQuota, v.Unit}.String()
	}
	return ""
}

func sanitize(s string) string {
	return strings.Replace(s, "-", "", -1)
}

func limesCCloudProjectQuotaV1WaitForProject(client *gophercloud.ServiceClient, domainID string, projectID string, services *limes.QuotaRequest, timeout time.Duration) error {
	var msg string
	var err error

	// This condition is required, otherwise zero timeout will always raise:
	// "timeout while waiting for state to become 'active'"
	if timeout > 0 {
		// Retryable case, when timeout is set
		waitForAgent := &resource.StateChangeConf{
			Target:         []string{"active"},
			Refresh:        limesCCloudProjectQuotaV1GetQuota(client, domainID, projectID, services, timeout),
			Timeout:        timeout,
			Delay:          1 * time.Second,
			MinTimeout:     1 * time.Second,
			NotFoundChecks: 1000, // workaround for default 20 retries, when the resource is nil
		}
		_, err = waitForAgent.WaitForState()
	} else {
		// When timeout is not set, just get the agent
		_, msg, err = limesCCloudProjectQuotaV1GetQuota(client, domainID, projectID, services, timeout)()
	}

	if len(msg) > 0 && msg != "active" {
		return fmt.Errorf(msg)
	}

	if err != nil {
		return err
	}

	return nil
}

func limesCCloudProjectQuotaV1GetQuota(client *gophercloud.ServiceClient, domainID string, projectID string, services *limes.QuotaRequest, timeout time.Duration) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		quota, err := projects.Get(client, domainID, projectID, projects.GetOpts{}).Extract()
		if err != nil {
			if _, ok := err.(gophercloud.ErrDefault404); ok && timeout > 0 {
				// Retryable case, when timeout is set
				return nil, fmt.Sprintf("Unable to retrieve %s/%s ccloud_project_quota_v1: %s", domainID, projectID, err), nil
			}
			return nil, "", fmt.Errorf("Unable to retrieve %s/%s ccloud_project_quota_v1: %s", domainID, projectID, err)
		}

		// detect whether the quota is fully initialized before processing
		// otherwise further PUT will return "no project report for resource" 500 error
		for k, service := range quota.Services {
			if _, ok := (*services)[k]; ok && len(service.Resources) == 0 && timeout > 0 {
				// Retryable case, when timeout is set
				return nil, fmt.Sprintf("There are empty resources: %v", service.Resources), nil
			}
		}

		log.Printf("[DEBUG] Retrieved ccloud_project_quota_v1 %s: %+v", projectID, *quota)

		return quota, "active", nil
	}
}
