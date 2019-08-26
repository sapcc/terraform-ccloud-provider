package ccloud

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/hashicorp/terraform/helper/pathorcontents"
	"github.com/hashicorp/terraform/terraform"
	"github.com/kayrus/gophercloud-arc/arc"
	"github.com/kayrus/gophercloud-lyra/automation"
	"github.com/sapcc/gophercloud-billing/billing"
	"github.com/sapcc/gophercloud-limes/resources"
)

type Config struct {
	CACertFile                  string
	ClientCertFile              string
	ClientKeyFile               string
	Cloud                       string
	DefaultDomain               string
	DomainID                    string
	DomainName                  string
	EndpointOverrides           map[string]interface{}
	EndpointType                string
	IdentityEndpoint            string
	Insecure                    *bool
	Password                    string
	ProjectDomainName           string
	ProjectDomainID             string
	Region                      string
	Swauth                      bool
	TenantID                    string
	TenantName                  string
	Token                       string
	UserDomainName              string
	UserDomainID                string
	Username                    string
	UserID                      string
	ApplicationCredentialID     string
	ApplicationCredentialName   string
	ApplicationCredentialSecret string
	useOctavia                  bool
	MaxRetries                  int

	OsClient *gophercloud.ProviderClient
}

// LoadAndValidate performs the authentication and initial configuration
// of an OpenStack Provider Client. This sets up the HTTP client and
// authenticates to an OpenStack cloud.
//
// Individual Service Clients are created later in this file.
func (c *Config) LoadAndValidate() error {
	// Make sure at least one of auth_url or cloud was specified.
	if c.IdentityEndpoint == "" && c.Cloud == "" {
		return fmt.Errorf("One of 'auth_url' or 'cloud' must be specified")
	}

	validEndpoint := false
	validEndpoints := []string{
		"internal", "internalURL",
		"admin", "adminURL",
		"public", "publicURL",
		"",
	}

	for _, endpoint := range validEndpoints {
		if c.EndpointType == endpoint {
			validEndpoint = true
		}
	}

	if !validEndpoint {
		return fmt.Errorf("Invalid endpoint type provided")
	}

	clientOpts := new(clientconfig.ClientOpts)

	// If a cloud entry was given, base AuthOptions on a clouds.yaml file.
	if c.Cloud != "" {
		clientOpts.Cloud = c.Cloud

		cloud, err := clientconfig.GetCloudFromYAML(clientOpts)
		if err != nil {
			return err
		}

		if c.Region == "" && cloud.RegionName != "" {
			c.Region = cloud.RegionName
		}

		if c.CACertFile == "" && cloud.CACertFile != "" {
			c.CACertFile = cloud.CACertFile
		}

		if c.ClientCertFile == "" && cloud.ClientCertFile != "" {
			c.ClientCertFile = cloud.ClientCertFile
		}

		if c.ClientKeyFile == "" && cloud.ClientKeyFile != "" {
			c.ClientKeyFile = cloud.ClientKeyFile
		}

		if c.Insecure == nil && cloud.Verify != nil {
			v := (!*cloud.Verify)
			c.Insecure = &v
		}
	} else {
		authInfo := &clientconfig.AuthInfo{
			AuthURL:                     c.IdentityEndpoint,
			DefaultDomain:               c.DefaultDomain,
			DomainID:                    c.DomainID,
			DomainName:                  c.DomainName,
			Password:                    c.Password,
			ProjectDomainID:             c.ProjectDomainID,
			ProjectDomainName:           c.ProjectDomainName,
			ProjectID:                   c.TenantID,
			ProjectName:                 c.TenantName,
			Token:                       c.Token,
			UserDomainID:                c.UserDomainID,
			UserDomainName:              c.UserDomainName,
			Username:                    c.Username,
			UserID:                      c.UserID,
			ApplicationCredentialID:     c.ApplicationCredentialID,
			ApplicationCredentialName:   c.ApplicationCredentialName,
			ApplicationCredentialSecret: c.ApplicationCredentialSecret,
		}
		clientOpts.AuthInfo = authInfo
	}

	ao, err := clientconfig.AuthOptions(clientOpts)
	if err != nil {
		return err
	}

	client, err := openstack.NewClient(ao.IdentityEndpoint)
	if err != nil {
		return err
	}

	// Set UserAgent
	client.UserAgent.Prepend(terraform.UserAgentString())

	config := &tls.Config{}
	if c.CACertFile != "" {
		caCert, _, err := pathorcontents.Read(c.CACertFile)
		if err != nil {
			return fmt.Errorf("Error reading CA Cert: %s", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caCert))
		config.RootCAs = caCertPool
	}

	if c.Insecure == nil {
		config.InsecureSkipVerify = false
	} else {
		config.InsecureSkipVerify = *c.Insecure
	}

	if c.ClientCertFile != "" && c.ClientKeyFile != "" {
		clientCert, _, err := pathorcontents.Read(c.ClientCertFile)
		if err != nil {
			return fmt.Errorf("Error reading Client Cert: %s", err)
		}
		clientKey, _, err := pathorcontents.Read(c.ClientKeyFile)
		if err != nil {
			return fmt.Errorf("Error reading Client Key: %s", err)
		}

		cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
		if err != nil {
			return err
		}

		config.Certificates = []tls.Certificate{cert}
		config.BuildNameToCertificate()
	}

	// if OS_DEBUG is set, log the requests and responses
	var osDebug bool
	if os.Getenv("OS_DEBUG") != "" {
		osDebug = true
	}

	transport := &http.Transport{Proxy: http.ProxyFromEnvironment, TLSClientConfig: config}
	client.HTTPClient = http.Client{
		Transport: &LogRoundTripper{
			Rt:         transport,
			OsDebug:    osDebug,
			MaxRetries: c.MaxRetries,
		},
	}

	// If using Swift Authentication, there's no need to validate authentication normally.
	if !c.Swauth {
		err = openstack.Authenticate(client, *ao)
		if err != nil {
			return err
		}
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries should be a positive value")
	}

	c.OsClient = client

	return nil
}

// determineEndpoint is a helper method to determine if the user wants to
// override an endpoint returned from the catalog.
func (c *Config) determineEndpoint(client *gophercloud.ServiceClient, service string) *gophercloud.ServiceClient {
	finalEndpoint := client.ResourceBaseURL()

	if v, ok := c.EndpointOverrides[service]; ok {
		if endpoint, ok := v.(string); ok && endpoint != "" {
			finalEndpoint = endpoint
			client.Endpoint = endpoint
			client.ResourceBase = ""
		}
	}

	log.Printf("[DEBUG] OpenStack Endpoint for %s: %s", service, finalEndpoint)

	return client
}

func (c *Config) determineRegion(region string) string {
	// If a resource-level region was not specified, and a provider-level region was set,
	// use the provider-level region.
	if region == "" && c.Region != "" {
		region = c.Region
	}

	log.Printf("[DEBUG] OpenStack Region is: %s", region)
	return region
}

// getEndpointType is a helper method to determine the endpoint type
// requested by the user.
func (c *Config) getEndpointType() gophercloud.Availability {
	if c.EndpointType == "internal" || c.EndpointType == "internalURL" {
		return gophercloud.AvailabilityInternal
	}
	if c.EndpointType == "admin" || c.EndpointType == "adminURL" {
		return gophercloud.AvailabilityAdmin
	}
	return gophercloud.AvailabilityPublic
}

func (c *Config) limesV1Client(region string) (*gophercloud.ServiceClient, error) {
	return resources.NewLimesV1(c.OsClient, gophercloud.EndpointOpts{
		Region:       c.determineRegion(region),
		Availability: c.getEndpointType(),
	})
}

func (c *Config) kubernikusV1Client(region string, isAdmin bool) (*Kubernikus, error) {
	serviceType := "kubernikus"
	if isAdmin {
		serviceType = "kubernikus-kubernikus"
	}

	return NewKubernikusV1(c.OsClient, gophercloud.EndpointOpts{
		Type:         serviceType,
		Region:       c.determineRegion(region),
		Availability: gophercloud.AvailabilityPublic,
	})
}

func (c *Config) arcV1Client(region string) (*gophercloud.ServiceClient, error) {
	client, err := arc.NewArcV1(c.OsClient, gophercloud.EndpointOpts{
		Region:       c.determineRegion(region),
		Availability: c.getEndpointType(),
	})

	if err != nil {
		return client, err
	}

	// Check if an endpoint override was specified for the arc service.
	client = c.determineEndpoint(client, "arc")

	return client, nil
}

func (c *Config) automationV1Client(region string) (*gophercloud.ServiceClient, error) {
	client, err := automation.NewAutomationV1(c.OsClient, gophercloud.EndpointOpts{
		Region:       c.determineRegion(region),
		Availability: c.getEndpointType(),
	})

	if err != nil {
		return client, err
	}

	// Check if an endpoint override was specified for the automation service.
	client = c.determineEndpoint(client, "automation")

	return client, nil
}

func (c *Config) computeV2Client(region string) (*gophercloud.ServiceClient, error) {
	client, err := openstack.NewComputeV2(c.OsClient, gophercloud.EndpointOpts{
		Region:       c.determineRegion(region),
		Availability: c.getEndpointType(),
	})

	if err != nil {
		return client, err
	}

	// Check if an endpoint override was specified for the compute service.
	client = c.determineEndpoint(client, "compute")

	return client, nil
}

func (c *Config) billingClient(region string) (*gophercloud.ServiceClient, error) {
	client, err := billing.NewBilling(c.OsClient, gophercloud.EndpointOpts{
		Region:       c.determineRegion(region),
		Availability: c.getEndpointType(),
	})

	if err != nil {
		return client, err
	}

	// Check if an endpoint override was specified for the billing service.
	client = c.determineEndpoint(client, "billing")

	return client, nil
}
