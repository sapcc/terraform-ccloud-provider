package domains

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// List returns a Pager which allows you to iterate over a collection of
// domains.
func List(c *gophercloud.ServiceClient) pagination.Pager {
	return pagination.NewPager(c, listURL(c), func(r pagination.PageResult) pagination.Page {
		return DomainPage{pagination.SinglePageBase(r)}
	})
}

// Get retrieves a specific domain based on its unique ID.
func Get(c *gophercloud.ServiceClient, id string) (r GetResult) {
	_, r.Err = c.Get(getURL(c, id), &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200},
	})
	return
}

// UpdateOptsBuilder allows extensions to add additional parameters to the
// Update request.
type UpdateOptsBuilder interface {
	ToDomainUpdateMap() (map[string]interface{}, error)
}

// UpdateOpts represents the attributes used when updating an existing
// domain.
type UpdateOpts struct {
	// ID of the domain
	DomainID string `json:"domain_id,omitempty"`
	// Name of the domain
	DomainName string `json:"domain_name,omitempty"`
	// Description of the domain
	Description string `json:"description,omitempty"`
	// SAP-User-Id of primary contact for the domain
	ResponsiblePrimaryContactID string `json:"responsible_primary_contact_id" required:"true"`
	// Email-address of primary contact for the domain
	ResponsiblePrimaryContactEmail string `json:"responsible_primary_contact_email" required:"true"`
	// SAP-User-Id of the controller who is responsible for the domain / the costobject
	ResponsibleControllerID string `json:"responsible_controller_id,omitempty"`
	// Email-address or DL of the person/group who is controlling the domain / the costobject
	ResponsibleControllerEmail string `json:"responsible_controller_email,omitempty"`
	// Freetext field for additional information for domain
	AdditionalInformation string `json:"additional_information,omitempty"`
	// The cost object structure
	CostObject CostObject `json:"cost_object" required:"true"`
	// Collector of the domain
	Collector string `json:"collector"`
	// Region of the domain
	Region string `json:"region"`
}

// ToDomainUpdateMap builds a request body from UpdateOpts.
func (opts UpdateOpts) ToDomainUpdateMap() (map[string]interface{}, error) {
	return gophercloud.BuildRequestBody(opts, "")
}

// Update accepts a UpdateOpts struct and updates an existing domain using
// the values provided.
func Update(c *gophercloud.ServiceClient, id string, opts UpdateOptsBuilder) (r UpdateResult) {
	b, err := opts.ToDomainUpdateMap()
	if err != nil {
		r.Err = err
		return
	}
	_, r.Err = c.Put(updateURL(c, id), b, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200},
	})
	return
}

func DomainToUpdateOpts(domain *Domain) UpdateOpts {
	return UpdateOpts{
		DomainID:                       domain.DomainID,
		DomainName:                     domain.DomainName,
		Description:                    domain.Description,
		ResponsiblePrimaryContactID:    domain.ResponsiblePrimaryContactID,
		ResponsiblePrimaryContactEmail: domain.ResponsiblePrimaryContactEmail,
		ResponsibleControllerID:        domain.ResponsibleControllerID,
		ResponsibleControllerEmail:     domain.ResponsibleControllerEmail,
		AdditionalInformation:          domain.AdditionalInformation,
		CostObject:                     domain.CostObject,
		Collector:                      domain.Collector,
		Region:                         domain.Region,
	}
}
