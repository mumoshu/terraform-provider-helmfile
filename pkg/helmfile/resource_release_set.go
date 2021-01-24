package helmfile

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/mumoshu/terraform-provider-eksctl/pkg/sdk/tfsdk"
	"github.com/rs/xid"
	"golang.org/x/xerrors"
	"log"
	"os"
	"runtime/debug"
	"strings"
)

const KeyValuesFiles = "values_files"
const KeyValues = "values"
const KeySelector = "selector"
const KeySelectors = "selectors"
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
const KeyReleasesValues = "releases_values"
const KeySkipDiffOnMissingFiles = "skip_diff_on_missing_files"

const HelmfileDefaultPath = "helmfile.yaml"

var ReleaseSetSchema = map[string]*schema.Schema{
	KeyAWSRegion: {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: false,
	},
	KeyAWSProfile: {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: false,
	},
	KeyAWSAssumeRole: tfsdk.SchemaAssumeRole(),
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
	KeySkipDiffOnMissingFiles: {
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
	KeySelectors: {
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: false,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
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
	KeyKubeconfig: {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: false,
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
	KeyVersion: {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: false,
		Default:  "",
	},
	KeyHelmVersion: {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: false,
		Default:  "",
	},
	KeyHelmDiffVersion: {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: false,
		Default:  "",
	},
	KeyEnvironment: {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: false,
		Default:  "",
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
	KeyReleasesValues: {
		Type:     schema.TypeMap,
		Optional: true,
		ForceNew: false,
	},
}

func resourceHelmfileReleaseSet() *schema.Resource {
	return &schema.Resource{
		Create:        resourceReleaseSetCreate,
		Delete:        resourceReleaseSetDelete,
		Read:          resourceReleaseSetRead,
		Update:        resourceReleaseSetUpdate,
		CustomizeDiff: resourceReleaseSetDiff,
		Importer: &schema.ResourceImporter{
			State: resourceReleaseSetImport,
		},
		Schema: ReleaseSetSchema,
	}
}

//helpers to unwravel the recursive bits by adding a base condition
func resourceReleaseSetCreate(d *schema.ResourceData, meta interface{}) (finalErr error) {
	defer func() {
		if err := recover(); err != nil {
			finalErr = fmt.Errorf("unhandled error: %v\n%s", err, debug.Stack())
		}
	}()

	fs, err := NewReleaseSet(d)
	if err != nil {
		return err
	}

	if err := CreateReleaseSet(newContext(d), fs, d); err != nil {
		return fmt.Errorf("creating release set: %w", err)
	}

	d.MarkNewResource()

	d.SetId(newId())

	return nil
}

func newId() string {
	//create random uuid for the id
	id := xid.New().String()

	return id
}

func resourceReleaseSetRead(d *schema.ResourceData, meta interface{}) (finalErr error) {
	defer func() {
		if err := recover(); err != nil {
			finalErr = fmt.Errorf("unhandled error: %v\n%s", err, debug.Stack())
		}
	}()

	fs, err := NewReleaseSet(d)
	if err != nil {
		return err
	}

	if err := ReadReleaseSet(newContext(d), fs, d); err != nil {
		return fmt.Errorf("reading release set: %w", err)
	}

	return nil
}

func resourceReleaseSetDiff(d *schema.ResourceDiff, meta interface{}) (finalErr error) {
	defer func() {
		if err := recover(); err != nil {
			finalErr = fmt.Errorf("unhandled error: %v\n%s", err, debug.Stack())
		}
	}()

	old, new := d.GetChange(KeyWorkingDirectory)
	log.Printf("Getting old and new working directories for id %q: old = %v, new = %v, got = %v", d.Id(), old, new, d.Get(KeyWorkingDirectory))

	fs, err := NewReleaseSet(d)
	if err != nil {
		return err
	}

	kubeconfig, err := getKubeconfig(fs)
	if err != nil {
		return fmt.Errorf("getting kubeconfig: %w", err)
	}

	if fs.Kubeconfig == "" {
		logf("Skipping helmfile-diff due to that kubeconfig is empty, which means that this operation has been called on a helmfile resource that depends on in-existent resource")

		return nil
	}

	if v, err := shouldDiff(fs); err != nil {
		return xerrors.Errorf("checking skip_diff_on_missing_files to determine if the provider needs to run helmfile-diff: %w", err)
	} else if !v {
		logf("Skipping helmfile-diff due to that one or more files listed in skip_diff_on_missing_files were missing")

		return nil
	}

	provider := meta.(*ProviderInstance)

	diff, err := DiffReleaseSet(newContext(d), fs, resourceDiffToFields(d), WithDiffConfig(DiffConfig{
		MaxDiffOutputLen: provider.MaxDiffOutputLen,
	}))
	if err != nil {
		// helmfile_release_set.kubeconfig or helmfile_releaset_set.environment_variables.KUBECONFIG can be empty
		// on `plan` if the value depends on another terraform resource.
		// This `plan` includes the implicit/automatic plan that is conducted before `terraform destroy`.
		// So, on `plan` helmfile diff can fail due to the missing KUBECONFIG. If we did return an error for that,
		// `terraform plan` or `terraform destroy` on helmfile_release_set will never succeed if the dependant resource is missing.
		if *kubeconfig != "" {
			// kubeconfig can be also empty when the kubeconfig path is static but not generated when terraform triggers
			// diff on this release_set.
			// We detect that situation by looking for the file.
			// If the kubeconfig_path is not empty AND the file is in-existent, we may safely say that
			// the path is static but the file is not yet generated.
			// In code below, `info == nil` or `os.IsNotExist(err)` means that the file is in-existent.
			if info, _ := os.Stat(*kubeconfig); info != nil {
				return fmt.Errorf("diffing release set: %w", err)
			}
		} else if !strings.Contains(err.Error(), "Kubernetes cluster unreachable") {
			return fmt.Errorf("diffing release set: %w", err)
		}
		log.Printf("Ignoring helmfile-diff error on plan because it may be due to that terraform's behaviour that "+
			"helmfile_releaset_set.kubeconfig that depends on another missing resource can be empty: %v", err)
	}

	if diff != "" {
		d.SetNewComputed(KeyApplyOutput)
	}

	return nil
}

func resourceReleaseSetUpdate(d *schema.ResourceData, meta interface{}) (finalErr error) {
	defer func() {
		if err := recover(); err != nil {
			finalErr = fmt.Errorf("unhandled error: %v\n%s", err, debug.Stack())
		}
	}()

	fs, err := NewReleaseSet(d)
	if err != nil {
		return err
	}

	return UpdateReleaseSet(newContext(d), fs, d)
}

func resourceReleaseSetDelete(d *schema.ResourceData, meta interface{}) (finalErr error) {
	defer func() {
		if err := recover(); err != nil {
			finalErr = fmt.Errorf("unhandled error: %v\n%s", err, debug.Stack())
		}
	}()

	fs, err := NewReleaseSet(d)
	if err != nil {
		return err
	}

	if err := DeleteReleaseSet(newContext(d), fs, d); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceReleaseSetImport(data *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
	data, err := ImportReleaseSet(data)
	if err != nil {
		return nil, fmt.Errorf("importing release set: %w", err)
	}

	return []*schema.ResourceData{data}, nil
}
