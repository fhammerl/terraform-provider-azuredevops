// +build all core data_projects

package azuredevops

// The tests in this file use the mock clients in mock_client.go to mock out
// the Azure DevOps client operations.

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/taskagent"
	"github.com/microsoft/terraform-provider-azuredevops/azdosdkmocks"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
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
