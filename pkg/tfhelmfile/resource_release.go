package tfhelmfile

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const KeyNamespace = "namespace"
const KeyName = "name"
const KeyChart = "chart"
const KeyVersion = "version"
const KeyVerify = "verify"
const KeyWait = "wait"
const KeyForce = "force"
const KeyAtomic = "atomic"
const KeyCleanupOnFail = "cleanup_on_fail"
const KeyTimeout = "timeout"
const KeyKubecontext = "kubecontext"
const KeyKubeconfig = "kubeconfig"

func resourceShellHelmfileRelease() *schema.Resource {
	return &schema.Resource{
		Create:        resourceReleaseCreate,
		Delete:        resourceReleaseDelete,
		Read:          resourceReleaseRead,
		Update:        resourceReleaseUpdate,
		CustomizeDiff: resourceReleaseDiff,
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
				Optional: true,
				ForceNew: false,
				Default:  "",
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
func resourceReleaseCreate(d *schema.ResourceData, meta interface{}) error {
	rs, err := mustReadReleasetSetForRelease(d)
	if err != nil {
		return err
	}
	return createRs(rs, d, meta, []string{"create"})
}

func resourceReleaseRead(d *schema.ResourceData, meta interface{}) error {
	rs, err := mustReadReleasetSetForRelease(d)
	if err != nil {
		return err
	}
	return readRs(rs, d, meta, []string{"read"})
}

func resourceReleaseUpdate(d *schema.ResourceData, meta interface{}) error {
	rs, err := mustReadReleasetSetForRelease(d)
	if err != nil {
		return err
	}
	return updateRs(rs, d, meta, []string{"update"})
}

func resourceReleaseDiff(d *schema.ResourceDiff, meta interface{}) error {
	rs, err := mustReadReleasetSetForRelease(d)
	if err != nil {
		return err
	}
	return diffRs(rs, d, meta)
}

func resourceReleaseDelete(d *schema.ResourceData, meta interface{}) error {
	rs, err := mustReadReleasetSetForRelease(d)
	if err != nil {
		return err
	}
	return deleteRs(rs, d, meta, []string{"delete"})
}

type Release struct {
	Name             string
	Namespace        string
	Chart            string
	Version          string
	Values           []interface{}
	WorkingDirectory string
	Verify           bool
	Wait             bool
	Force            bool
	Atomic           bool
	CleanupOnFail    bool
	Timeout          int
	Kubeconfig       string
	Kubecontext      string
	Bin              string
	HelmBin          string
	DiffOutput       string
	ApplyOutput      string
}

func mustReadReleasetSetForRelease(d resource) (*ReleaseSet, error) {
	return generateHelmfileYaml(mustReadRelease(d))
}

type resource interface {
	Get(string) interface{}
	Id() string
}

func mustReadRelease(d resource) *Release {
	f := Release{}
	f.Namespace = d.Get(KeyNamespace).(string)
	f.Name = d.Get(KeyName).(string)
	if f.Name == "" {
		f.Name = d.Id()
	}
	f.Chart = d.Get(KeyChart).(string)
	f.Version = d.Get(KeyVersion).(string)
	f.Values = d.Get(KeyValues).([]interface{})
	f.WorkingDirectory = d.Get(KeyWorkingDirectory).(string)
	f.Verify = d.Get(KeyVerify).(bool)
	f.Wait = d.Get(KeyWait).(bool)
	f.Force = d.Get(KeyForce).(bool)
	f.Atomic = d.Get(KeyAtomic).(bool)
	f.CleanupOnFail = d.Get(KeyCleanupOnFail).(bool)
	f.Timeout = d.Get(KeyTimeout).(int)
	f.Kubeconfig = d.Get(KeyKubeconfig).(string)
	f.Kubecontext = d.Get(KeyKubecontext).(string)
	f.Bin = d.Get(KeyBin).(string)
	f.HelmBin = d.Get(KeyHelmBin).(string)
	f.DiffOutput = d.Get(KeyDiffOutput).(string)
	f.ApplyOutput = d.Get(KeyApplyOutput).(string)
	return &f
}

func generateHelmfileYaml(r *Release) (*ReleaseSet, error) {
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
