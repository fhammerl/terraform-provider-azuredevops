package azuredevops

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/microsoft/azure-devops-go-api/azuredevops/taskagent"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/datahelper"
)

func dataAgentPools() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAgentPoolsRead,

		Schema: map[string]*schema.Schema{
			"agentpools": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      getAgentPoolHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"pool_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"auto_provision": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func getAgentPoolHash(v interface{}) int {
	return hashcode.String(v.(map[string]interface{})["id"].(string))
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
	err = d.Set("agentPools", results)
	if err != nil {
		return err
	}
	return nil
}

func flattenAgentPoolReferences(input *[]taskagent.TaskAgentPool) ([]interface{}, error) {
	if input == nil {
		return []interface{}{}, nil
	}

	results := make([]interface{}, 0)

	for _, element := range *input {
		output := make(map[string]interface{})
		if element.Name != nil {
			output["name"] = element.Name
		}

		if element.Id != nil {
			output["id"] = element.Id
		}

		if element.PoolType != nil {
			output["pool_type"] = element.PoolType
		}

		if element.AutoProvision != nil {
			output["auto_provision"] = element.AutoProvision
		}

		results = append(results, output)
	}

	return results, nil
}

func getAgentPools(clients *config.AggregatedClient) (*[]taskagent.TaskAgentPool, error) {
	return clients.TaskAgentClient.GetAgentPools(clients.Ctx, taskagent.GetAgentPoolsArgs{})
}
