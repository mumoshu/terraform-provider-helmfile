package helmfile

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rs/xid"
	"log"
)

const KeyValuesFiles = "values_files"
const KeyValues = "values"
const KeySelector = "selector"
const KeyEnvironmentVariables = "environment_variables"
const KeyWorkingDirectory = "working_directory"
const KeyPath = "path"
const KeyContent = "content"
const KeyEnvironment = "environment"
const KeyBin = "binary"
const KeyHelmBin = "helm_binary"
const KeyDiffOutput = "diff_output"
const KeyError = "error"
const KeyApplyOutput = "apply_output"
const KeyDirty = "dirty"
const KeyConcurrency = "concurrency"

const HelmfileDefaultPath = "helmfile.yaml"

var ReleaseSetSchema = map[string]*schema.Schema{
	KeyValuesFiles: {
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: false,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	KeyValues: {
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: false,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	KeySelector: {
		Type:     schema.TypeMap,
		Optional: true,
		ForceNew: false,
	},
	KeyEnvironmentVariables: {
		Type:     schema.TypeMap,
		Optional: true,
		Elem:     schema.TypeString,
	},
	KeyWorkingDirectory: {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: false,
		Default:  "",
	},
	KeyPath: {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: false,
		Default:  "",
	},
	KeyContent: {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: false,
	},
	KeyBin: {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: false,
		Default:  "helmfile",
	},
	KeyHelmBin: {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: false,
		Default:  "helm",
	},
	KeyEnvironment: {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: false,
		Default:  "helm",
	},
	KeyDiffOutput: {
		Type:     schema.TypeString,
		Computed: true,
	},
	KeyApplyOutput: {
		Type:     schema.TypeString,
		Computed: true,
	},
	KeyError: {
		Type:     schema.TypeString,
		Computed: true,
	},
	KeyDirty: {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
	},
	KeyConcurrency: {
		Type:     schema.TypeInt,
		Optional: true,
		Default:  0,
	},
}

func resourceShellHelmfileReleaseSet() *schema.Resource {
	return &schema.Resource{
		Create:        resourceReleaseSetCreate,
		Delete:        resourceReleaseSetDelete,
		Read:          resourceReleaseSetRead,
		Update:        resourceReleaseSetUpdate,
		CustomizeDiff: resourceReleaseSetDiff,
		Schema:        ReleaseSetSchema,
	}
}

//helpers to unwravel the recursive bits by adding a base condition
func resourceReleaseSetCreate(d *schema.ResourceData, meta interface{}) error {
	fs, err := NewReleaseSet(d)
	if err != nil {
		return err
	}

	if err := CreateReleaseSet(fs, d); err != nil {
		return err
	}

	d.MarkNewResource()

	//create random uuid for the id
	id := xid.New().String()
	d.SetId(id)

	return nil
}

func resourceReleaseSetRead(d *schema.ResourceData, meta interface{}) error {
	fs, err := NewReleaseSet(d)
	if err != nil {
		return err
	}

	return ReadReleaseSet(fs, d)
}

func resourceReleaseSetDiff(d *schema.ResourceDiff, meta interface{}) error {
	old, new := d.GetChange(KeyWorkingDirectory)
	log.Printf("Getting old and new working directories for id %q: old = %v, new = %v, got = %v", d.Id(), old, new, d.Get(KeyWorkingDirectory))

	fs, err := NewReleaseSet(d)
	if err != nil {
		return err
	}

	diff, err := DiffReleaseSet(fs, resourceDiffToFields(d))
	if err != nil {
		return err
	}

	if diff != "" {
		d.SetNewComputed(KeyApplyOutput)
	}

	return nil
}

func resourceReleaseSetUpdate(d *schema.ResourceData, meta interface{}) error {
	fs, err := NewReleaseSet(d)
	if err != nil {
		return err
	}

	return UpdateReleaseSet(fs, d)
}

func resourceReleaseSetDelete(d *schema.ResourceData, meta interface{}) error {
	fs, err := NewReleaseSet(d)
	if err != nil {
		return err
	}

	if err := DeleteReleaseSet(fs, d); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
