package ccloud

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/sapcc/gophercloud-arc/arc/v1/agents"
)

func dataSourceCCloudArcAgentIDsV1() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCCloudArcAgentIDsV1Read,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"filter": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// computed attributes
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceCCloudArcAgentIDsV1Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	filter := d.Get("filter").(string)

	listOpts := agents.ListOpts{Filter: filter}

	log.Printf("[DEBUG] ccloud_arc_agent_ids_v1 list options: %#v", listOpts)

	allPages, err := agents.List(arcClient, listOpts).AllPages()
	if err != nil {
		return fmt.Errorf("Unable to list ccloud_arc_agent_ids_v1: %s", err)
	}

	allAgents, err := agents.ExtractAgents(allPages)
	if err != nil {
		return fmt.Errorf("Unable to retrieve ccloud_arc_agent_ids_v1: %s", err)
	}

	var agentIDs []string
	for _, a := range allAgents {
		agentIDs = append(agentIDs, a.AgentID)
	}

	log.Printf("[DEBUG] Retrieved %d agents in ccloud_arc_agent_ids_v1: %+v", len(allAgents), allAgents)

	d.SetId(fmt.Sprintf("%d", hashcode.String(strings.Join(agentIDs, ""))))
	d.Set("ids", agentIDs)

	return nil
}
