package azuredevops

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"strconv"
)

func dataAzureAgentPool() *schema.Resource {
	return &schema.Resource{
		Read: resourceAzureAgentPoolReadByPoolId,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true},
			"pool_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_provision": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"pool_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceAzureAgentPoolReadByPoolId(d *schema.ResourceData, m interface{}) error {
	poolId := strconv.Itoa(d.Get("pool_id").(int))
	d.SetId(poolId)
	err := resourceAzureAgentPoolRead(d, m)
	return err
}
