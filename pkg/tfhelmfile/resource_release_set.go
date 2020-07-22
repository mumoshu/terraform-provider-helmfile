package tfhelmfile

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
const KeyError = "error"
const KeyApplyOutput = "apply_output"
const KeyDirty = "dirty"
const KeyConcurrency = "concurrency"

const HelmfileDefaultPath = "helmfile.yaml"

func resourceShellHelmfileReleaseSet() *schema.Resource {
	return &schema.Resource{
		Create:        resourceReleaseSetCreate,
		Delete:        resourceReleaseSetDelete,
		Read:          resourceReleaseSetRead,
		Update:        resourceReleaseSetUpdate,
		CustomizeDiff: resourceReleaseSetDiff,
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

func resourceReleaseSetDiff(d *schema.ResourceDiff, meta interface{}) error {
	return diff(d, meta)
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

type ResourceFields interface {
	Id() string
	Get(string) interface{}
}

func MustRead(d ResourceFields) *ReleaseSet {
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

	log.Printf("Printing raw working directory for %q: %s", d.Id(), f.WorkingDirectory)

	if f.Path != "" {
		if info, err := os.Stat(f.Path); err != nil {
			panic(err)
		} else if info != nil && info.IsDir() {
			f.WorkingDirectory = f.Path
		} else {
			f.WorkingDirectory = filepath.Dir(f.Path)
		}
	}

	log.Printf("Printing computed working directory for %q: %s", d.Id(), f.WorkingDirectory)

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

func GenerateCommand(fs *ReleaseSet, additionalArgs ...string) (*exec.Cmd, error) {
	if fs.Content != "" && fs.Path != "" && fs.Path != HelmfileDefaultPath {
		return nil, fmt.Errorf("content and path can't be specified together: content=%q, path=%q", fs.Content, fs.Path)
	}

	if fs.WorkingDirectory != "" {
		if err := os.MkdirAll(fs.WorkingDirectory, 0755); err != nil {
			return nil, fmt.Errorf("creating working directory %q: %w", fs.WorkingDirectory, err)
		}
	}

	var path string
	if fs.Content != "" {
		bs := []byte(fs.Content)
		first := sha256.New()
		first.Write(bs)
		path = fmt.Sprintf("helmfile-%x.yaml", first.Sum(nil))
		if err := ioutil.WriteFile(filepath.Join(fs.WorkingDirectory, path), bs, 0700); err != nil {
			return nil, err
		}
	} else {
		path = fs.Path
	}

	log.Printf("Taking diff with %+v", *fs)

	args := []string{
		"--environment", fs.Environment,
		"--file", path,
		"--helm-binary", fs.HelmBin,
		"--no-color",
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
		if err := ioutil.WriteFile(filepath.Join(fs.WorkingDirectory, tmpf), js, 0700); err != nil {
			return nil, err
		}
		args = append(args, "--state-values-file", tmpf)
	}
	cmd := exec.Command(fs.Bin, append(args, additionalArgs...)...)
	cmd.Dir = fs.WorkingDirectory
	cmd.Env = append(os.Environ(), readEnvironmentVariables(fs.EnvironmentVariables)...)
	log.Printf("[DEBUG] Generated command: wd = %s, args = %s", fs.WorkingDirectory, strings.Join(cmd.Args, " "))
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
		"--suppress-secrets",
	}

	cmd, err := GenerateCommand(fs, args...)
	if err != nil {
		return err
	}
	d.MarkNewResource()
	//obtain exclusive lock
	mutexKV.Lock(fs.WorkingDirectory)
	defer mutexKV.Unlock(fs.WorkingDirectory)

	state := NewState()
	st, err := runCommand(cmd, state, false)
	if err != nil {
		return err
	}

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

func diff(d *schema.ResourceDiff, meta interface{}) error {
	old, new := d.GetChange(KeyWorkingDirectory)
	log.Printf("Getting old and new working directories for id %q: old = %v, new = %v, got = %v", d.Id(), old, new, d.Get(KeyWorkingDirectory))

	fs := MustRead(d)
	return diffRs(fs, d, meta)
}

func readRs(fs *ReleaseSet, d *schema.ResourceData, meta interface{}, stack []string) error {
	log.Printf("[DEBUG] Reading release set resource...")

	// We run `helmfile build` against the state BEFORE the planned change,
	// to make sure any error in helmfile.yaml before successful apply is shown to the user.
	_, err := runBuild(fs)
	if err != nil {
		log.Printf("[DEBUG] Build error detected: %v", err)

		d.Set(KeyError, err.Error())

		return nil
	}

	//d.Set(KeyDiffOutput, state.Output)

	return nil
}

func runBuild(fs *ReleaseSet) (*State, error) {
	args := []string{
		"build",
	}

	cmd, err := GenerateCommand(fs, args...)
	if err != nil {
		return nil, err
	}

	//obtain exclusive lock
	mutexKV.Lock(fs.WorkingDirectory)
	defer mutexKV.Unlock(fs.WorkingDirectory)

	state := NewState()
	return runCommand(cmd, state, true)
}

func runDiff(fs *ReleaseSet) (*State, error) {
	args := []string{
		"diff",
		"--concurrency", strconv.Itoa(fs.Concurrency),
		"--detailed-exitcode",
		"--suppress-secrets",
		"--context", "5",
	}

	cmd, err := GenerateCommand(fs, args...)
	if err != nil {
		return nil, err
	}

	//obtain exclusive lock
	mutexKV.Lock(fs.WorkingDirectory)
	defer mutexKV.Unlock(fs.WorkingDirectory)

	state := NewState()
	return runCommand(cmd, state, true)
}

func diffRs(fs *ReleaseSet, d *schema.ResourceDiff, meta interface{}) error {
	log.Printf("[DEBUG] Detecting changes on release set resource...")

	state, err := runDiff(fs)
	if err != nil {
		log.Printf("[DEBUG] Diff error detected: %v", err)

		// Make sure errors due to the latest `helmfile diff` run is shown to the user
		// d.SetNew(KeyError, err.Error())

		// We return the error to stop terraform from modifying the state AND
		// let the user knows about the error.
		return err
	}

	// We should ideally show this like `~ diff_output = <DIFF> -> (known after apply)`,
	// but it's shown as `~ diff_output = <DIFF>`, which is counter-intuitive.
	// But I wasn't able to find any way to achieve that.
	d.SetNew(KeyDiffOutput, state.Output)
	//d.SetNewComputed(KeyDiffOutput)

	// Show the possibly transient error to disappear after successful apply.
	//
	// Seems like SetNew(KEY, "") is equivalent to SetNewComputed(KEY), according to the result below that is obtained
	// with SetNew:
	//         ~ error                 = "/Users/c-ykuoka/go/bin/helmfile: exit status 1\nin ./helmfile-b96f019fb6b4f691ffca8269edb33ffb16cb60a20c769013049c1181ebf7ecc9.yaml: failed to read helmfile-b96f019fb6b4f691ffca8269edb33ffb16cb60a20c769013049c1181ebf7ecc9.yaml: reading document at index 1: yaml: line 2: mapping values are not allowed in this context\n" -> (known after apply)
	d.SetNew(KeyError, "")
	//d.SetNewComputed(KeyError)

	d.SetNewComputed(KeyApplyOutput)

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
		"--suppress-secrets",
	}

	cmd, err := GenerateCommand(fs, args...)
	if err != nil {
		return err
	}

	//obtain exclusive lock
	mutexKV.Lock(fs.WorkingDirectory)
	defer mutexKV.Unlock(fs.WorkingDirectory)

	state := NewState()
	st, err := runCommand(cmd, state, false)
	if err != nil {
		d.State()

		d.Set(KeyError, err.Error())
		d.Set(KeyApplyOutput, "")

		return err
	}

	SetDiffOutput(d, "")
	d.Set(KeyError, "")
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
	mutexKV.Lock(fs.WorkingDirectory)
	defer mutexKV.Unlock(fs.WorkingDirectory)

	state := NewState()
	_, err = runCommand(cmd, state, false)
	if err != nil {
		return err
	}

	SetDiffOutput(d, "")

	d.SetId("")

	return nil
}
