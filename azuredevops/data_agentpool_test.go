// +build all core resource_project

package azuredevops

// The tests in this file use the mock clients in mock_client.go to mock out
// the Azure DevOps client operations.

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/testhelper"
)

/**
 * Begin acceptance tests
 */

// Verifies that the following sequence of events occurrs without error:
//	(1) That tf can create a agent pool and that data source can find the created agent pool
func TestAccAgentPool_DataSource(t *testing.T) {
	agentPoolName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	createAgentPool := testhelper.TestAccAgentPoolResource(agentPoolName)
	createAndGetAgentPoolData := fmt.Sprintf("%s\n%s", createAgentPool, testhelper.TestAccAgentPoolDataSource())

	tfNode := "data.azuredevops_agent_pool.pool"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testhelper.TestAccPreCheck(t, nil) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: createAndGetAgentPoolData,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "pool_id"),
					resource.TestCheckResourceAttr(tfNode, "name", agentPoolName),
					resource.TestCheckResourceAttr(tfNode, "auto_provision", "false"),
					resource.TestCheckResourceAttr(tfNode, "pool_type", "automation"),
				),
			},
		},
	})
}

func init() {
	InitProvider()
}
