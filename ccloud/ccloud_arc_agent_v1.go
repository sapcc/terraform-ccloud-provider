package ccloud

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/sapcc/gophercloud-arc/arc/v1/agents"
)

func arcCCloudArcAgentV1ReadAgent(d *schema.ResourceData, arcClient *gophercloud.ServiceClient, agent *agents.Agent, region string) {
	if len(agent.Facts) == 0 {
		facts, err := agents.GetFacts(arcClient, agent.AgentID).Extract()
		if err != nil {
			log.Printf("Unable to retrieve facts for ccloud_arc_agent_v1: %s", err)
		}
		agent.Facts = facts
		log.Printf("[DEBUG] Retrieved ccloud_arc_agent_v1 facts %s: %+v", agent.AgentID, agent.Facts)
	}

	d.Set("display_name", agent.DisplayName)
	d.Set("agent_id", agent.AgentID)
	d.Set("project", agent.Project)
	d.Set("organization", agent.Organization)
	d.Set("all_tags", agent.Tags)
	d.Set("created_at", agent.CreatedAt.Format(time.RFC3339))
	d.Set("updated_at", agent.UpdatedAt.Format(time.RFC3339))
	d.Set("updated_with", agent.UpdatedWith)
	d.Set("updated_by", agent.UpdatedBy)

	d.Set("facts", expandToMapStringString(agent.Facts))
	factsAgents, _ := agent.Facts["agents"]
	if v, ok := factsAgents.(map[string]interface{}); ok {
		d.Set("facts_agents", expandToMapStringString(v))
	} else {
		d.Set("facts_agents", map[string]string{})
	}

	d.Set("region", region)
}

func arcCCloudArcAgentV1WaitForAgent(arcClient *gophercloud.ServiceClient, agentID, filter string, timeout time.Duration) (*agents.Agent, error) {
	var agent interface{}
	var msg string
	var err error

	// This condition is required, otherwise zero timeout will always raise:
	// "timeout while waiting for state to become 'active'"
	if timeout > 0 {
		// Retryable case, when timeout is set
		waitForAgent := &resource.StateChangeConf{
			Target:         []string{"active"},
			Refresh:        arcCCloudArcAgentV1GetAgent(arcClient, agentID, filter, timeout),
			Timeout:        timeout,
			Delay:          1 * time.Second,
			MinTimeout:     1 * time.Second,
			NotFoundChecks: 1000, // workaround for default 20 retries, when the resource is nil
		}
		agent, err = waitForAgent.WaitForState()
	} else {
		// When timeout is not set, just get the agent
		agent, msg, err = arcCCloudArcAgentV1GetAgent(arcClient, agentID, filter, timeout)()
	}

	if len(msg) > 0 && msg != "active" {
		return nil, fmt.Errorf(msg)
	}

	if err != nil {
		return nil, err
	}

	return agent.(*agents.Agent), nil
}

func arcCCloudArcAgentV1GetAgent(arcClient *gophercloud.ServiceClient, agentID, filter string, timeout time.Duration) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		var agent *agents.Agent
		var err error

		if len(agentID) == 0 && len(filter) == 0 {
			return nil, "", fmt.Errorf("At least one of agent_id or filter parameters is expected in ccloud_arc_agent_v1")
		}

		if len(agentID) > 0 {
			agent, err = agents.Get(arcClient, agentID).Extract()
			if err != nil {
				if _, ok := err.(gophercloud.ErrDefault404); ok && timeout > 0 {
					// Retryable case, when timeout is set
					return nil, fmt.Sprintf("Unable to retrieve %s ccloud_arc_agent_v1: %s", agentID, err), nil
				}
				return nil, "", fmt.Errorf("Unable to retrieve %s ccloud_arc_agent_v1: %s", agentID, err)
			}
		} else {
			listOpts := agents.ListOpts{Filter: filter}

			log.Printf("[DEBUG] ccloud_arc_agent_v1 list options: %#v", listOpts)

			allPages, err := agents.List(arcClient, listOpts).AllPages()
			if err != nil {
				return nil, "", fmt.Errorf("Unable to list ccloud_arc_agent_v1: %s", err)
			}

			allAgents, err := agents.ExtractAgents(allPages)
			if err != nil {
				return nil, "", fmt.Errorf("Unable to retrieve ccloud_arc_agent_v1: %s", err)
			}

			if len(allAgents) == 0 {
				// Retryable case, when timeout is set
				return nil, "No ccloud_arc_agent_v1 found", nil
			}

			if len(allAgents) > 1 {
				return nil, "", fmt.Errorf("More than one ccloud_arc_agent_v1 found (%d)", len(allAgents))
			}

			agent = &allAgents[0]
		}

		log.Printf("[DEBUG] Retrieved ccloud_arc_agent_v1 %s: %+v", agent.AgentID, *agent)

		return agent, "active", nil
	}
}

func updateArcAgentTagsV1(arcClient *gophercloud.ServiceClient, agentID string, oldTagsRaw, newTagsRaw interface{}) error {
	var tagsToDelete []string
	oldTags, _ := oldTagsRaw.(map[string]interface{})
	newTags, _ := newTagsRaw.(map[string]interface{})

	// Determine if any tag keys were removed from the configuration.
	// Then request those keys to be deleted.
	for oldKey := range oldTags {
		var found bool
		for newKey := range newTags {
			if oldKey == newKey {
				found = true
			}
		}

		if !found {
			tagsToDelete = append(tagsToDelete, oldKey)
		}
	}

	for _, key := range tagsToDelete {
		err := agents.DeleteTag(arcClient, agentID, key).ExtractErr()
		if err != nil {
			return fmt.Errorf("Error deleting %s tag from %s ccloud_arc_agent_v1: %s", key, agentID, err)
		}
	}

	// Update existing tags and add any new tags.
	tagsOpts := make(agents.Tags)
	for k, v := range newTags {
		tagsOpts[k] = v.(string)
	}

	err := agents.CreateTags(arcClient, agentID, tagsOpts).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error updating tags for %s ccloud_arc_agent_v1: %s", agentID, err)
	}

	return nil
}

func arcAgentV1ParseTimeout(raw interface{}) (time.Duration, error) {
	if list, ok := raw.([]interface{}); ok {
		for _, t := range list {
			if timeout, ok := t.(map[string]interface{}); ok {
				if v, ok := timeout["read"]; ok {
					return time.ParseDuration(v.(string))
				}
			}
		}
	}

	return time.Duration(0), nil
}

func ServerV2StateRefreshFunc(client *gophercloud.ServiceClient, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		s, err := servers.Get(client, instanceID).Extract()
		if err != nil {
			if _, ok := err.(gophercloud.ErrDefault404); ok {
				return s, "DELETED", nil
			}
			return nil, "", err
		}

		return s, s.Status, nil
	}
}
