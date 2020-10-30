package ccloud

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/sapcc/gophercloud-sapcc/clients"
)

func (c *Config) limesV1Client(region string) (*gophercloud.ServiceClient, error) {
	return c.CommonServiceClientInit(clients.NewLimesV1, region, "resources")
}

func (c *Config) kubernikusV1Client(region string, isAdmin bool) (*Kubernikus, error) {
	if err := c.Authenticate(); err != nil {
		return nil, err
	}

	serviceType := "kubernikus"
	if isAdmin {
		serviceType = "kubernikus-kubernikus"
	}

	return NewKubernikusV1(c, gophercloud.EndpointOpts{
		Type:         serviceType,
		Region:       c.DetermineRegion(region),
		Availability: clientconfig.GetEndpointType(c.EndpointType),
	})
}

func (c *Config) arcV1Client(region string) (*gophercloud.ServiceClient, error) {
	return c.CommonServiceClientInit(clients.NewArcV1, region, "arc")
}

func (c *Config) automationV1Client(region string) (*gophercloud.ServiceClient, error) {
	return c.CommonServiceClientInit(clients.NewAutomationV1, region, "automation")
}

func (c *Config) billingClient(region string) (*gophercloud.ServiceClient, error) {
	return c.CommonServiceClientInit(clients.NewBilling, region, "sapcc-billing")
}
