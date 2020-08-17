package helmfile

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

type ResourceRead interface {
	Id() string
	Get(string) interface{}
}

type ResourceReadWrite interface {
	ResourceRead
	Set(string, interface{}) error
}

type ResourceReadWriteEmbedded struct {
	m map[string]interface{}
}

func (m *ResourceReadWriteEmbedded) Id() string {
	return ""
}

func (m *ResourceReadWriteEmbedded) Get(k string) interface{} {
	return m.m[k]
}

func (m *ResourceReadWriteEmbedded) Set(k string, v interface{}) error {
	m.m[k] = v
	return nil
}

type ResourceReadWriteDiff struct {
	*schema.ResourceDiff
}

func (d *ResourceReadWriteDiff) Set(key string, value interface{}) error {
	return d.SetNew(key, value)
}

func resourceDiffToFields(d *schema.ResourceDiff) ResourceReadWrite {
	return &ResourceReadWriteDiff{
		ResourceDiff: d,
	}
}
