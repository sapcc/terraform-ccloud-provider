package ccloud

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceCCloudArcJobIDsV1() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCCloudArcJobIDsV1Read,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"agent_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 86400),
			},

			"agent": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"chef", "execute",
				}, false),
			},

			"action": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"script", "zero", "tarball",
				}, false),
			},

			"status": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"queued", "executing", "failed", "complete",
				}, false),
			},

			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceCCloudArcJobIDsV1Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	arcClient, err := config.arcV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack Arc client: %s", err)
	}

	jobs, err := arcCCloudArcJobV1Filter(d, arcClient, "ccloud_arc_job_ids_v1")
	if err != nil {
		return err
	}

	var jobIDs []string
	for _, j := range jobs {
		jobIDs = append(jobIDs, j.RequestID)
	}

	log.Printf("[DEBUG] Retrieved %d jobs in ccloud_arc_job_ids_v1: %+v", len(jobs), jobs)

	d.SetId(fmt.Sprintf("%d", hashcode.String(strings.Join(jobIDs, ""))))
	d.Set("ids", jobIDs)

	return nil
}
