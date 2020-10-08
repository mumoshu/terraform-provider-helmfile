package helmfile

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/rs/xid"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const KeyNamespace = "namespace"
const KeyName = "name"
const KeyChart = "chart"
const KeyVersion = "version"
const KeyHelmVersion = "helm_version"
const KeyHelmDiffVersion = "helm_diff_version"
const KeyVerify = "verify"
const KeyWait = "wait"
const KeyForce = "force"
const KeyAtomic = "atomic"
const KeyCleanupOnFail = "cleanup_on_fail"
const KeyTimeout = "timeout"
const KeyKubecontext = "kubecontext"
const KeyKubeconfig = "kubeconfig"

func resourceHelmfileRelease() *schema.Resource {
	return &schema.Resource{
		Create:        resourceHelmfileReleaseCreate,
		Delete:        resourceHelmfileReleaseDelete,
		Read:          resourceHelmfileReleaseRead,
		Update:        resourceHelmfileReleaseUpdate,
		CustomizeDiff: resourceHelmfileReleaseDiff,
		Schema: map[string]*schema.Schema{
			KeyNamespace: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "default",
			},
			KeyName: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			KeyChart: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
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
			KeyValues: {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			KeyWorkingDirectory: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  "",
			},
			KeyVerify: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			KeyWait: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			KeyForce: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			KeyAtomic: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			KeyCleanupOnFail: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			KeyTimeout: {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			KeyKubeconfig: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			KeyKubecontext: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  "",
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
		},
	}
}

//helpers to unwravel the recursive bits by adding a base condition
func resourceHelmfileReleaseCreate(d *schema.ResourceData, meta interface{}) error {
	rs, err := NewReleaseSetWithSingleRelease(d)
	if err != nil {
		return err
	}
	if err := CreateReleaseSet(rs, d); err != nil {
		return err
	}

	d.MarkNewResource()

	//create random uuid for the id
	id := xid.New().String()
	d.SetId(id)

	return nil
}

func resourceHelmfileReleaseRead(d *schema.ResourceData, meta interface{}) error {
	rs, err := NewReleaseSetWithSingleRelease(d)
	if err != nil {
		return err
	}

	return ReadReleaseSet(rs, d)
}

func resourceHelmfileReleaseUpdate(d *schema.ResourceData, meta interface{}) error {
	rs, err := NewReleaseSetWithSingleRelease(d)
	if err != nil {
		return err
	}

	return UpdateReleaseSet(rs, d)
}

func resourceHelmfileReleaseDiff(d *schema.ResourceDiff, meta interface{}) error {
	rs, err := NewReleaseSetWithSingleRelease(d)
	if err != nil {
		return err
	}

	diff, err := DiffReleaseSet(rs, resourceDiffToFields(d))
	if err != nil {
		return err
	}

	if diff != "" {
		d.SetNewComputed(KeyApplyOutput)
	}

	return nil
}

func resourceHelmfileReleaseDelete(d *schema.ResourceData, meta interface{}) error {
	rs, err := NewReleaseSetWithSingleRelease(d)
	if err != nil {
		return err
	}

	if err := DeleteReleaseSet(rs, d); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func NewReleaseSetWithSingleRelease(d ResourceRead) (*ReleaseSet, error) {
	r := NewRelease(d)

	var values []interface{}
	for _, v := range r.Values {
		var vv map[string]interface{}
		if err := json.Unmarshal([]byte(fmt.Sprintf("%s", v)), &vv); err != nil {
			return nil, err
		}
		values = append(values, vv)
	}
	content := map[string]interface{}{
		"releases": []interface{}{
			map[string]interface{}{
				"namespace":     r.Namespace,
				"name":          r.Name,
				"chart":         r.Chart,
				"version":       r.Version,
				"values":        values,
				"verify":        r.Verify,
				"wait":          r.Wait,
				"force":         r.Force,
				"atomic":        r.Atomic,
				"cleanupOnFail": r.CleanupOnFail,
				"timeout":       r.Timeout,
				"kubeContext":   r.Kubecontext,
			},
		},
	}
	bs, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	first := sha256.New()
	first.Write(bs)
	path := fmt.Sprintf("helmfile-%x.yaml", first.Sum(nil))
	if err := ioutil.WriteFile(path, bs, 0755); err != nil {
		return nil, err
	}

	rs := &ReleaseSet{
		Bin:              r.Bin,
		HelmBin:          r.HelmBin,
		Path:             path,
		Environment:      "default",
		WorkingDirectory: r.WorkingDirectory,
		Kubeconfig:       r.Kubeconfig,
	}

	return rs, nil
}
