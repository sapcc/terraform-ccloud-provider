package ccloud

import (
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/sapcc/kubernikus/pkg/api/client/operations"
	"github.com/sapcc/kubernikus/pkg/api/models"
)

const (
	KlusterNameRegex = "[a-z][-a-z0-9]{0,18}[a-z0-9]?"
	PoolNameRegex    = "[a-z][a-z0-9]{0,19}"
)

func resourceCCloudKubernetes() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,

		Read:   resourceCCloudKubernetesRead,
		Update: resourceCCloudKubernetesUpdate,
		Create: resourceCCloudKubernetesCreate,
		Delete: resourceCCloudKubernetesDelete,

		Schema: map[string]*schema.Schema{
			"is_admin": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateKlusterName(),
			},

			"advertise_address": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "1.1.1.1",
				ValidateFunc: validation.SingleIP(),
			},

			"cluster_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "10.100.0.0/16",
				ValidateFunc: validation.CIDRNetwork(8, 16),
			},

			"service_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "198.18.128.0/17",
				ValidateFunc: validation.CIDRNetwork(8, 24),
			},

			"dns_address": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.SingleIP(),
			},

			"dns_domain": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "cluster.local",
			},

			"ssh_public_key": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"node_pools": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validatePoolName(),
						},
						"flavor": {
							Type:     schema.TypeString,
							Required: true,
						},
						"image": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "coreos-stable-amd64",
						},
						"size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntAtMost(127),
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"taints": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"labels": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"openstack": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lb_floating_network_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"lb_subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"network_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"router_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"security_group_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceCCloudKubernetesRead(d *schema.ResourceData, meta interface{}) error {
	return get(d, meta, true)
}

func get(d *schema.ResourceData, meta interface{}, updateState bool) error {
	config := meta.(*Config)
	log.Printf("[KUBERNETES] Reading Kubernikus Kluster in project %s", config.TenantID)

	kubernikus, err := config.kubernikusV1Client(GetRegion(d, config), d.Get("is_admin").(bool))
	if err != nil {
		return fmt.Errorf("Error creating Kubernikus client: %s", err)
	}

	result, err := kubernikus.ShowCluster(operations.NewShowClusterParams().WithName(d.Get("name").(string)), kubernikus.authFunc())
	switch err.(type) {
	case *operations.ShowClusterDefault:
		result := err.(*operations.ShowClusterDefault)

		if result.Payload.Message == "Not found" {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error reading Kubernikus cluster: %s", result.Payload.Message)
	case error:
		return fmt.Errorf("Error reading Kubernikus cluster: %s", err)
	}

	if updateState {
		updateStateFromAPIResponse(d, result.Payload)
	}

	return nil
}

func resourceCCloudKubernetesCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	log.Printf("[KUBERNETES] Creating Kubernikus Kluster in project %s", config.TenantID)

	kubernikus, err := config.kubernikusV1Client(GetRegion(d, config), d.Get("is_admin").(bool))
	if err != nil {
		return fmt.Errorf("Error creating Kubernikus client: %s", err)
	}

	cluster := &models.Kluster{
		Spec: models.KlusterSpec{
			NodePools: []models.NodePool{},
			Openstack: models.OpenstackSpec{},
		},
	}

	if name, ok := d.GetOk("name"); ok {
		cluster.Name = name.(string)
	}
	if advertise_address, ok := d.GetOk("advertise_address"); ok {
		cluster.Spec.AdvertiseAddress = advertise_address.(string)
	}
	if cluster_cidr, ok := d.GetOk("cluster_cidr"); ok {
		cluster.Spec.ClusterCIDR = cluster_cidr.(string)
	}
	if dns_address, ok := d.GetOk("dns_address"); ok {
		cluster.Spec.DNSAddress = dns_address.(string)
	}
	if dns_domain, ok := d.GetOk("dns_domain"); ok {
		cluster.Spec.DNSDomain = dns_domain.(string)
	}
	if ssh_public_key, ok := d.GetOk("ssh_public_key"); ok {
		cluster.Spec.SSHPublicKey = ssh_public_key.(string)
	}
	if service_cidr, ok := d.GetOk("service_cidr"); ok {
		cluster.Spec.ServiceCIDR = service_cidr.(string)
	}

	if v, ok := d.GetOk("node_pools"); ok {
		nodePools := v.([]interface{})
		for i := range nodePools {
			newPool := models.NodePool{}

			if name, ok := d.GetOk(fmt.Sprintf("node_pools.%d.name", i)); ok {
				newPool.Name = name.(string)
			}
			if flavor, ok := d.GetOk(fmt.Sprintf("node_pools.%d.flavor", i)); ok {
				newPool.Flavor = flavor.(string)
			}
			if image, ok := d.GetOk(fmt.Sprintf("node_pools.%d.image", i)); ok {
				newPool.Image = image.(string)
			}
			if size, ok := d.GetOk(fmt.Sprintf("node_pools.%d.size", i)); ok {
				newPool.Size = int64(size.(int))
			}
			if avz, ok := d.GetOk(fmt.Sprintf("node_pools.%d.availability_zone", i)); ok {
				newPool.AvailabilityZone = avz.(string)
			}
			if taints, ok := d.GetOk(fmt.Sprintf("node_pools.%d.taints", i)); ok {
				t := taints.([]interface{})
				for j := range t {
					if taint, ok := d.GetOk(fmt.Sprintf("node_pools.%d.taints.%d", i, j)); ok {
						newPool.Taints = append(newPool.Taints, taint.(string))
					}
				}
			}
			if labels, ok := d.GetOk(fmt.Sprintf("node_pools.%d.labels", i)); ok {
				l := labels.([]interface{})
				for j := range l {
					if label, ok := d.GetOk(fmt.Sprintf("node_pools.%d.labels.%d", i, j)); ok {
						newPool.Labels = append(newPool.Labels, label.(string))
					}
				}
			}
			cluster.Spec.NodePools = append(cluster.Spec.NodePools, newPool)
		}
	}

	if lb_floating_network_id, ok := d.GetOk("openstack.0.lb_floating_network_id"); ok {
		cluster.Spec.Openstack.LBFloatingNetworkID = lb_floating_network_id.(string)
	}

	if lb_subnet_id, ok := d.GetOk("openstack.0.lb_subnet_id"); ok {
		cluster.Spec.Openstack.LBSubnetID = lb_subnet_id.(string)
	}

	if network_id, ok := d.GetOk("openstack.0.network_id"); ok {
		cluster.Spec.Openstack.NetworkID = network_id.(string)
	}

	if router_id, ok := d.GetOk("openstack.0.router_id"); ok {
		cluster.Spec.Openstack.RouterID = router_id.(string)
	}

	if security_group_name, ok := d.GetOk("openstack.0.security_group_name"); ok {
		cluster.Spec.Openstack.SecurityGroupName = security_group_name.(string)
	}

	result, err := kubernikus.CreateCluster(operations.NewCreateClusterParams().WithBody(cluster), kubernikus.authFunc())
	switch err.(type) {
	case *operations.CreateClusterDefault:
		result := err.(*operations.CreateClusterDefault)
		return fmt.Errorf("Error creating cluster: %s", result.Payload.Message)
	case error:
		return fmt.Errorf("Error creating cluster: %s", err)
	}

	d.SetId(result.Payload.Name)
	updateStateFromAPIResponse(d, result.Payload)

	return nil
}

func resourceCCloudKubernetesUpdate(d *schema.ResourceData, meta interface{}) error {
	err := get(d, meta, false)
	if err != nil {
		return err
	}

	config := meta.(*Config)
	log.Printf("[KUBERNETES] Updating Kubernikus Kluster in project %s", config.TenantID)

	kubernikus, err := config.kubernikusV1Client(GetRegion(d, config), d.Get("is_admin").(bool))
	if err != nil {
		return fmt.Errorf("Error creating Kubernikus client: %s", err)
	}

	cluster := &models.Kluster{
		Spec: models.KlusterSpec{
			NodePools: []models.NodePool{},
			Openstack: models.OpenstackSpec{},
		},
	}

	if name, ok := d.GetOk("name"); ok {
		cluster.Name = name.(string)
	}
	if ssh_public_key, ok := d.GetOk("ssh_public_key"); ok {
		cluster.Spec.SSHPublicKey = ssh_public_key.(string)
	}

	if v, ok := d.GetOk("node_pools"); ok {
		nodePools := v.([]interface{})
		for i := range nodePools {
			newPool := models.NodePool{}

			if name, ok := d.GetOk(fmt.Sprintf("node_pools.%d.name", i)); ok {
				newPool.Name = name.(string)
			}
			if flavor, ok := d.GetOk(fmt.Sprintf("node_pools.%d.flavor", i)); ok {
				newPool.Flavor = flavor.(string)
			}
			if image, ok := d.GetOk(fmt.Sprintf("node_pools.%d.image", i)); ok {
				newPool.Image = image.(string)
			}
			if size, ok := d.GetOk(fmt.Sprintf("node_pools.%d.size", i)); ok {
				newPool.Size = int64(size.(int))
			}
			if avz, ok := d.GetOk(fmt.Sprintf("node_pools.%d.availability_zone", i)); ok {
				newPool.AvailabilityZone = avz.(string)
			}
			if taints, ok := d.GetOk(fmt.Sprintf("node_pools.%d.taints", i)); ok {
				t := taints.([]interface{})
				for j := range t {
					if taint, ok := d.GetOk(fmt.Sprintf("node_pools.%d.taints.%d", i, j)); ok {
						newPool.Taints = append(newPool.Taints, taint.(string))
					}
				}
			}
			if labels, ok := d.GetOk(fmt.Sprintf("node_pools.%d.labels", i)); ok {
				l := labels.([]interface{})
				for j := range l {
					if label, ok := d.GetOk(fmt.Sprintf("node_pools.%d.labels.%d", i, j)); ok {
						newPool.Labels = append(newPool.Labels, label.(string))
					}
				}
			}
			cluster.Spec.NodePools = append(cluster.Spec.NodePools, newPool)
		}
	}

	if security_group_name, ok := d.GetOk("openstack.0.security_group_name"); ok {
		cluster.Spec.Openstack.SecurityGroupName = security_group_name.(string)
	}

	result, err := kubernikus.UpdateCluster(operations.NewUpdateClusterParams().WithName(d.Get("name").(string)).WithBody(cluster), kubernikus.authFunc())
	switch err.(type) {
	case *operations.UpdateClusterDefault:
		result := err.(*operations.UpdateClusterDefault)
		return fmt.Errorf("Error updating cluster: %s", result.Payload.Message)
	case error:
		return fmt.Errorf("Error updating cluster: %s", err)
	}

	updateStateFromAPIResponse(d, result.Payload)

	return nil
}

func resourceCCloudKubernetesDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	log.Printf("[KUBERNETES] Deleting Kubernikus Kluster in project %s", config.TenantID)

	kubernikus, err := config.kubernikusV1Client(GetRegion(d, config), d.Get("is_admin").(bool))
	if err != nil {
		return fmt.Errorf("Error creating Kubernikus client: %s", err)
	}

	_, err = kubernikus.TerminateCluster(operations.NewTerminateClusterParams().WithName(d.Get("name").(string)), kubernikus.authFunc())
	switch err.(type) {
	case *operations.TerminateClusterDefault:
		result := err.(*operations.TerminateClusterDefault)
		return fmt.Errorf("Error deleting cluster: %s", result.Payload.Message)
	case error:
		return fmt.Errorf("Error deleting cluster: %s", err)
	}

	d.SetId("")
	return nil
}

func validateKlusterName() schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		value := v.(string)

		if !regexp.MustCompile(KlusterNameRegex).MatchString(value) {
			errors = append(errors, fmt.Errorf(
				"%q name must be 1 to 20 characters with lowercase and uppercase letters, numbers and hyphens.", value))
		}
		return
	}
}

func validatePoolName() schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		value := v.(string)

		if !regexp.MustCompile(PoolNameRegex).MatchString(value) {
			errors = append(errors, fmt.Errorf(
				"%q name must be 1 to 20 characters with lowercase and uppercase letters, numbers.", value))
		}
		return
	}
}

func updateStateFromAPIResponse(d *schema.ResourceData, kluster *models.Kluster) {
	d.Set("advertise_address", kluster.Spec.AdvertiseAddress)
	d.Set("cluster_cidr", kluster.Spec.ClusterCIDR)
	d.Set("dns_address", kluster.Spec.DNSAddress)
	d.Set("dns_domain", kluster.Spec.DNSDomain)
	d.Set("name", kluster.Spec.Name)
	d.Set("ssh_public_key", kluster.Spec.SSHPublicKey)
	d.Set("service_cidr", kluster.Spec.ServiceCIDR)
	d.Set("version", kluster.Spec.Version)
	d.Set("openstack.0.lb_floating_network_id", kluster.Spec.Openstack.LBFloatingNetworkID)
	d.Set("openstack.0.lb_subnet_id", kluster.Spec.Openstack.LBSubnetID)
	d.Set("openstack.0.network_id", kluster.Spec.Openstack.NetworkID)
	d.Set("openstack.0.project_id", kluster.Spec.Openstack.ProjectID)
	d.Set("openstack.0.router_id", kluster.Spec.Openstack.RouterID)
	d.Set("openstack.0.security_group_name", kluster.Spec.Openstack.SecurityGroupName)

	nodePools := make([]map[string]interface{}, 0, 1)
	for _, p := range kluster.Spec.NodePools {
		nodePools = append(nodePools, map[string]interface{}{
			"availabilityZone": p.AvailabilityZone,
			"flavor":           p.Flavor,
			"image":            p.Image,
			"name":             p.Name,
			"size":             p.Size,
			"taints":           p.Taints,
			"labels":           p.Labels,
		})
	}

	d.Set("node_pools", nodePools)
}
