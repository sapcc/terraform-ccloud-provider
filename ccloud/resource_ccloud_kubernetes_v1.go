package ccloud

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/sapcc/kubernikus/pkg/api/client/operations"
	"github.com/sapcc/kubernikus/pkg/api/models"
)

func resourceCCloudKubernetesV1() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,

		Read:   resourceCCloudKubernetesV1Read,
		Update: resourceCCloudKubernetesV1Update,
		Create: resourceCCloudKubernetesV1Create,
		Delete: resourceCCloudKubernetesV1Delete,
		Importer: &schema.ResourceImporter{
			State: resourceCCloudKubernetesV1Import,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: kubernikusValidateClusterName,
			},

			"is_admin": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},

			"advertise_address": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.SingleIP(),
			},

			"cluster_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.CIDRNetwork(8, 16),
			},

			"service_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
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
				Computed: true,
			},

			"ssh_public_key": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"no_cloud": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},

			"version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateKubernetesVersion,
			},

			"node_pools": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: kubernikusValidatePoolName,
						},
						"flavor": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"image": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 127),
						},
						"availability_zone": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.NoZeroValues,
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
						"config": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allow_reboot": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
									"allow_replace": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			"openstack": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lb_floating_network_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"network_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"lb_subnet_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"project_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"router_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"security_group_name": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},

			"phase": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"wormhole": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"kube_config": {
				Type:      schema.TypeList,
				Computed:  true,
				Sensitive: true,
				MaxItems:  1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_key": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"cluster_ca_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"not_before": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"not_after": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"kube_config_raw": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceCCloudKubernetesV1Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	log.Printf("[KUBERNETES] Creating Kubernikus Kluster in project %s", config.TenantID)

	klient, err := config.kubernikusV1Client(GetRegion(d, config), d.Get("is_admin").(bool))
	if err != nil {
		return fmt.Errorf("Error creating Kubernikus client: %s", err)
	}

	cluster := &models.Kluster{
		Spec: models.KlusterSpec{
			NodePools: []models.NodePool{},
			Openstack: models.OpenstackSpec{},
		},
	}

	cluster.Name = d.Get("name").(string)
	cluster.Spec.AdvertiseAddress = d.Get("advertise_address").(string)
	cluster.Spec.ClusterCIDR = d.Get("cluster_cidr").(string)
	cluster.Spec.DNSAddress = d.Get("dns_address").(string)
	cluster.Spec.DNSDomain = d.Get("dns_domain").(string)
	cluster.Spec.SSHPublicKey = d.Get("ssh_public_key").(string)
	cluster.Spec.NoCloud = d.Get("no_cloud").(bool)
	cluster.Spec.ServiceCIDR = d.Get("service_cidr").(string)
	cluster.Spec.Version = d.Get("version").(string)
	err = verifySupportedKubernetesVersion(klient, cluster.Spec.Version)
	if err != nil {
		return err
	}
	cluster.Spec.NodePools, err = kubernikusExpandNodePoolsV1(d.Get("node_pools"))
	if err != nil {
		return err
	}
	if v := kubernikusExpandOpenstackSpecV1(d.Get("openstack")); v != nil {
		cluster.Spec.Openstack = *v
	}

	_, err = klient.CreateCluster(operations.NewCreateClusterParams().WithBody(cluster), klient.authFunc())
	if err != nil {
		return kubernikusHandleErrorV1("Error creating cluster", err)
	}

	d.SetId(cluster.Name)

	// waiting for Running state
	timeout := d.Timeout(schema.TimeoutCreate)
	target := "Running"
	pending := []string{"Pending", "Creating", "Upgrading"}
	err = kubernikusWaitForClusterV1(klient, cluster.Name, target, pending, timeout)
	if err != nil {
		return kubernikusHandleErrorV1("Error waiting for running cluster state", err)
	}

	return resourceCCloudKubernetesV1Read(d, meta)
}

func resourceCCloudKubernetesV1Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	log.Printf("[KUBERNETES] Reading Kubernikus Kluster in project %s", config.TenantID)

	klient, err := config.kubernikusV1Client(GetRegion(d, config), d.Get("is_admin").(bool))
	if err != nil {
		return fmt.Errorf("Error creating Kubernikus client: %s", err)
	}

	result, err := klient.ShowCluster(operations.NewShowClusterParams().WithName(d.Id()), klient.authFunc())
	if err != nil {
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
		return err
	}

	d.Set("advertise_address", result.Payload.Spec.AdvertiseAddress)
	d.Set("cluster_cidr", result.Payload.Spec.ClusterCIDR)
	d.Set("dns_address", result.Payload.Spec.DNSAddress)
	d.Set("dns_domain", result.Payload.Spec.DNSDomain)
	d.Set("name", result.Payload.Spec.Name)
	d.Set("ssh_public_key", result.Payload.Spec.SSHPublicKey)
	d.Set("no_cloud", result.Payload.Spec.NoCloud)
	d.Set("service_cidr", result.Payload.Spec.ServiceCIDR)
	d.Set("version", result.Payload.Spec.Version)
	d.Set("phase", result.Payload.Status.Phase)
	d.Set("wormhole", result.Payload.Status.Wormhole)
	d.Set("openstack", kubernikusFlattenOpenstackSpecV1(&result.Payload.Spec.Openstack))
	d.Set("node_pools", kubernikusFlattenNodePoolsV1(result.Payload.Spec.NodePools))

	d.Set("region", GetRegion(d, config))

	// if cluster is in pending state, than there are no credentials yet
	if result.Payload.Status.Phase != models.KlusterPhasePending {
		kubeConfigRaw, kubeConfig, err := getCredentials(klient, d.Id(), d.Get("kube_config_raw").(string))
		if err != nil {
			return err
		}
		d.Set("kube_config", kubeConfig)
		d.Set("kube_config_raw", kubeConfigRaw)
	}

	return nil
}

func resourceCCloudKubernetesV1Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	log.Printf("[KUBERNETES] Updating Kubernikus Kluster in project %s", config.TenantID)

	klient, err := config.kubernikusV1Client(GetRegion(d, config), d.Get("is_admin").(bool))
	if err != nil {
		return fmt.Errorf("Error creating Kubernikus client: %s", err)
	}

	timeout := d.Timeout(schema.TimeoutUpdate)
	cluster := &models.Kluster{
		Spec: models.KlusterSpec{
			NodePools: []models.NodePool{},
			Openstack: models.OpenstackSpec{},
		},
	}

	cluster.Name = d.Id()

	if v, ok := d.GetOk("ssh_public_key"); ok {
		cluster.Spec.SSHPublicKey = v.(string)
	}

	if v, ok := d.GetOk("version"); ok {
		cluster.Spec.Version = v.(string)
		err = verifySupportedKubernetesVersion(klient, cluster.Spec.Version)
		if err != nil {
			return err
		}
	}

	if v, ok := d.GetOk("openstack.0.security_group_name"); ok {
		cluster.Spec.Openstack.SecurityGroupName = v.(string)
	}

	o, n := d.GetChange("node_pools")

	// wait for the cluster to be upgraded, when new API version was specified
	target := "Running"
	pending := []string{"Pending", "Creating", "Terminating", "Upgrading"}
	err = kubernikusUpdateNodePoolsV1(klient, cluster, o, n, target, pending, timeout)
	if err != nil {
		return kubernikusHandleErrorV1("Error waiting for cluster to be updated", err)
	}

	return resourceCCloudKubernetesV1Read(d, meta)
}

func resourceCCloudKubernetesV1Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	log.Printf("[KUBERNETES] Deleting Kubernikus Kluster in project %s", config.TenantID)

	klient, err := config.kubernikusV1Client(GetRegion(d, config), d.Get("is_admin").(bool))
	if err != nil {
		return fmt.Errorf("Error creating Kubernikus client: %s", err)
	}

	timeout := d.Timeout(schema.TimeoutDelete)

	_, err = klient.TerminateCluster(operations.NewTerminateClusterParams().WithName(d.Id()), klient.authFunc())
	if err != nil {
		return kubernikusHandleErrorV1("Error deleting cluster", err)
	}

	target := "Terminated"
	pending := []string{"Pending", "Creating", "Running", "Terminating", "Upgrading"}
	err = kubernikusWaitForClusterV1(klient, d.Id(), target, pending, timeout)
	if err != nil {
		return kubernikusHandleErrorV1("Error waiting for cluster to be deleted", err)
	}

	return nil
}

func resourceCCloudKubernetesV1Import(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	name := parts[0]

	if len(name) == 0 {
		err := fmt.Errorf("Invalid format specified for Kubernetes. Format must be <name>[/<is_admin>]")
		return nil, err
	}

	var isAdmin bool
	var err error
	if len(parts) == 2 {
		isAdmin, err = strconv.ParseBool(parts[1])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse is_admin field: %s", err)
		}
	}

	d.SetId(name)
	d.Set("is_admin", isAdmin)

	return []*schema.ResourceData{d}, nil
}
