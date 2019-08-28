package domains

import (
	"encoding/json"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

type commonResult struct {
	gophercloud.Result
}

// Extract is a function that accepts a result and extracts a domain
// resource.
func (r commonResult) Extract() (*Domain, error) {
	var s Domain
	err := r.ExtractInto(&s)
	return &s, err
}

func (r commonResult) ExtractInto(v interface{}) error {
	return r.Result.ExtractIntoStructPtr(v, "")
}

// GetResult represents the result of a get operation. Call its Extract method
// to interpret it as a Domain.
type GetResult struct {
	commonResult
}

// UpdateResult represents the result of an update operation. Call its Extract
// method to interpret it as a Domain.
type UpdateResult struct {
	commonResult
}

// Domain represents a Billing Domain.
type Domain struct {
	// Instance ID
	IID int `json:"iid"`
	// ID of the domain
	DomainID string `json:"domain_id"`
	// Name of the domain
	DomainName string `json:"domain_name"`
	// Description of the domain
	Description string `json:"description"`
	// SAP-User-Id of primary contact for the domain
	ResponsiblePrimaryContactID string `json:"responsible_primary_contact_id"`
	// Email-address of primary contact for the domain
	ResponsiblePrimaryContactEmail string `json:"responsible_primary_contact_email"`
	// SAP-User-Id of the controller who is responsible for the domain / the costobject
	ResponsibleControllerID string `json:"responsible_controller_id"`
	// Email-address or DL of the person/group who is controlling the domain / the costobject
	ResponsibleControllerEmail string `json:"responsible_controller_email"`
	// Freetext field for additional information for domain
	AdditionalInformation string `json:"additional_information"`
	// The cost object structure
	CostObject CostObject `json:"cost_object"`
	// The date, when the domain was created.
	CreatedAt time.Time `json:"-"`
	// The date, when the domain was updated.
	ChangedAt time.Time `json:"-"`
	// The ID of the user, who did the last change.
	ChangedBy string `json:"changed_by"`
	// Only contained in Server response: True, if the given masterdata are complete; Otherwise false
	IsComplete bool `json:"is_complete"`
	// Only contained in Server response: Human readable text, showing, what information are missing
	MissingAttributes string `json:"missing_attributes"`
	// Collector of the domain
	Collector string `json:"collector"`
	// Region of the domain
	Region string `json:"region"`
}

// The cost object structure
type CostObject struct {
	// Set to true, if the costobject should be inheritable for subprojects
	ProjectsCanInherit bool `json:"projects_can_inherit"`
	// Name of the costobject
	Name string `json:"name,omitempty"`
	// Costobject-Type Type of the costobject
	// IO, CC, WBS, SO
	Type string `json:"type,omitempty"`
}

func (r *Domain) UnmarshalJSON(b []byte) error {
	type tmp Domain
	var s struct {
		tmp
		CreatedAt gophercloud.JSONRFC3339MilliNoZ `json:"created_at"`
		ChangedAt gophercloud.JSONRFC3339MilliNoZ `json:"changed_at"`
	}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r = Domain(s.tmp)

	r.CreatedAt = time.Time(s.CreatedAt)
	r.ChangedAt = time.Time(s.ChangedAt)

	return nil
}

func (r *Domain) MarshalJSON() ([]byte, error) {
	type ext struct {
		CreatedAt string `json:"created_at"`
		ChangedAt string `json:"changed_at"`
	}

	type tmp struct {
		Domain
		ext
	}

	s := tmp{
		*r,
		ext{
			CreatedAt: r.CreatedAt.Format(gophercloud.RFC3339MilliNoZ),
			ChangedAt: r.ChangedAt.Format(gophercloud.RFC3339MilliNoZ),
		},
	}

	return json.Marshal(s)
}

// DomainPage is the page returned by a pager when traversing over a collection
// of domains.
type DomainPage struct {
	pagination.SinglePageBase
}

// ExtractDomains accepts a Page struct, specifically a DomainPage
// struct, and extracts the elements into a slice of Domain structs. In
// other words, a generic collection is mapped into a relevant slice.
func ExtractDomains(r pagination.Page) ([]Domain, error) {
	var s []Domain
	err := ExtractDomainsInto(r, &s)
	return s, err
}

func ExtractDomainsInto(r pagination.Page, v interface{}) error {
	return r.(DomainPage).Result.ExtractIntoSlicePtr(v, "")
}
