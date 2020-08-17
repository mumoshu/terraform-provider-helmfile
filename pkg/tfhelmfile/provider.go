package tfhelmfile

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{
			"helmfile_release_set":       resourceShellHelmfileReleaseSet(),
			"helmfile_release":           resourceHelmfileRelease(),
			"helmfile_embedding_example": resourceHelmfileEmbeddingExample(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	return New(), nil
}

// This is a global MutexKV for use within this plugin.
var mutexKV = mutexkv.NewMutexKV()
