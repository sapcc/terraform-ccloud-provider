package ccloud

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"

	"github.com/ghodss/yaml"
	"github.com/sapcc/kubernikus/pkg/api/client/operations"
	"github.com/sapcc/kubernikus/pkg/api/models"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api/v1"
)

const (
	KlusterNameRegex = "^[a-z][-a-z0-9]{0,18}[a-z0-9]?$"
	PoolNameRegex    = "^[a-z][a-z0-9]{0,19}$"
)

func kubernikusValidateClusterName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if !regexp.MustCompile(KlusterNameRegex).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q must be 1 to 20 characters with lowercase and uppercase letters, numbers and hyphens.", k))
	}
	return
}

func kubernikusValidatePoolName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if !regexp.MustCompile(PoolNameRegex).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q must be 1 to 20 characters with lowercase and uppercase letters and numbers.", k))
	}
	return
}

func kubernikusFlattenOpenstackSpecV1(spec *models.OpenstackSpec) []map[string]interface{} {
	var res []map[string]interface{}

	if spec == (&models.OpenstackSpec{}) {
		return res
	}

	return append(res, map[string]interface{}{
		"lb_floating_network_id": spec.LBFloatingNetworkID,
		"lb_subnet_id":           spec.LBSubnetID,
		"network_id":             spec.NetworkID,
		"project_id":             spec.ProjectID,
		"router_id":              spec.RouterID,
		"security_group_name":    spec.SecurityGroupName,
	})
}

func kubernikusFlattenNodePoolsV1(nodePools []models.NodePool) []map[string]interface{} {
	var res []map[string]interface{}
	for _, p := range nodePools {
		res = append(res, map[string]interface{}{
			"availability_zone": p.AvailabilityZone,
			"flavor":            p.Flavor,
			"image":             p.Image,
			"name":              p.Name,
			"size":              p.Size,
			"taints":            p.Taints,
			"labels":            p.Labels,
			"config": []map[string]interface{}{
				{
					"allow_reboot":  p.Config.AllowReboot,
					"allow_replace": p.Config.AllowReplace,
				},
			},
		})
	}
	return res
}

func kubernikusExpandOpenstackSpecV1(raw interface{}) *models.OpenstackSpec {
	if raw != nil {
		if v, ok := raw.([]interface{}); ok {
			for _, v := range v {
				if v, ok := v.(map[string]interface{}); ok {
					res := new(models.OpenstackSpec)

					if v, ok := v["lb_floating_network_id"]; ok {
						res.LBFloatingNetworkID = v.(string)
					}
					if v, ok := v["lb_subnet_id"]; ok {
						res.LBSubnetID = v.(string)
					}
					if v, ok := v["network_id"]; ok {
						res.NetworkID = v.(string)
					}
					if v, ok := v["project_id"]; ok {
						res.ProjectID = v.(string)
					}
					if v, ok := v["router_id"]; ok {
						res.RouterID = v.(string)
					}
					if v, ok := v["security_group_name"]; ok {
						res.SecurityGroupName = v.(string)
					}

					return res
				}
			}
		}
	}

	return nil
}

func kubernikusExpandNodePoolsV1(raw interface{}) ([]models.NodePool, error) {
	var names []string

	if raw != nil {
		if v, ok := raw.([]interface{}); ok {
			var res []models.NodePool

			for _, v := range v {
				if v, ok := v.(map[string]interface{}); ok {
					var p models.NodePool

					if v, ok := v["name"]; ok {
						p.Name = v.(string)
						if strSliceContains(names, p.Name) {
							return nil, fmt.Errorf("Duplicate node pool name found: %s", p.Name)
						}
						names = append(names, p.Name)
					}
					if v, ok := v["flavor"]; ok {
						p.Flavor = v.(string)
					}
					if v, ok := v["image"]; ok {
						p.Image = v.(string)
					}
					if v, ok := v["size"]; ok {
						p.Size = int64(v.(int))
					}
					if v, ok := v["availability_zone"]; ok {
						p.AvailabilityZone = v.(string)
					}
					if v, ok := v["taints"]; ok {
						p.Taints = expandToStringSlice(v.([]interface{}))
					}
					if v, ok := v["labels"]; ok {
						p.Labels = expandToStringSlice(v.([]interface{}))
					}
					if v, ok := v["config"]; ok {
						p.Config = expandToNodePoolConfig(v.([]interface{}))
					}

					res = append(res, p)
				}
			}

			return res, nil
		}
	}

	return nil, nil
}

func kubernikusWaitForClusterV1(klient *Kubernikus, name string, target string, pending []string, timeout time.Duration) error {
	// Phase: "Pending","Creating","Running","Terminating","Upgrading"
	log.Printf("[DEBUG] Waiting for %s cluster to become %s.", name, target)

	stateConf := &resource.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    kubernikusKlusterV1GetPhase(klient, target, name),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForState()

	if err != nil {
		if e, ok := err.(*operations.ShowClusterDefault); ok && target == "Terminated" && e.Payload.Message == "Not found" {
			return nil
		}
	}

	return err
}

func kubernikusKlusterV1GetPhase(klient *Kubernikus, target string, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := klient.ShowCluster(operations.NewShowClusterParams().WithName(name), klient.authFunc())
		if err != nil {
			return nil, "", err
		}

		pretty, _ := json.MarshalIndent(result.Payload, "", "  ")
		log.Printf("[DEBUG] Payload state response: %s", string(pretty))

		if target != "Terminated" {
			events, err := klient.GetClusterEvents(operations.NewGetClusterEventsParams().WithName(name), klient.authFunc())
			if err != nil {
				return nil, "", err
			}

			if len(events.Payload) > 0 {
				// check, whether there are error events
				event := events.Payload[len(events.Payload)-1]
				pretty, _ = json.MarshalIndent(event, "", "  ")
				log.Printf("[DEBUG] Latest event response: %s", string(pretty))

				if strings.Contains(event.Reason, "Error") || strings.Contains(event.Reason, "Failed") {
					return nil, event.Reason, fmt.Errorf(event.Message)
				}
			}

			for _, a := range result.Payload.Spec.NodePools {
				// workaround for the upgrade status race condition
				if result.Payload.Status.Phase == models.KlusterPhaseRunning &&
					result.Payload.Spec.Version != result.Payload.Status.ApiserverVersion {
					return result.Payload, string(models.KlusterPhaseUpgrading), nil
				}

				for _, s := range result.Payload.Status.NodePools {
					if a.Name == s.Name {
						// sometimes status size doesn't reflect the actual size, therefore we use "a.Size"
						if a.Size != s.Healthy {
							return result.Payload, "Pending", nil
						}
					}
				}
			}

			if len(result.Payload.Spec.NodePools) != len(result.Payload.Status.NodePools) {
				return result.Payload, "Pending", nil
			}
		}

		return result.Payload, string(result.Payload.Status.Phase), nil
	}
}

func kubernikusUpdateNodePoolsV1(klient *Kubernikus, cluster *models.Kluster, oldNodePoolsRaw, newNodePoolsRaw interface{}, target string, pending []string, timeout time.Duration) error {
	var poolsToKeep []models.NodePool
	var poolsToDelete []models.NodePool
	oldNodePools, err := kubernikusExpandNodePoolsV1(oldNodePoolsRaw)
	if err != nil {
		return err
	}
	newNodePools, err := kubernikusExpandNodePoolsV1(newNodePoolsRaw)
	if err != nil {
		return err
	}

	pretty, _ := json.MarshalIndent(oldNodePools, "", "  ")
	log.Printf("[DEBUG] Old node pools: %s", string(pretty))
	pretty, _ = json.MarshalIndent(newNodePools, "", "  ")
	log.Printf("[DEBUG] New node pools: %s", string(pretty))

	// Determine if any node pools removed from the configuration.
	// Then downscale those pools and delete.
	for _, op := range oldNodePools {
		var found bool
		for _, np := range newNodePools {
			if op.Name == np.Name && op.Flavor == np.Flavor && op.Image == np.Image && (np.AvailabilityZone == "" || op.AvailabilityZone == np.AvailabilityZone) {
				tmp := np
				// copy previously "computed" AZ
				if np.AvailabilityZone == "" {
					tmp.AvailabilityZone = op.AvailabilityZone
				}
				poolsToKeep = append(poolsToKeep, tmp)
				found = true
			}
		}

		if !found {
			tmp := op
			tmp.Size = 0
			poolsToDelete = append(poolsToDelete, tmp)
		}
	}

	pretty, _ = json.MarshalIndent(poolsToKeep, "", "  ")
	log.Printf("[DEBUG] Keep node pools: %s", string(pretty))
	pretty, _ = json.MarshalIndent(poolsToDelete, "", "  ")
	log.Printf("[DEBUG] Downscale node pools: %s", string(pretty))

	if len(poolsToDelete) > 0 {
		// downscale
		cluster.Spec.NodePools = append(poolsToKeep, poolsToDelete...)
		err = kubernikusUpdateAndWait(klient, cluster, target, pending, timeout)
		if err != nil {
			return err
		}
	}

	// delete old
	cluster.Spec.NodePools = poolsToKeep
	err = kubernikusUpdateAndWait(klient, cluster, target, pending, timeout)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(poolsToKeep, newNodePools) {
		// create new
		cluster.Spec.NodePools = newNodePools
		err = kubernikusUpdateAndWait(klient, cluster, target, pending, timeout)
		if err != nil {
			return err
		}
	}

	return nil
}

func kubernikusHandleErrorV1(msg string, err error) error {
	switch err.(type) {
	case *operations.TerminateClusterDefault:
		result := err.(*operations.TerminateClusterDefault)
		return fmt.Errorf("%s: %s", msg, result.Payload.Message)
	case error:
		return fmt.Errorf("%s: %s", msg, err)
	}
	return err
}

func kubernikusUpdateAndWait(klient *Kubernikus, cluster *models.Kluster, target string, pending []string, timeout time.Duration) error {
	pretty, _ := json.MarshalIndent(cluster, "", "  ")
	log.Printf("[DEBUG] Payload request: %s", string(pretty))

	result, err := klient.UpdateCluster(operations.NewUpdateClusterParams().WithName(cluster.Name).WithBody(cluster), klient.authFunc())
	if err != nil {
		return kubernikusHandleErrorV1("Error updating cluster", err)
	}

	pretty, _ = json.MarshalIndent(result.Payload, "", "  ")
	log.Printf("[DEBUG] Payload response: %s", string(pretty))

	err = kubernikusWaitForClusterV1(klient, cluster.Name, target, pending, timeout)
	if err != nil {
		return kubernikusHandleErrorV1("Error waiting for cluster node pools Running state", err)
	}

	return nil
}

func getCredentials(klient *Kubernikus, name string, creds string) (string, []map[string]string, error) {
	var err error
	var kubeConfig []map[string]string
	var crt *x509.Certificate

	if creds == "" {
		creds, kubeConfig, err = downloadCredentials(klient, name)
		if err != nil {
			return "", nil, err
		}
	} else {
		kubeConfig, crt, err = flattenKubernetesClusterKubeConfig(creds)
		if err != nil {
			return "", nil, err
		}

		// Check so that the certificate is valid now
		now := time.Now()
		if now.Before(crt.NotBefore) || now.After(crt.NotAfter) {
			log.Printf("[DEBUG] The Kubernikus certificate is not valid")
			creds, kubeConfig, err = downloadCredentials(klient, name)
			if err != nil {
				return "", nil, err
			}
		}
	}

	return creds, kubeConfig, nil
}

func flattenKubernetesClusterKubeConfig(creds string) ([]map[string]string, *x509.Certificate, error) {
	var cfg clientcmdapi.Config
	var values = make(map[string]string)
	var crt *x509.Certificate

	err := yaml.Unmarshal([]byte(creds), &cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to unmarshal Kubernikus kubeconfig: %s", err)
	}

	for _, v := range cfg.Clusters {
		values["host"] = v.Cluster.Server
		values["cluster_ca_certificate"] = base64.StdEncoding.EncodeToString(v.Cluster.CertificateAuthorityData)
	}

	for _, v := range cfg.AuthInfos {
		values["username"] = v.Name
		values["client_certificate"] = base64.StdEncoding.EncodeToString(v.AuthInfo.ClientCertificateData)
		values["client_key"] = base64.StdEncoding.EncodeToString(v.AuthInfo.ClientKeyData)

		// parse certificate date
		pem, _ := pem.Decode(v.AuthInfo.ClientCertificateData)
		if pem == nil {
			return nil, nil, fmt.Errorf("Failed to decode PEM")
		}
		crt, err = x509.ParseCertificate(pem.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to parse Kubernikus certificate %s", err)
		}
		values["not_before"] = crt.NotBefore.Format(time.RFC3339)
		values["not_after"] = crt.NotAfter.Format(time.RFC3339)
	}

	if crt == nil {
		return nil, nil, fmt.Errorf("Failed to get Kubernikus kubeconfig credentials %s", err)
	}

	return []map[string]string{values}, crt, nil
}

func downloadCredentials(klient *Kubernikus, name string) (string, []map[string]string, error) {
	credentials, err := klient.GetClusterCredentials(operations.NewGetClusterCredentialsParams().WithName(name), klient.authFunc())
	if err != nil {
		return "", nil, fmt.Errorf("Failed to download Kubernikus kubeconfig: %s", err)
	}

	kubeConfig, _, err := flattenKubernetesClusterKubeConfig(credentials.Payload.Kubeconfig)
	if err != nil {
		return "", nil, err
	}

	return credentials.Payload.Kubeconfig, kubeConfig, nil
}