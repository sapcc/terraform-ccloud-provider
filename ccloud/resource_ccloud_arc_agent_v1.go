package ccloud

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/sapcc/gophercloud-sapcc/arc/v1/agents"
)

func resourceCCloudArcAgentV1() *schema.Resource {
	return &schema.Resource{
		Read:   resourceCCloudArcAgentV1Read,
		Create: resourceCCloudArcAgentV1Create,
		Update: resourceCCloudArcAgentV1Update,
		Delete: resourceCCloudArcAgentV1Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"agent_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"filter"},
				ValidateFunc:  validation.NoZeroValues,
			},

			"filter": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"agent_id"},
				ValidateFunc:  validation.NoZeroValues,
			},

			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
			},

			"force_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			// Computed attributes
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"project": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"updated_with": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"updated_by": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"all_tags": {
				Type:     schema.TypeMap,
				Computed: true,
			},

			"facts": {
				Type:     schema.TypeMap,
				Computed: true,
			},

			"facts_agents": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func resourceCCloudArcAgentV1Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	agentID := d.Get("agent_id").(string)
	filter := d.Get("filter").(string)
	timeout := d.Timeout(schema.TimeoutCreate)

	agent, err := arcCCloudArcAgentV1WaitForAgent(arcClient, agentID, filter, timeout)
	if err != nil {
		return err
	}

	d.SetId(agent.AgentID)

	err = updateArcAgentTagsV1(arcClient, d.Id(), nil, d.Get("tags"))
	if err != nil {
		return err
	}

	return resourceCCloudArcAgentV1Read(d, meta)
}

func resourceCCloudArcAgentV1Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	agent, err := agents.Get(arcClient, d.Id()).Extract()
	if err != nil {
		return CheckDeleted(d, err, "Unable to retrieve ccloud_arc_agent_v1")
	}

	arcCCloudArcAgentV1ReadAgent(d, arcClient, agent, GetRegion(d, config))

	return nil
}

func resourceCCloudArcAgentV1Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	oldTags, newTags := d.GetChange("tags")
	err = updateArcAgentTagsV1(arcClient, d.Id(), oldTags, newTags)
	if err != nil {
		return err
	}

	return resourceCCloudArcAgentV1Read(d, meta)
}

func resourceCCloudArcAgentV1Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	if !d.Get("force_delete").(bool) {
		// Wait for the instance to delete before moving on.
		log.Printf("[DEBUG] Waiting for compute instance (%s) to delete", d.Id())

		computeClient, err := config.ComputeV2Client(GetRegion(d, config))
		if err != nil {
			return fmt.Errorf("Error creating OpenStack compute client: %s", err)
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"ACTIVE", "SHUTOFF"},
			Target:     []string{"DELETED", "SOFT_DELETED"},
			Refresh:    ServerV2StateRefreshFunc(computeClient, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutDelete),
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("Error waiting for compute instance (%s) to delete: %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting ccloud_arc_agent_v1: %s", d.Id())
	err = agents.Delete(arcClient, d.Id()).ExtractErr()
	if err != nil {
		return CheckDeleted(d, err, "Error deleting ccloud_arc_agent_v1")
	}

	return nil
}
