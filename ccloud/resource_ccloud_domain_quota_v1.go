package ccloud

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/sapcc/gophercloud-sapcc/resources/v1/domains"
	"github.com/sapcc/limes"
)

func resourceCCloudDomainQuotaV1() *schema.Resource {
	quotaResource := &schema.Resource{
		SchemaVersion: 1,

		Read:   resourceCCloudDomainQuotaV1Read,
		Update: resourceCCloudDomainQuotaV1CreateOrUpdate,
		Create: resourceCCloudDomainQuotaV1CreateOrUpdate,
		Delete: resourceCCloudDomainQuotaV1Delete,
		Importer: &schema.ResourceImporter{
			State: resourceCCloudDomainQuotaV1Import,
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
				Required: true,
				ForceNew: true,
			},
		},
	}

	for service, resources := range limesServices {
		elem := &schema.Resource{
			Schema: make(map[string]*schema.Schema, len(resources)),
		}

		for resource := range resources {
			elem.Schema[resource] = &schema.Schema{
				Type:     schema.TypeFloat,
				Required: false,
				Optional: true,
				Computed: true,
			}
		}

		quotaResource.Schema[sanitize(service)] = &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     elem,
			MaxItems: 1,
		}
	}

	return quotaResource
}

func resourceCCloudDomainQuotaV1Read(d *schema.ResourceData, meta interface{}) error {
	domainID := d.Get("domain_id").(string)

	log.Printf("[QUOTA] Reading Quota for: %s", domainID)

	config := meta.(*Config)
	limes, err := config.limesV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack limes client: %s", err)
	}

	quota, err := domains.Get(limes, domainID, domains.GetOpts{}).Extract()
	if err != nil {
		return fmt.Errorf("Error getting Limes domain: %s", err)
	}

	for service, resources := range limesServices {
		res := make(map[string]float64)
		for resource := range resources {
			if quota.Services[service] == nil || quota.Services[service].Resources[resource] == nil {
				continue
			}
			res[resource] = float64(quota.Services[service].Resources[resource].DomainQuota)
			log.Printf("[QUOTA] %s.%s: %s", service, resource, toString(quota.Services[service].Resources[resource]))
		}
		d.Set(sanitize(service), []map[string]float64{res})
	}

	d.Set("region", GetRegion(d, config))

	return nil
}

func resourceCCloudDomainQuotaV1CreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	domainID := d.Get("domain_id").(string)
	services := limes.QuotaRequest{}

	log.Printf("[QUOTA] Updating Quota for: %s", domainID)

	config := meta.(*Config)
	client, err := config.limesV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack limes client: %s", err)
	}

	for _service, resources := range limesServices {
		service := sanitize(_service)
		if _, ok := d.GetOk(service); ok && d.HasChange(service) {
			log.Printf("[QUOTA] Service Changed: %s", service)

			quota := limes.ServiceQuotaRequest{Resources: make(limes.ResourceQuotaRequest)}
			for resource, unit := range resources {
				key := fmt.Sprintf("%s.0.%s", service, resource)

				if d.HasChange(key) {
					v := d.Get(key)
					log.Printf("[QUOTA] Resource Changed: %s", key)
					quota.Resources[resource] = limes.ValueWithUnit{Value: uint64(v.(float64)), Unit: unit}
					log.Printf("[QUOTA] %s.%s: %s", service, resource, quota.Resources[resource].String())
				}
			}
			services[_service] = quota
		}
	}

	opts := domains.UpdateOpts{Services: services}
	err = domains.Update(client, domainID, opts).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error updating Limes domain: %s", err)
	}

	log.Printf("[QUOTA] Resulting Quota for: %s", domainID)

	d.SetId(domainID)

	return resourceCCloudDomainQuotaV1Read(d, meta)
}

func resourceCCloudDomainQuotaV1Delete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}

func resourceCCloudDomainQuotaV1Import(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("domain_id", d.Id())

	return []*schema.ResourceData{d}, nil
}
