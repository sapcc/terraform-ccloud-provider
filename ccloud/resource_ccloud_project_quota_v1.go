package ccloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/sapcc/gophercloud-limes/resources/v1/projects"
	"github.com/sapcc/limes"
)

func resourceCCloudProjectQuotaV1() *schema.Resource {
	quotaResource := &schema.Resource{
		SchemaVersion: 1,

		Read:   resourceCCloudProjectQuotaV1Read,
		Update: resourceCCloudProjectQuotaV1CreateOrUpdate,
		Create: resourceCCloudProjectQuotaV1CreateOrUpdate,
		Delete: resourceCCloudProjectQuotaV1Delete,
		Importer: &schema.ResourceImporter{
			State: resourceCCloudProjectQuotaV1Import,
		},

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
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}

	for service, resources := range SERVICES {
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

func resourceCCloudProjectQuotaV1Read(d *schema.ResourceData, meta interface{}) error {
	domainID := d.Get("domain_id").(string)
	projectID := d.Get("project_id").(string)

	log.Printf("[QUOTA] Reading Quota for: %s/%s", domainID, projectID)

	config := meta.(*Config)
	limes, err := config.limesV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack limes client: %s", err)
	}

	quota, err := projects.Get(limes, domainID, projectID, projects.GetOpts{}).Extract()
	if err != nil {
		return fmt.Errorf("Error getting Limes project: %s", err)
	}

	for service, resources := range SERVICES {
		res := make(map[string]float64)
		for resource := range resources {
			if quota.Services[service] == nil || quota.Services[service].Resources[resource] == nil {
				continue
			}
			res[resource] = float64(quota.Services[service].Resources[resource].Quota)
			log.Printf("[QUOTA] %s.%s: %s", service, resource, toString(quota.Services[service].Resources[resource]))
		}
		d.Set(sanitize(service), []map[string]float64{res})
	}

	d.Set("region", GetRegion(d, config))

	return nil
}

func resourceCCloudProjectQuotaV1CreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	domainID := d.Get("domain_id").(string)
	projectID := d.Get("project_id").(string)
	services := limes.QuotaRequest{}

	log.Printf("[QUOTA] Updating Quota for: %s/%s", domainID, projectID)

	config := meta.(*Config)
	client, err := config.limesV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack limes client: %s", err)
	}

	if d.Id() == "" {
		// when the project was just created, it may not yet appeared in the limes
		if err := limesCCloudProjectQuotaV1WaitForProject(client, domainID, projectID, d.Timeout(schema.TimeoutCreate)); err != nil {
			return err
		}
	}

	for _service, resources := range SERVICES {
		service := sanitize(_service)
		if _, ok := d.GetOk(service); ok && d.HasChange(service) {
			log.Printf("[QUOTA] Service Changed: %s", service)

			quota := limes.ServiceQuotaRequest{}
			for resource, unit := range resources {
				key := fmt.Sprintf("%s.0.%s", service, resource)

				if d.HasChange(key) {
					v := d.Get(key)
					log.Printf("[QUOTA] Resource Changed: %s", key)
					quota[resource] = limes.ValueWithUnit{uint64(v.(float64)), unit}
					log.Printf("[QUOTA] %s.%s: %s", service, resource, quota[resource].String())
				}
			}
			services[_service] = quota
		}
	}

	opts := projects.UpdateOpts{Services: services}
	_, err = projects.Update(client, domainID, projectID, opts)
	if err != nil {
		return fmt.Errorf("Error updating Limes project: %s", err)
	}

	log.Printf("[QUOTA] Resulting Quota for: %s/%s", domainID, projectID)

	d.SetId(projectID)

	return resourceCCloudProjectQuotaV1Read(d, meta)
}

func resourceCCloudProjectQuotaV1Delete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}

func resourceCCloudProjectQuotaV1Import(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		err := fmt.Errorf("Invalid format specified for Quota. Format must be <domain id>/<project id>")
		return nil, err
	}

	d.SetId(parts[1])
	d.Set("domain_id", parts[0])
	d.Set("project_id", parts[1])

	return []*schema.ResourceData{d}, nil
}
