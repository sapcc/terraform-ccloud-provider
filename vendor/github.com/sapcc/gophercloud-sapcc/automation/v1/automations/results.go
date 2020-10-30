package automations

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

const (
	invalidMarker = "-1"
)

type commonResult struct {
	gophercloud.Result
}

// Extract is a function that accepts a result and extracts an automation
// resource.
func (r commonResult) Extract() (*Automation, error) {
	var s Automation
	err := r.ExtractInto(&s)
	return &s, err
}

func (r commonResult) ExtractInto(v interface{}) error {
	return r.Result.ExtractIntoStructPtr(v, "")
}

// CreateResult represents the result of a create operation. Call its Extract
// method to interpret it as an Automation.
type CreateResult struct {
	commonResult
}

// GetResult represents the result of a get operation. Call its Extract method
// to interpret it as an Automation.
type GetResult struct {
	commonResult
}

// UpdateResult represents the result of an update operation. Call its Extract
// method to interpret it as an Automation.
type UpdateResult struct {
	commonResult
}

// DeleteResult represents the result of a delete operation. Call its ExtractErr
// method to determine if the request succeeded or failed.
type DeleteResult struct {
	gophercloud.ErrResult
}

// Automation represents a Lyra Automation.
type Automation struct {
	// An automation ID
	ID string `json:"-"`

	// Human-readable name for the automation. Might not be unique.
	Name string `json:"name"`

	// A valid URL to the automation repository.
	Repository string `json:"repository"`

	// The repository revision.
	RepositoryRevision string `json:"repository_revision"`

	// RepositoryAuthenticationEnabled is set to true when a repository_credentials
	// is set
	RepositoryAuthenticationEnabled bool `json:"repository_authentication_enabled"`

	// The parent Openstack project ID.
	ProjectID string `json:"project_id"`

	// The automation timeout in seconds.
	Timeout int `json:"timeout"`

	// The automation tags. Doesn't work yet.
	Tags map[string]string `json:"tags"`

	// The date, when the automation was created.
	CreatedAt time.Time `json:"-"`

	// The date, when the automation was updated.
	UpdatedAt time.Time `json:"-"`

	// The type of the automation. Can be Script or Chef.
	Type string `json:"type"`

	// An ordered list of Chef roles and/or recipes that are run in the exact
	// order.
	RunList []string `json:"run_list"`

	// A map of Chef cookbook attributes.
	ChefAttributes map[string]interface{} `json:"chef_attributes"`

	// The automation log level. Doesn't work yet.
	LogLevel string `json:"log_level"`

	// An enabled debug mode will not delete the temporary working directory
	// on the instance when the automation job exists
	Debug bool `json:"debug"`

	// The Chef version to run the cookbook.
	ChefVersion string `json:"chef_version"`

	// The Script path.
	Path string `json:"path"`

	// The Script arguments list.
	Arguments []string `json:"arguments"`

	// The Script environment map.
	Environment map[string]string `json:"environment"`
}

func (r *Automation) UnmarshalJSON(b []byte) error {
	type tmp Automation
	var s struct {
		tmp
		ID        int                          `json:"id"`
		CreatedAt gophercloud.JSONRFC3339Milli `json:"created_at"`
		UpdatedAt gophercloud.JSONRFC3339Milli `json:"updated_at"`
	}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r = Automation(s.tmp)

	r.ID = strconv.Itoa(s.ID)
	r.CreatedAt = time.Time(s.CreatedAt)
	r.UpdatedAt = time.Time(s.UpdatedAt)

	return nil
}

// AutomationPage is the page returned by a pager when traversing over a collection
// of automations.
type AutomationPage struct {
	pagination.MarkerPageBase
}

// NextPageURL is invoked when a paginated collection of automations has reached
// the end of a page and the pager seeks to traverse over a new one. In order
// to do this, it needs to construct the next page's URL.
func (r AutomationPage) NextPageURL() (string, error) {
	currentURL := r.URL
	mark, err := r.Owner.LastMarker()
	if err != nil {
		return "", err
	}
	if mark == invalidMarker {
		return "", nil
	}

	q := currentURL.Query()
	q.Set("page", mark)
	currentURL.RawQuery = q.Encode()
	return currentURL.String(), nil
}

// LastMarker returns the next page in a ListResult.
func (r AutomationPage) LastMarker() (string, error) {
	totalPages := -1
	currentPage := -1
	var err error

	page := r.URL.Query().Get("page")
	if page == "" {
		currentPage = 1
	} else {
		currentPage, err = strconv.Atoi(page)
		if err != nil {
			return invalidMarker, err
		}
		if currentPage < 1 {
			currentPage = 1
		}
	}

	if pages, ok := r.Header["Pagination-Pages"]; ok {
		for _, p := range pages {
			totalPages, err = strconv.Atoi(p)
			if err != nil {
				return invalidMarker, err
			}
			break
		}
	}

	if currentPage >= totalPages {
		return invalidMarker, nil
	}

	return strconv.Itoa(currentPage + 1), nil
}

// IsEmpty checks whether an AutomationPage struct is empty.
func (r AutomationPage) IsEmpty() (bool, error) {
	automations, err := ExtractAutomations(r)
	return len(automations) == 0, err
}

// ExtractAutomations accepts a Page struct, specifically an AutomationPage
// struct, and extracts the elements into a slice of Automation structs. In
// other words, a generic collection is mapped into a relevant slice.
func ExtractAutomations(r pagination.Page) ([]Automation, error) {
	var s []Automation
	err := ExtractAutomationsInto(r, &s)
	return s, err
}

func ExtractAutomationsInto(r pagination.Page, v interface{}) error {
	return r.(AutomationPage).Result.ExtractIntoSlicePtr(v, "")
}
