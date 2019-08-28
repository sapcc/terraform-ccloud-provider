package ccloud

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/sapcc/gophercloud-billing/billing/domains"
)

func dataSourceCCloudBillingDomainMasterdata() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCCloudBillingDomainMasterdataRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"domain_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"additional_information": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_primary_contact_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_primary_contact_email": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_controller_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"responsible_controller_email": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cost_object": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"projects_can_inherit": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"changed_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"changed_by": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"is_complete": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"missing_attributes": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"collector": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCCloudBillingDomainMasterdataRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	billing, err := config.billingClient(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack billing client: %s", err)
	}

	domainID := d.Get("domain_id").(string)

	var domain *domains.Domain
	if domainID == "" {
		allPages, err := domains.List(billing).AllPages()
		if err != nil {
			return fmt.Errorf("Error getting billing domain masterdata: %s", err)
		}

		allDomains, err := domains.ExtractDomains(allPages)
		if err != nil {
			return fmt.Errorf("Error extracting billing domains masterdata: %s", err)
		}

		if len(allDomains) != 1 {
			return fmt.Errorf("Error getting billing domain masterdata: expecting 1 domain, got %d", len(allDomains))
		}

		domain = &allDomains[0]
	} else {
		domain, err = domains.Get(billing, domainID).Extract()
		if err != nil {
			return fmt.Errorf("Error getting billing domain masterdata: %s", err)
		}
	}

	log.Printf("[DEBUG] Retrieved domain masterdata: %+v", domain)

	d.SetId(domain.DomainID)

	d.Set("domain_id", domain.DomainID)
	d.Set("domain_name", domain.DomainName)
	d.Set("description", domain.Description)
	d.Set("responsible_primary_contact_id", domain.ResponsiblePrimaryContactID)
	d.Set("responsible_primary_contact_email", domain.ResponsiblePrimaryContactEmail)
	d.Set("responsible_controller_id", domain.ResponsibleControllerID)
	d.Set("responsible_controller_email", domain.ResponsibleControllerEmail)
	d.Set("additional_information", domain.AdditionalInformation)
	d.Set("cost_object", billingDomainFlattenCostObject(domain.CostObject))
	d.Set("created_at", domain.CreatedAt.Format(time.RFC3339))
	d.Set("changed_at", domain.ChangedAt.Format(time.RFC3339))
	d.Set("changed_by", domain.ChangedBy)
	d.Set("is_complete", domain.IsComplete)
	d.Set("missing_attributes", domain.MissingAttributes)
	d.Set("collector", domain.Collector)

	d.Set("region", GetRegion(d, config))

	return nil
}
