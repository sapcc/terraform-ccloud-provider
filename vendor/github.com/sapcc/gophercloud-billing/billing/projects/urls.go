package projects

import "github.com/gophercloud/gophercloud"

func resourceURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL("projects", id)
}

func rootURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("projects")
}

func listURL(c *gophercloud.ServiceClient) string {
	return rootURL(c)
}

func getURL(c *gophercloud.ServiceClient, id string) string {
	return resourceURL(c, id)
}

func updateURL(c *gophercloud.ServiceClient, id string) string {
	return resourceURL(c, id)
}
