// +build all core data_projects

package azuredevops

// The tests in this file use the mock clients in mock_client.go to mock out
// the Azure DevOps client operations.

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/taskagent"
	"github.com/microsoft/terraform-provider-azuredevops/azdosdkmocks"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/testhelper"
	"github.com/stretchr/testify/require"
)

func TestDataSourceAgentPool_Read_TestEmptyAgentPoolList(t *testing.T) {
	agentPoolListEmpty := []taskagent.TaskAgentPool{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskAgentClient := azdosdkmocks.NewMockTaskagentClient(ctrl)
	clients := &config.AggregatedClient{
		TaskAgentClient: taskAgentClient,
		Ctx:             context.Background(),
	}

	taskAgentClient.
		EXPECT().
		GetAgentPools(clients.Ctx, taskagent.GetAgentPoolsArgs{}).
		Return(&agentPoolListEmpty, nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, dataAzureAgentPools().Schema, nil)
	err := dataSourceAgentPoolsRead(resourceData, clients)
	require.Nil(t, err)
	agentPoolSet := resourceData.Get("agent_pools").(*schema.Set)
	require.NotNil(t, agentPoolSet)
	require.Equal(t, 0, agentPoolSet.Len())
}

var dataTestAgentPools = []taskagent.TaskAgentPool{
	{
		Id:            converter.Int(111),
		Name:          converter.String("AgentPool"),
		PoolType:      &taskagent.TaskAgentPoolTypeValues.Automation,
		AutoProvision: converter.Bool(false),
	},
	{
		Id:            converter.Int(65092),
		Name:          converter.String("AgentPool_AutoProvisioned"),
		PoolType:      &taskagent.TaskAgentPoolTypeValues.Automation,
		AutoProvision: converter.Bool(true),
	},
	{
		Id:            converter.Int(650792),
		Name:          converter.String("AgentPool_Deployment"),
		PoolType:      &taskagent.TaskAgentPoolTypeValues.Deployment,
		AutoProvision: converter.Bool(false),
	},
}

func TestDataSourceAgentPool_Read_TestFindAllAgentPools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	taskAgentClient := azdosdkmocks.NewMockTaskagentClient(ctrl)
	clients := &config.AggregatedClient{
		TaskAgentClient: taskAgentClient,
		Ctx:             context.Background(),
	}

	taskAgentClient.
		EXPECT().
		GetAgentPools(clients.Ctx, taskagent.GetAgentPoolsArgs{}).
		Return(&dataTestAgentPools, nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, dataAzureAgentPools().Schema, nil)
	err := dataSourceAgentPoolsRead(resourceData, clients)
	require.Nil(t, err)
	agentPoolSet := resourceData.Get("agent_pools").(*schema.Set)
	require.NotNil(t, agentPoolSet)
	require.Equal(t, len(dataTestAgentPools), agentPoolSet.Len())
}

/**
 * Begin acceptance tests
 */

// Verifies that the following sequence of events occurrs without error:
//	(1) TF can create a project
//	(2) A data source is added to the configuration, and that data source can find the created project

func TestAccAgentPools_DataSource(t *testing.T) {
	agentPoolName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	agentPool1Name := agentPoolName + "_1"

	createAgent1 := testhelper.TestAccAgentPoolResourceAppendPoolNameToResourceName(agentPool1Name)
	createAgent2 := testhelper.TestAccAgentPoolResourceAppendPoolNameToResourceName(agentPoolName + "_2")
	agentPoolsData := testhelper.TestAccAgentPoolsDataSource()
	createAndGetAgentPools := fmt.Sprintf("%s\n%s\n%s", createAgent1, createAgent2, agentPoolsData)

	tfNode := "data.azuredevops_agent_pools.pools"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testhelper.TestAccPreCheck(t, nil) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: createAndGetAgentPools,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "agent_pools.0.name"),
					resource.TestCheckResourceAttr(tfNode, "agent_pools.0.name", agentPool1Name),
					resource.TestCheckResourceAttrSet(tfNode, "agent_pools"),
				),
			},
		},
	})
}
