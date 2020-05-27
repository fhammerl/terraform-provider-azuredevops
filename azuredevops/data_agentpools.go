package azuredevops

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/taskagent"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/datahelper"
)

func dataAzureAgentPools() *schema.Resource {
	baseSchema := resourceAzureAgentPool()
	baseSchema.Schema["id"] = &schema.Schema{
		Type: schema.TypeInt,
	}

	for k, v := range baseSchema.Schema {
		baseSchema.Schema[k] = &schema.Schema{
			Type:     v.Type,
			Computed: true,
		}
	}

	return &schema.Resource{
		Read: dataSourceAgentPoolsRead,

		Schema: map[string]*schema.Schema{
			"agent_pools": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: baseSchema.Schema,
				},
			},
		},
	}
}

func dataSourceAgentPoolsRead(d *schema.ResourceData, m interface{}) error {
	clients := m.(*config.AggregatedClient)

	agentPools, err := getAgentPools(clients)
	if err != nil {
		return fmt.Errorf("Error finding agent pools. Error: %v", err)
	}
	log.Printf("[TRACE] plugin.terraform-provider-azuredevops: Read [%d] agent pools from current organization", len(*agentPools))

	results, err := flattenAgentPoolReferences(agentPools)
	if err != nil {
		return fmt.Errorf("Error flattening agentPools. Error: %v", err)
	}

	h := sha1.New()
	agentPoolNames, err := datahelper.GetAttributeValues(results, "name")
	if err != nil {
		return fmt.Errorf("Failed to get list of agent pool names: %v", err)
	}
	if _, err := h.Write([]byte(strings.Join(agentPoolNames, "-"))); err != nil {
		return fmt.Errorf("Unable to compute hash for agent pool names: %v", err)
	}
	d.SetId("agentPools#" + base64.URLEncoding.EncodeToString(h.Sum(nil)))
	err = d.Set("agent_pools", results)
	if err != nil {
		return err
	}
	return nil
}

func flattenAgentPoolReferences(input *[]taskagent.TaskAgentPool) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	results := make([]interface{}, 0)

	for _, element := range *input {
		output := make(map[string]interface{})
		if element.Name != nil {
			output["name"] = *element.Name
		}

		if element.Id != nil {
			output["id"] = *element.Id
		}

		if element.PoolType != nil {
			output["pool_type"] = string(*element.PoolType)
		}

		if element.AutoProvision != nil {
			output["auto_provision"] = *element.AutoProvision
		}

		results = append(results, output)
	}

	return results
}

func getAgentPools(clients *config.AggregatedClient) (*[]taskagent.TaskAgentPool, error) {
	return clients.TaskAgentClient.GetAgentPools(clients.Ctx, taskagent.GetAgentPoolsArgs{})
}
