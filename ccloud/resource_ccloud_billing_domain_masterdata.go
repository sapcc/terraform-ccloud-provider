package ccloud

import (
	"fmt"
	"log"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/sapcc/gophercloud-billing/billing/domains"
)

func resourceCCloudBillingDomainMasterdata() *schema.Resource {
	return &schema.Resource{
		Read:   resourceCCloudBillingDomainMasterdataRead,
		Update: resourceCCloudBillingDomainMasterdataCreateOrUpdate,
		Create: resourceCCloudBillingDomainMasterdataCreateOrUpdate,
		Delete: schema.RemoveFromState,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"domain_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"domain_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"additional_information": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"responsible_primary_contact_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"responsible_primary_contact_email": {
				Type:     schema.TypeString,
				Required: true,
			},

			"responsible_controller_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"responsible_controller_email": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"cost_object": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"projects_can_inherit": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"IO", "CC", "WBS", "SO",
							}, false),
						},
					},
				},
			},

			"collector": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			// computed parameters
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
		},
	}
}

func resourceCCloudBillingDomainMasterdataCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	billing, err := config.billingClient(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack billing client: %s", err)
	}

	domainID := d.Get("domain_id").(string)

	var domain *domains.Domain
	if d.Id() == "" && domainID == "" {
		// first call, expecting to get current scope domain
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
		// admin mode, when the domain doesn't correspond to the scope
		// or during the update, when domain_id was already set
		domain, err = domains.Get(billing, domainID).Extract()
		if err != nil {
			if _, ok := err.(gophercloud.ErrDefault404); d.Id() != "" || !ok {
				return fmt.Errorf("Error getting billing domain masterdata: %s", err)
			}
			log.Printf("[DEBUG] Error getting billing domain masterdata, probably this domain was not created yet: %s", err)
			domain = &domains.Domain{DomainID: domainID}
		}
	}

	log.Printf("[DEBUG] Retrieved domain masterdata before the created/update: %+v", domain)

	// API doesn't support partial update, thus prefilling the update options with the existing data
	opts := domains.DomainToUpdateOpts(domain)
	opts.DomainID = replaceEmpty(d, "domain_id", opts.DomainID)
	opts.DomainName = replaceEmpty(d, "domain_name", opts.DomainName)
	opts.ResponsiblePrimaryContactID = replaceEmpty(d, "responsible_primary_contact_id", opts.ResponsiblePrimaryContactID)
	opts.ResponsiblePrimaryContactEmail = replaceEmpty(d, "responsible_primary_contact_email", opts.ResponsiblePrimaryContactEmail)
	opts.ResponsibleControllerID = replaceEmpty(d, "responsible_controller_id", opts.ResponsibleControllerID)
	opts.ResponsibleControllerEmail = replaceEmpty(d, "responsible_controller_email", opts.ResponsibleControllerEmail)
	opts.AdditionalInformation = replaceEmpty(d, "additional_information", opts.AdditionalInformation)
	opts.Collector = replaceEmpty(d, "collector", opts.Collector)

	if v := billingDomainExpandCostObject(d.Get("cost_object")); v != (domains.CostObject{}) {
		opts.CostObject = v
	}

	log.Printf("[QUOTA] Updating %s domain masterdata: %+v", opts.DomainID, opts)

	_, err = domains.Update(billing, opts.DomainID, opts).Extract()
	if err != nil {
		return fmt.Errorf("Error updating billing domain masterdata: %s", err)
	}

	if d.Id() == "" {
		d.SetId(opts.DomainID)
	}

	return resourceCCloudBillingDomainMasterdataRead(d, meta)
}

func resourceCCloudBillingDomainMasterdataRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	billing, err := config.billingClient(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack billing client: %s", err)
	}

	domain, err := domains.Get(billing, d.Id()).Extract()
	if err != nil {
		return CheckDeleted(d, err, "Error getting billing domain masterdata")
	}

	log.Printf("[DEBUG] Retrieved domain masterdata: %+v", domain)

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
