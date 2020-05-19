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
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
	agentPools := resourceData.Get("agent_pools").([]interface{})
	require.NotNil(t, agentPools)
	require.Equal(t, 0, len(agentPools))
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
	agentPools := resourceData.Get("agent_pools").([]interface{})
	require.NotNil(t, agentPools)
	require.Equal(t, len(dataTestAgentPools), len(agentPools))
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
	agentPool2Name := agentPoolName + "_2"

	createAgent1 := testhelper.TestAccAgentPoolResourceAppendPoolNameToResourceName(agentPool1Name)
	createAgent2 := testhelper.TestAccAgentPoolResourceAppendPoolNameToResourceName(agentPool2Name)
	agentPoolsData := testhelper.TestAccAgentPoolsDataSource()
	createAgentPools := fmt.Sprintf("%s\n%s", createAgent1, createAgent2)

	tfNode := "data.azuredevops_agent_pools.pools"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testhelper.TestAccPreCheck(t, nil) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: createAgentPools,
			},
			{
				Config: agentPoolsData,
				Check: resource.ComposeTestCheckFunc(
					testAgentPoolExists(tfNode, agentPool1Name),
					testAgentPoolExists(tfNode, agentPool2Name),
				),
			},
		},
	})
}

func testAgentPoolExists(tfNode string, poolName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rootModule := s.RootModule()
		resource, ok := rootModule.Resources[tfNode]
		if !ok {
			return fmt.Errorf("Did not find a project in the TF state")
		}

		is := resource.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", tfNode, rootModule.Path)
		}
		if !containsValue(is.Attributes, poolName) {
			return fmt.Errorf("%s does not contain a pool with name %s", tfNode, poolName)
		}
		return nil
	}
}

func containsValue(m map[string]string, v string) bool {
	for _, x := range m {
		if x == v {
			return true
		}
	}
	return false
}
