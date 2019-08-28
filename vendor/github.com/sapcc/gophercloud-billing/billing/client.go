package billing

import (
	"github.com/gophercloud/gophercloud"
)

// NewBilling creates a ServiceClient that may be used with the billing package.
func NewBilling(client *gophercloud.ProviderClient, endpointOpts gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {
	sc := new(gophercloud.ServiceClient)
	endpointOpts.ApplyDefaults("sapcc-billing")
	url, err := client.EndpointLocator(endpointOpts)
	if err != nil {
		return sc, err
	}

	resourceBase := url + "masterdata/"
	return &gophercloud.ServiceClient{
		ProviderClient: client,
		Endpoint:       url,
		Type:           "sapcc-billing",
		ResourceBase:   resourceBase,
	}, nil
}
