package ccloud

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/sapcc/kubernikus/pkg/api/client/operations"
	"log"
)

func dataSourceCCloudKubernetesBootstrapV1() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCCloudKubernetesBootstrapV1Read,
		Schema: map[string]*schema.Schema{
			"kube_config": {
				Type: schema.TypeString,
				Computed: true,
			},
			"kube_client_ca": {
				Type: schema.TypeString,
				Computed: true,
			},
			"kube_client_ca_file": {
				Type: schema.TypeString,
				Computed: true,
			},
			"config": {
				Type: schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type: schema.TypeString,
				Required: true,
			},
			"is_admin": {
				Type: schema.TypeBool,
				Optional: true,
				Default: false,
			},
		},
	}
}

func dataSourceCCloudKubernetesBootstrapV1Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	klient, err := config.kubernikusV1Client(GetRegion(d, config), d.Get("is_admin").(bool))
	if err != nil {
		return fmt.Errorf("error creating Kubernikus client: %s", err)
	}

	params := operations.NewGetBootstrapConfigParams().WithName(d.Get("cluster_name").(string))
	configOK, err :=  klient.GetBootstrapConfig(params, klient.authFunc())
	if err != nil {
		return fmt.Errorf("error getting bootstrap config for cluster: %s", err)
	}

	pretty, _ := json.MarshalIndent(configOK.Payload, "", "  ")
	log.Printf("[DEBUG] bootstrap payload: %s", string(pretty))

	_ = d.Set("kube_config", configOK.Payload.Kubeconfig)
	_ = d.Set("kube_client_ca", configOK.Payload.KubeletClientsCA)
	_ = d.Set("kube_client_ca_file", configOK.Payload.KubeletClientsCAFile)
	_ = d.Set("config", configOK.Payload.Config)
	d.SetId( d.Get("cluster_name").(string) )


	return nil
}
