package ccloud

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/gophercloud/gophercloud"
	"github.com/sapcc/gophercloud-lyra/automation/v1/runs"
)

func resourceCCloudAutomationRunV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceCCloudAutomationRunV1Create,
		Read:   resourceCCloudAutomationRunV1Read,
		Delete: func(*schema.ResourceData, interface{}) error { return nil },

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},

			"automation_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"selector": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			// Computed
			"automation_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"repository": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "This field is not returned by the Lyra API",
			},

			"repository_revision": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"automation_attributes": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"log": {
				Type:     schema.TypeString,
				Computed: true,
				// Don't print the huge log during the terraform plan/apply
				Sensitive: true,
			},

			"jobs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"project_id": {
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

			"owner": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceCCloudAutomationRunV1Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	automationClient, err := config.automationV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	createOpts := runs.CreateOpts{
		AutomationID: d.Get("automation_id").(string),
		Selector:     d.Get("selector").(string),
	}

	log.Printf("[DEBUG] ccloud_automation_run_v1 create options: %#v", createOpts)

	run, err := runs.Create(automationClient, createOpts).Extract()
	if err != nil {
		return fmt.Errorf("Error creating ccloud_automation_run_v1: %s", err)
	}

	d.SetId(run.ID)

	timeout := d.Timeout(schema.TimeoutCreate)
	target := []string{"completed", "failed"}
	pending := []string{"preparing", "executing"}
	err = waitForAutomationRunV1(automationClient, run.ID, target, pending, timeout)
	if err != nil {
		return err
	}

	return resourceCCloudAutomationRunV1Read(d, meta)
}

func resourceCCloudAutomationRunV1Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	automationClient, err := config.automationV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack Automation client: %s", err)
	}

	run, err := runs.Get(automationClient, d.Id()).Extract()
	if err != nil {
		return CheckDeleted(d, err, "Unable to retrieve ccloud_automation_run_v1")
	}

	d.Set("automation_id", run.AutomationID)
	d.Set("automation_name", run.AutomationName)
	d.Set("selector", run.Selector)
	d.Set("repository_revision", run.RepositoryRevision)

	automationAttributes, err := json.Marshal(run.AutomationAttributes)
	if err != nil {
		log.Printf("[DEBUG] resourceCCloudAutomationRunV1Read: Cannot marshal run.AutomationAttributes: %s", err)
	}
	d.Set("automation_attributes", string(automationAttributes))

	d.Set("state", run.State)
	d.Set("created_at", run.CreatedAt.Format(time.RFC3339))
	d.Set("updated_at", run.UpdatedAt.Format(time.RFC3339))
	d.Set("log", run.Log)
	d.Set("jobs", run.Jobs)
	d.Set("owner", flattenAutomationiOwnerV1(run.Owner))
	d.Set("project_id", run.ProjectID)

	d.Set("region", GetRegion(d, config))

	return nil
}

func flattenAutomationiOwnerV1(owner runs.Owner) []interface{} {
	return []interface{}{map[string]interface{}{
		"id":          owner.ID,
		"name":        owner.Name,
		"domain_id":   owner.DomainID,
		"domain_name": owner.DomainName,
	}}
}

func waitForAutomationRunV1(automationClient *gophercloud.ServiceClient, id string, target []string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for %s run to become %v.", id, target)

	stateConf := &resource.StateChangeConf{
		Target:     target,
		Pending:    pending,
		Refresh:    automationRunV1GetState(automationClient, id),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func automationRunV1GetState(automationClient *gophercloud.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		run, err := runs.Get(automationClient, id).Extract()
		if err != nil {
			return nil, "", fmt.Errorf("Unable to retrieve %s ccloud_automation_run_v1: %s", id, err)
		}

		return run, run.State, nil
	}
}
