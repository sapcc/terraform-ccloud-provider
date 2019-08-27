package projects

import (
	"encoding/json"
	"time"
	//"log"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

type commonResult struct {
	gophercloud.Result
}

// Extract is a function that accepts a result and extracts an project
// resource.
func (r commonResult) Extract() (*Project, error) {
	var s Project
	err := r.ExtractInto(&s)
	return &s, err
}

func (r commonResult) ExtractInto(v interface{}) error {
	return r.Result.ExtractIntoStructPtr(v, "")
}

// GetResult represents the result of a get operation. Call its Extract method
// to interpret it as an Project.
type GetResult struct {
	commonResult
}

// UpdateResult represents the result of an update operation. Call its Extract
// method to interpret it as an Project.
type UpdateResult struct {
	commonResult
}

// Project represents a Billing Project.
type Project struct {
	// A project ID
	ProjectID string `json:"project_id"`
	// Human-readable name for the project. Might not be unique.
	ProjectName string `json:"project_name"`
	// Technical of the domain in which the project is contained
	DomainID string `json:"domain_id"`
	// Name of the domain
	DomainName string `json:"domain_name"`
	// Description of the project
	Description string `json:"description"`
	// A project parent ID
	ParentID string `json:"parent_id"`
	// A project type
	ProjectType string `json:"project_type"`
	// SAP-User-Id of primary contact for the project
	ResponsiblePrimaryContactID string `json:"responsible_primary_contact_id"`
	// Email-address of primary contact for the project
	ResponsiblePrimaryContactEmail string `json:"responsible_primary_contact_email"`
	// SAP-User-Id of the person who is responsible for operating the project
	ResponsibleOperatorID string `json:"responsible_operator_id"`
	// Email-address or DL of the person/group who is operating the project
	ResponsibleOperatorEmail string `json:"responsible_operator_email"`
	// SAP-User-Id of the person who is responsible for the security of the project
	ResponsibleSecurityExpertID string `json:"responsible_security_expert_id"`
	// Email-address or DL of the person/group who is responsible for the security of the project
	ResponsibleSecurityExpertEmail string `json:"responsible_security_expert_email"`
	// SAP-User-Id of the product owner
	ResponsibleProductOwnerID string `json:"responsible_product_owner_id"`
	// Email-address or DL of the product owner
	ResponsibleProductOwnerEmail string `json:"responsible_product_owner_email"`
	// SAP-User-Id of the controller who is responsible for the project / the costobject
	ResponsibleControllerID string `json:"responsible_controller_id"`
	// Email-address or DL of the person/group who is controlling the project / the costobject
	ResponsibleControllerEmail string `json:"responsible_controller_email"`
	// Indicating if the project is directly or indirectly creating revenue
	// Allowed values: [generating, enabling, other]
	RevenueRelevance string `json:"revenue_relevance"`
	// Indicates how important the project for the business is. Possible values: [dev,test,prod]
	// Allowed values: [dev, test, prod]
	BusinessCriticality string `json:"business_criticality"`
	// If the number is unclear, always provide the lower end --> means always > number_of_endusers (-1 indicates that it is infinite)
	NumberOfEndusers int `json:"number_of_endusers"`
	// Freetext field for additional information for project
	AdditionalInformation string `json:"additional_information"`
	// The cost object structure
	CostObject CostObject `json:"cost_object"`
	// The date, when the project was created.
	CreatedAt time.Time `json:"-"`
	// The date, when the project was updated.
	ChangedAt time.Time `json:"-"`
	// The ID of the user, who did the last change.
	ChangedBy string `json:"changed_by"`
	// Only contained in Server response: True, if the given masterdata are complete; Otherwise false
	IsComplete bool `json:"is_complete"`
	// Only contained in Server response: Human readable text, showing, what information are missing
	MissingAttributes string `json:"missing_attributes"`
	// Only contained in Server response: Collector of the project
	Collector string `json:"collector"`
	// Only contained in Server response: Region of the project
	Region string `json:"region"`
}

// The cost object structure
type CostObject struct {
	// Shows, if the CO is inherited. Mandatory, if name/type not set
	Inherited bool `json:"inherited"`
	// Name of the costobject. Mandatory, if inherited not true
	Name string `json:"name,omitempty"`
	// Costobject-Type Type of the costobject. Mandatory, if inherited not true
	// IO, CC, WBS, SO
	Type string `json:"type,omitempty"`
}

func (r *Project) UnmarshalJSON(b []byte) error {
	type tmp Project
	var s struct {
		tmp
		CreatedAt gophercloud.JSONRFC3339MilliNoZ `json:"created_at"`
		ChangedAt gophercloud.JSONRFC3339MilliNoZ `json:"changed_at"`
	}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r = Project(s.tmp)

	r.CreatedAt = time.Time(s.CreatedAt)
	r.ChangedAt = time.Time(s.ChangedAt)

	return nil
}

func (r *Project) MarshalJSON() ([]byte, error) {
	type ext struct {
		CreatedAt string `json:"created_at"`
		ChangedAt string `json:"changed_at"`
	}

	type tmp struct {
		Project
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

// ProjectPage is the page returned by a pager when traversing over a collection
// of projects.
type ProjectPage struct {
	pagination.SinglePageBase
}

// ExtractProjects accepts a Page struct, specifically an ProjectPage
// struct, and extracts the elements into a slice of Project structs. In
// other words, a generic collection is mapped into a relevant slice.
func ExtractProjects(r pagination.Page) ([]Project, error) {
	var s []Project
	err := ExtractProjectsInto(r, &s)
	return s, err
}

func ExtractProjectsInto(r pagination.Page, v interface{}) error {
	return r.(ProjectPage).Result.ExtractIntoSlicePtr(v, "")
}
