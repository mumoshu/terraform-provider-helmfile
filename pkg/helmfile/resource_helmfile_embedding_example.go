package helmfile

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rs/xid"
)

func resourceHelmfileEmbeddingExample() *schema.Resource {
	return &schema.Resource{
		Create:        resourceHelmfileEmbeddingExampleCreate,
		Delete:        resourceHelmfileEmbeddingExampleDelete,
		Read:          resourceHelmfileEmbeddingExampleRead,
		Update:        resourceHelmfileEmbeddingExampleUpdate,
		CustomizeDiff: resourceHelmfileEmbeddingExampleCustomizeDiff,
		Schema: map[string]*schema.Schema{
			"embedded": {
				Type:     schema.TypeList,
				Optional: true,
				// Seems like we need to make the whole attributed as computed, rather than only a subset of
				// nested fields as computed.
				// Otherwise, we get nested computed fields like diff_output always shown as "(known after apply)",
				// rather than the actual planned value
				Computed: true,
				Elem: &schema.Resource{
					Schema: ReleaseSetSchema,
				},
			},
		},
	}
}

func ExtractEmbeddedReleaseSetResources(data ResourceRead, attr string) ([]map[string]interface{}, error) {
	var entries []map[string]interface{}

	if d := data.Get(attr); d == nil {
		return nil, fmt.Errorf("getting field: no attribute named %q found", attr)
	} else {
		ifs := d.([]interface{})

		for _, i := range ifs {
			entries = append(entries, i.(map[string]interface{}))
		}
	}

	return entries, nil
}

func resourceHelmfileEmbeddingExampleCreate(data *schema.ResourceData, i interface{}) error {
	embeddedResources, err := ExtractEmbeddedReleaseSetResources(data, "embedded")
	if err != nil {
		return err
	}

	for _, e := range embeddedResources {
		fs := &ResourceReadWriteEmbedded{m: e}

		rs, err := NewReleaseSet(fs)
		if err != nil {
			return err
		}

		if err := CreateReleaseSet(rs, fs); err != nil {
			return err
		}
	}

	// Note: If you missed marking new resource and setting the id, it may end up unintuitive tf error like:
	//   "... produced an unexpected new value for was present, but now absent."
	//

	data.MarkNewResource()

	//create random uuid for the id
	id := xid.New().String()
	data.SetId(id)

	data.Set("embedded", embeddedResources)

	dump("create", embeddedResources)

	return nil
}

func resourceHelmfileEmbeddingExampleDelete(data *schema.ResourceData, i interface{}) error {
	embeddedResources, err := ExtractEmbeddedReleaseSetResources(data, "embedded")
	if err != nil {
		return err
	}

	for _, e := range embeddedResources {
		fs := &ResourceReadWriteEmbedded{m: e}

		rs, err := NewReleaseSet(fs)
		if err != nil {
			return err
		}

		if err := DeleteReleaseSet(rs, fs); err != nil {
			return err
		}
	}
	return nil
}

func resourceHelmfileEmbeddingExampleRead(data *schema.ResourceData, i interface{}) error {
	embeddedResources, err := ExtractEmbeddedReleaseSetResources(data, "embedded")
	if err != nil {
		return err
	}

	for _, e := range embeddedResources {
		fs := &ResourceReadWriteEmbedded{m: e}

		rs, err := NewReleaseSet(fs)
		if err != nil {
			return err
		}

		if err := ReadReleaseSet(rs, fs); err != nil {
			return err
		}
	}
	return nil
}

func resourceHelmfileEmbeddingExampleUpdate(data *schema.ResourceData, i interface{}) error {
	embeddedResources, err := ExtractEmbeddedReleaseSetResources(data, "embedded")
	if err != nil {
		return err
	}

	for _, e := range embeddedResources {
		fs := &ResourceReadWriteEmbedded{m: e}

		rs, err := NewReleaseSet(fs)
		if err != nil {
			return err
		}

		if err := UpdateReleaseSet(rs, fs); err != nil {
			return err
		}
	}

	data.Set("embedded", embeddedResources)

	return nil
}

func resourceHelmfileEmbeddingExampleCustomizeDiff(resourceDiff *schema.ResourceDiff, i interface{}) error {
	embeddedResources, err := ExtractEmbeddedReleaseSetResources(resourceDiff, "embedded")
	if err != nil {
		return err
	}

	var hasDiff bool

	for _, e := range embeddedResources {
		fs := &ResourceReadWriteEmbedded{m: e}

		rs, err := NewReleaseSet(fs)
		if err != nil {
			return err
		}

		// DryRun=true should be set if terraform-provider-helmfile is integrated into an another provider
		// and the helmfile_release_set resource is embedded into a resource tha also declares the target K8s cluster,
		// which means before creating the cluster the provider needs to show helmfile-diff result without K8s
		//
		// DryRun=false and Kubeconfig!="" should be set if the K8s cluster is already there and you have the kubeconfig to
		// access the K8s API
		diff, err := DiffReleaseSet(rs, fs, WithDiffConfig(DiffConfig{DryRun: false, Kubeconfig: ""}))
		if err != nil {
			return err
		}

		if diff != "" {
			hasDiff = true
		}
	}

	if hasDiff {
		resourceDiff.SetNew("embedded", embeddedResources)
	}

	dump("diff", embeddedResources)

	return nil
}
