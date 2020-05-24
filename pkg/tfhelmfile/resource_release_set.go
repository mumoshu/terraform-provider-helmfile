package tfhelmfile

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rs/xid"
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
const KeyApplyOutput = "apply_output"
const KeyDirty = "dirty"
const KeyConcurrency = "concurrency"

const HelmfileDefaultPath = "helmfile.yaml"

func resourceShellHelmfileReleaseSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceReleaseSetCreate,
		Delete: resourceReleaseSetDelete,
		Read:   resourceReleaseSetRead,
		Update: resourceReleaseSetUpdate,
		Schema: map[string]*schema.Schema{
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
				Default:  ".",
			},
			KeyPath: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  HelmfileDefaultPath,
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
				Optional: true,
				// So that we can set this in `read` to instruct `terraform plan` to show diff as being disappear on `terraform apply`
				Computed: false,
			},
			KeyApplyOutput: {
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
		},
	}
}

//helpers to unwravel the recursive bits by adding a base condition
func resourceReleaseSetCreate(d *schema.ResourceData, meta interface{}) error {
	return create(d, meta, []string{"create"})
}

func resourceReleaseSetRead(d *schema.ResourceData, meta interface{}) error {
	return read(d, meta, []string{"read"})
}

func resourceReleaseSetUpdate(d *schema.ResourceData, meta interface{}) error {
	return update(d, meta, []string{"update"})
}

func resourceReleaseSetDelete(d *schema.ResourceData, meta interface{}) error {
	return delete(d, meta, []string{"delete"})
}

type ReleaseSet struct {
	Bin                  string
	Values               []interface{}
	ValuesFiles          []interface{}
	HelmBin              string
	Path                 string
	Content              string
	DiffOutput           string
	ApplyOutput          string
	Environment          string
	Selector             map[string]interface{}
	EnvironmentVariables map[string]interface{}
	WorkingDirectory     string
	Kubeconfig           string
	Concurrency          int
}

func MustRead(d *schema.ResourceData) *ReleaseSet {
	f := ReleaseSet{}
	f.Environment = d.Get(KeyEnvironment).(string)
	f.Path = d.Get(KeyPath).(string)
	f.Content = d.Get(KeyContent).(string)
	f.DiffOutput = d.Get(KeyDiffOutput).(string)
	f.ApplyOutput = d.Get(KeyApplyOutput).(string)
	f.HelmBin = d.Get(KeyHelmBin).(string)
	f.Selector = d.Get(KeySelector).(map[string]interface{})
	f.ValuesFiles = d.Get(KeyValuesFiles).([]interface{})
	f.Values = d.Get(KeyValues).([]interface{})
	f.Bin = d.Get(KeyBin).(string)
	f.WorkingDirectory = d.Get(KeyWorkingDirectory).(string)
	f.EnvironmentVariables = d.Get(KeyEnvironmentVariables).(map[string]interface{})
	f.Concurrency = d.Get(KeyConcurrency).(int)
	return &f
}

func SetDiffOutput(d *schema.ResourceData, v string) {
	d.Set(KeyDiffOutput, v)
}

func SetApplyOutput(d *schema.ResourceData, v string) {
	d.Set(KeyApplyOutput, v)
}

func GenerateCommand(fs *ReleaseSet, additionals ...string) (*exec.Cmd, error) {
	if fs.Content != "" && fs.Path != "" && fs.Path != HelmfileDefaultPath {
		return nil, fmt.Errorf("content and path can't be specified together: content=%q, path=%q", fs.Content, fs.Path)
	}
	var path string
	if fs.Content != "" {
		bs := []byte(fs.Content)
		first := sha256.New()
		first.Write(bs)
		path := fmt.Sprintf("helmfile-%x.yaml", first.Sum(nil))
		if err := ioutil.WriteFile(path, bs, 0700); err != nil {
			return nil, err
		}
	} else {
		path = fs.Path
	}
	args := []string{
		"--environment", fs.Environment,
		"--file", path,
		"--helm-binary", fs.HelmBin,
	}
	for k, v := range fs.Selector {
		args = append(args, "--selector", fmt.Sprintf("%s=%s", k, v))
	}
	for _, f := range fs.ValuesFiles {
		args = append(args, "--state-values-file", fmt.Sprintf("%v", f))
	}
	for _, vs := range fs.Values {
		js := []byte(fmt.Sprintf("%s", vs))
		first := sha256.New()
		first.Write(js)
		tmpf := fmt.Sprintf("temp.values-%x.yaml", first.Sum(nil))
		if err := ioutil.WriteFile(tmpf, js, 0700); err != nil {
			return nil, err
		}
		args = append(args, "--state-values-file", tmpf)
	}
	cmd := exec.Command(fs.Bin, append(args, additionals...)...)
	cmd.Dir = fs.WorkingDirectory
	cmd.Env = append(os.Environ(), readEnvironmentVariables(fs.EnvironmentVariables)...)
	log.Printf("[DEBUG] Cmd: %s", strings.Join(cmd.Args, " "))
	return cmd, nil
}

func create(d *schema.ResourceData, meta interface{}, stack []string) error {
	fs := MustRead(d)
	return createRs(fs, d, meta, stack)
}

func createRs(fs *ReleaseSet, d *schema.ResourceData, meta interface{}, stack []string) error {
	log.Printf("[DEBUG] Creating release set resource...")
	printStackTrace(stack)

	args := []string{
		"apply",
		"--concurrency", strconv.Itoa(fs.Concurrency),
	}

	cmd, err := GenerateCommand(fs, args...)
	if err != nil {
		return err
	}
	d.MarkNewResource()
	//obtain exclusive lock
	helmfileMutexKV.Lock(releaseSetMutexKey)

	state := NewState()
	st, err := runCommand(cmd, state, false)
	if err != nil {
		return err
	}
	helmfileMutexKV.Unlock(releaseSetMutexKey)

	//// Assume we won't have any diff after successful apply
	//SetDiffOutput(d, "")

	//create random uuid for the id
	id := xid.New().String()
	d.SetId(id)

	SetApplyOutput(d, st.Output)
	SetDiffOutput(d, "")

	return nil
}

func read(d *schema.ResourceData, meta interface{}, stack []string) error {
	fs := MustRead(d)
	return readRs(fs, d, meta, stack)
}

func readRs(fs *ReleaseSet, d *schema.ResourceData, meta interface{}, stack []string) error {
	log.Printf("[DEBUG] Reading release set resource...")
	printStackTrace(stack)

	args := []string{
		"diff",
		"--concurrency", strconv.Itoa(fs.Concurrency),
		"--detailed-exitcode",
	}

	cmd, err := GenerateCommand(fs, args...)
	if err != nil {
		return err
	}

	//obtain exclusive lock
	helmfileMutexKV.Lock(releaseSetMutexKey)

	state := NewState()
	newState, err := runCommand(cmd, state, true)
	if err != nil {
		return err
	}
	output := newState.Output

	helmfileMutexKV.Unlock(releaseSetMutexKey)
	if newState == nil {
		log.Printf("[DEBUG] State from read operation was nil. Marking resource for deletion.")
		d.SetId("")
		return nil
	}
	log.Printf("[DEBUG] output:|%v|", output)
	log.Printf("[DEBUG] new output:|%v|", newState.Output)

	SetDiffOutput(d, output)
	SetApplyOutput(d, "")

	isStateEqual := reflect.DeepEqual(fs.DiffOutput, newState.Output)
	isNewResource := d.IsNewResource()
	isUpdatedResource := stack[0] == "update"
	if !isStateEqual && !isNewResource && !isUpdatedResource {
		log.Printf("[DEBUG] Previous state not equal to new state. Marking resource as dirty to trigger update.")
		d.Set(KeyDirty, true)
		return nil
	}

	return nil
}

func update(d *schema.ResourceData, meta interface{}, stack []string) error {
	fs := MustRead(d)
	return updateRs(fs, d, meta, stack)
}

func updateRs(fs *ReleaseSet, d *schema.ResourceData, meta interface{}, stack []string) error {
	log.Printf("[DEBUG] Updating release set resource...")
	d.Set(KeyDirty, false)
	printStackTrace(stack)

	args := []string{
		"apply",
		"--concurrency", strconv.Itoa(fs.Concurrency),
	}

	cmd, err := GenerateCommand(fs, args...)
	if err != nil {
		return err
	}

	//obtain exclusive lock
	helmfileMutexKV.Lock(releaseSetMutexKey)

	state := NewState()
	st, err := runCommand(cmd, state, false)
	if err != nil {
		return err
	}

	SetApplyOutput(d, st.Output)

	helmfileMutexKV.Unlock(releaseSetMutexKey)

	//if err := read(d, meta, stack); err != nil {
	//	return err
	//}
	//
	SetDiffOutput(d, "")
	SetApplyOutput(d, st.Output)

	return nil
}

func delete(d *schema.ResourceData, meta interface{}, stack []string) error {
	fs := MustRead(d)
	return deleteRs(fs, d, meta, stack)
}

func deleteRs(fs *ReleaseSet, d *schema.ResourceData, meta interface{}, stack []string) error {
	log.Printf("[DEBUG] Deleting release set resource...")
	printStackTrace(stack)
	cmd, err := GenerateCommand(fs, "destroy")
	if err != nil {
		return err
	}

	//obtain exclusive lock
	helmfileMutexKV.Lock(releaseSetMutexKey)
	defer helmfileMutexKV.Unlock(releaseSetMutexKey)

	state := NewState()
	_, err = runCommand(cmd, state, false)
	if err != nil {
		return err
	}

	SetDiffOutput(d, "")

	d.SetId("")

	return nil
}
