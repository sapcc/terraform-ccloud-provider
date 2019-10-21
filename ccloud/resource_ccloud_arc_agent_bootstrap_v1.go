package ccloud

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/sapcc/gophercloud-arc/arc/v1/agents"
)

func resourceCCloudArcAgentBootstrapV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceCCloudArcAgentBootstrapV1Create,
		Read:   func(*schema.ResourceData, interface{}) error { return nil },
		Delete: func(*schema.ResourceData, interface{}) error { return nil },

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "cloud-config",
				ValidateFunc: validation.StringInSlice([]string{
					"linux", "windows", "cloud-config", "json",
				}, false),
			},

			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},

			// computed attributes
			"user_data": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"raw_map": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func resourceCCloudArcAgentBootstrapV1Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	var bootstrapType string
	switch d.Get("type") {
	case "linux":
		bootstrapType = "text/x-shellscript"
	case "windows":
		bootstrapType = "text/x-powershellscript"
	case "cloud-config":
		bootstrapType = "text/cloud-config"
	case "json":
		bootstrapType = "application/json"
	}

	createOpts := agents.InitOpts{Accept: bootstrapType}

	log.Printf("[DEBUG] ccloud_arc_agent_bootstrap_v1 create options: %#v", createOpts)

	res := agents.Init(arcClient, createOpts)
	if res.Err != nil {
		return fmt.Errorf("Error creating ccloud_arc_agent_bootstrap_v1: %s", res.Err)
	}

	headers, err := res.ExtractHeaders()
	if err != nil {
		return fmt.Errorf("Error extracting headers while creating ccloud_arc_agent_bootstrap_v1: %s", err)
	}

	if bootstrapType != headers.ContentType {
		return fmt.Errorf("Error verifying headers while creating ccloud_arc_agent_bootstrap_v1: wants '%s', got '%s'", bootstrapType, headers.ContentType)
	}

	data, err := res.ExtractContent()
	if err != nil {
		return fmt.Errorf("Error extracting content while creating ccloud_arc_agent_bootstrap_v1: %s", err)
	}

	userData := string(data)

	d.SetId(fmt.Sprintf("%d", hashcode.String(userData)))

	if bootstrapType == "application/json" {
		var initMap map[string]string
		err = json.Unmarshal(data, &initMap)
		if err != nil {
			return fmt.Errorf("Error unmarshalling JSON content while creating ccloud_arc_agent_bootstrap_v1: %s", err)
		}
		d.Set("raw_map", initMap)
	} else {
		d.Set("raw_map", map[string]string{})
	}

	d.Set("user_data", userData)
	d.Set("region", GetRegion(d, config))

	return nil
}
