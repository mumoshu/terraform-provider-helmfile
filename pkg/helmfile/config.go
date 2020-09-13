package helmfile

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

type ProviderInstance struct {
	MaxDiffOutputLen int
}

func New(d *schema.ResourceData) *ProviderInstance {
	return &ProviderInstance{
		MaxDiffOutputLen: d.Get(KeyMaxDiffOutputLen).(int),
	}
}
