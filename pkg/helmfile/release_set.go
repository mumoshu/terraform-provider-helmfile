package helmfile

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

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

func NewReleaseSet(d ResourceRead) (*ReleaseSet, error) {
	f := ReleaseSet{}

	// environment defaults to "helm" for helmfile_release_set but it's always nil for helmfile_release.
	// This nil-check is required to handle the latter case. Otherwise it ends up with:
	//   panic: interface conversion: interface {} is nil, not string
	if env := d.Get(KeyEnvironment); env != nil {
		f.Environment = env.(string)
	}
	// environment defaults to "" for helmfile_release_set but it's always nil for helmfile_release.
	// This nil-check is required to handle the latter case. Otherwise it ends up with:
	//   panic: interface conversion: interface {} is nil, not string
	if path := d.Get(KeyPath); path != nil {
		f.Path = path.(string)
	}

	if content := d.Get(KeyContent); content != nil {
		f.Content = content.(string)
	}

	f.DiffOutput = d.Get(KeyDiffOutput).(string)
	f.ApplyOutput = d.Get(KeyApplyOutput).(string)
	f.HelmBin = d.Get(KeyHelmBin).(string)

	if selector := d.Get(KeySelector); selector != nil {
		f.Selector = selector.(map[string]interface{})
	}

	if valuesFiles := d.Get(KeyValuesFiles); valuesFiles != nil {
		f.ValuesFiles = valuesFiles.([]interface{})
	}

	f.Values = d.Get(KeyValues).([]interface{})
	f.Bin = d.Get(KeyBin).(string)
	f.WorkingDirectory = d.Get(KeyWorkingDirectory).(string)

	logf("Printing raw working directory for %q: %s", d.Id(), f.WorkingDirectory)

	if f.Path != "" {
		if info, err := os.Stat(f.Path); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("verifying working_directory %q: %w", f.Path, err)
			}
		} else if info != nil && info.IsDir() {
			f.WorkingDirectory = f.Path
		} else {
			f.WorkingDirectory = filepath.Dir(f.Path)
		}
	}

	logf("Printing computed working directory for %q: %s", d.Id(), f.WorkingDirectory)

	if environmentVariables := d.Get(KeyEnvironmentVariables); environmentVariables != nil {
		f.EnvironmentVariables = environmentVariables.(map[string]interface{})
	}

	if concurrency := d.Get(KeyConcurrency); concurrency != nil {
		f.Concurrency = concurrency.(int)
	}
	return &f, nil
}

func NewCommand(fs *ReleaseSet, args ...string) (*exec.Cmd, error) {
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

	logf("Running helmfile %s on %+v", strings.Join(args, " "), *fs)

	flags := []string{
		"--environment", fs.Environment,
		"--file", path,
		"--helm-binary", fs.HelmBin,
		"--no-color",
	}
	for k, v := range fs.Selector {
		flags = append(flags, "--selector", fmt.Sprintf("%s=%s", k, v))
	}
	for _, f := range fs.ValuesFiles {
		flags = append(flags, "--state-values-file", fmt.Sprintf("%v", f))
	}
	for _, vs := range fs.Values {
		js := []byte(fmt.Sprintf("%s", vs))
		first := sha256.New()
		first.Write(js)
		tmpf := fmt.Sprintf("temp.values-%x.yaml", first.Sum(nil))
		if err := ioutil.WriteFile(filepath.Join(fs.WorkingDirectory, tmpf), js, 0700); err != nil {
			return nil, err
		}
		flags = append(flags, "--state-values-file", tmpf)
	}
	cmd := exec.Command(fs.Bin, append(flags, args...)...)
	cmd.Dir = fs.WorkingDirectory
	cmd.Env = append(os.Environ(), readEnvironmentVariables(fs.EnvironmentVariables)...)

	logf("[DEBUG] Generated command: wd = %s, args = %s", fs.WorkingDirectory, strings.Join(cmd.Args, " "))
	return cmd, nil
}

func CreateReleaseSet(fs *ReleaseSet, d ResourceReadWrite) error {
	logf("[DEBUG] Creating release set resource...")

	args := []string{
		"apply",
		"--concurrency", strconv.Itoa(fs.Concurrency),
		"--suppress-secrets",
	}

	cmd, err := NewCommand(fs, args...)
	if err != nil {
		return err
	}
	//obtain exclusive lock
	mutexKV.Lock(fs.WorkingDirectory)
	defer mutexKV.Unlock(fs.WorkingDirectory)

	state := NewState()
	st, err := runCommand(cmd, state, false)
	if err != nil {
		return err
	}

	d.Set(KeyApplyOutput, st.Output)
	//SetDiffOutput(d, "")

	return nil
}

func ReadReleaseSet(fs *ReleaseSet, d ResourceReadWrite) error {
	logf("[DEBUG] Reading release set resource...")

	// We treat diff_output as always empty, to show `helmfile diff` output as a complete diff,
	// rather than a diff of diffs.
	//
	// `terraform plan` shows diff on diff_output between the value after Read and CustomizeDiff.
	// So we set it empty here, in terraform resource's Read,
	// and set it non-empty later, in terraform resource's CustomizeDiff.
	// This way, terraform shows the diff between an empty string and non-empty string(full helmfile diff output),
	// which gives us what we want.
	//
	// Note that just emptying diff_output on storing it to the terraform state in StateFunc doesn't work.
	// StateFunc is called after Read and CustomizeDiff, which results in terraform showing diff of
	// an empty string against an empty string, which is ovbiously not what we want.
	d.Set(KeyDiffOutput, "")
	d.Set(KeyApplyOutput, "")

	// We run `helmfile build` against the state BEFORE the planned change,
	// to make sure any error in helmfile.yaml before successful apply is shown to the user.
	_, err := runBuild(fs)
	if err != nil {
		logf("[DEBUG] Build error detected: %v", err)

		return nil
	}

	//d.Set(KeyDiffOutput, state.Output)

	return nil
}

func runBuild(fs *ReleaseSet) (*State, error) {
	args := []string{
		"build",
	}

	cmd, err := NewCommand(fs, args...)
	if err != nil {
		return nil, err
	}

	//obtain exclusive lock
	mutexKV.Lock(fs.WorkingDirectory)
	defer mutexKV.Unlock(fs.WorkingDirectory)

	state := NewState()
	return runCommand(cmd, state, true)
}

func runTemplate(fs *ReleaseSet) (*State, error) {
	args := []string{
		"template",
	}

	cmd, err := NewCommand(fs, args...)
	if err != nil {
		return nil, err
	}

	//obtain exclusive lock
	mutexKV.Lock(fs.WorkingDirectory)
	defer mutexKV.Unlock(fs.WorkingDirectory)

	state := NewState()
	return runCommand(cmd, state, false)
}

type DiffConfig struct {
	DryRun     bool
	Kubeconfig string
}

type DiffOption func(*DiffConfig)

func WithDiffConfig(c DiffConfig) DiffOption {
	return func(p *DiffConfig) {
		*p = c
	}
}

func runDiff(fs *ReleaseSet, opts ...DiffOption) (*State, error) {
	var options DiffConfig
	for _, o := range opts {
		o(&options)
	}

	args := []string{
		"diff",
		"--concurrency", strconv.Itoa(fs.Concurrency),
		"--detailed-exitcode",
		"--suppress-secrets",
		"--context", "3",
	}

	if options.DryRun {
		args = append(args, "--dry-run")
	}

	cmd, err := NewCommand(fs, args...)
	if err != nil {
		return nil, err
	}

	if options.Kubeconfig != "" {
		cmd.Env = append(cmd.Env, "KUBECONFIG="+options.Kubeconfig)
	}

	//obtain exclusive lock
	mutexKV.Lock(fs.WorkingDirectory)
	defer mutexKV.Unlock(fs.WorkingDirectory)

	state := NewState()
	return runCommand(cmd, state, true)
}

func getDiffFile(fs *ReleaseSet) (string, error) {
	tmpl, err := runTemplate(fs)
	if err != nil {
		return "", err
	}

	determinisitcOutput, err := removeNondeterministicLogLines(tmpl.Output)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	hash.Write([]byte(determinisitcOutput))
	bs, err := json.Marshal(fs)
	if err != nil {
		return "", err
	}
	hash.Write(bs)
	diffFile := filepath.Join(".terraform", "helmfile", fmt.Sprintf("diff-%x", hash.Sum(nil)))

	return diffFile, nil
}

func writeDiffFile(fs *ReleaseSet, content string) error {
	diffFile, err := getDiffFile(fs)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(diffFile), 0755); err != nil {
		return fmt.Errorf("creating directory for diff file %s: %v", diffFile, err)
	}

	logf("Writing diff file to %s", diffFile)

	if err := ioutil.WriteFile(diffFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing diff to %s: %v", diffFile, err)
	}

	return nil
}

func readDiffFile(fs *ReleaseSet) (string, error) {
	diffFile, err := getDiffFile(fs)
	if err != nil {
		return "", err
	}

	bs, err := ioutil.ReadFile(diffFile)
	if err != nil {
		return "", err
	}

	if len(bs) > 0 {
		logf("[DEBUG] Skipped running helmfile-diff on resource because we already have changes on diff: %+v", *fs)
	}

	return string(bs), nil
}

// DiffReleaseSet detects diff to be included in the terraform plan by runnning `helmfile diff`.
// Beware that this function MUST be idempotent and the result is reliable.
//
// `terraform apply` seem to run diff twice, and if this function emitted a result different than the first run results in
// errors like:
//
//   When expanding the plan for helmfile_release_set.mystack to include new values
//   learned so far during apply, provider "registry.terraform.io/-/helmfile"
//   produced an invalid new value for .diff_output: was cty.StringVal("Adding repo
//   ...
//   a lot of text
//   ...
//   but now cty.StringVal("Adding repo stable
//   ...
//   a lot of text
//   ...
func DiffReleaseSet(fs *ReleaseSet, d ResourceReadWrite, opts ...DiffOption) (string, error) {
	logf("[DEBUG] Detecting changes on release set resource...")

	if fs.Path != "" {
		_, err := os.Stat(fs.Path)
		if err != nil {
			return "", fmt.Errorf("verifying path %q: %w", fs.Path, err)
		}
	}

	diff, err := readDiffFile(fs)
	if err != nil {
		state, err := runDiff(fs, opts...)
		if err != nil {
			logf("[DEBUG] Diff error detected: %v", err)

			// Make sure errors due to the latest `helmfile diff` run is shown to the user
			// d.SetNew(KeyError, err.Error())

			// We return the error to stop terraform from modifying the state AND
			// let the user knows about the error.
			return "", err
		}

		// We should ideally show this like `~ diff_output = <DIFF> -> (known after apply)`,
		// but it's shown as `~ diff_output = <DIFF>`, which is counter-intuitive.
		// But I wasn't able to find any way to achieve that.
		//d.SetNew(KeyDiffOutput, state.Output)
		//d.SetNewComputed(KeyDiffOutput)

		// Show the possibly transient error to disappear after successful apply.
		//
		// Seems like SetNew(KEY, "") is equivalent to SetNewComputed(KEY), according to the result below that is obtained
		// with SetNew:
		//         ~ error                 = "/Users/c-ykuoka/go/bin/helmfile: exit status 1\nin ./helmfile-b96f019fb6b4f691ffca8269edb33ffb16cb60a20c769013049c1181ebf7ecc9.yaml: failed to read helmfile-b96f019fb6b4f691ffca8269edb33ffb16cb60a20c769013049c1181ebf7ecc9.yaml: reading document at index 1: yaml: line 2: mapping values are not allowed in this context\n" -> (known after apply)
		//d.SetNew(KeyError, "")
		//d.SetNewComputed(KeyError)

		// Mark apply output for changes to instruct the user to run `terraform apply`
		// Marking it when there's no diff output means `terraform plan` always show changes, which defeats the purpose of
		// `plan`.
		if state.Output != "" {
			diff, err = removeNondeterministicLogLines(state.Output)
			if err != nil {
				return "", err
			}

			if err := writeDiffFile(fs, diff); err != nil {
				return "", err
			}
		}
	}

	// Executing d.Set(KeyDiffOutput, "") still internally records the update to the state
	// even if d.Get(KeyDiffOutput) is already "", which breaks our acceptance test.
	// Guard against that here.
	if diff != "" {
		d.Set(KeyDiffOutput, diff)
	}

	//var previousApplyOutput string
	//if v := d.Get(KeyApplyOutput); v != nil {
	//	previousApplyOutput = v.(string)
	//}
	//
	//if diff == "" && previousApplyOutput != "" {
	//	// When the diff is empty, we should still proceed with updating the state to empty apply_output
	//	// We set apply_output to "", so that the terraform is notified that this resource needs to be updated
	//	// In `UpdateReleaseSet` func, we check if `diff_output` is empty, and then set empty string to apply_output again,
	//	// so that the `apply_output=""` in plan matches `apply_output=""` in update.
	//	d.SetNew(KeyApplyOutput, "")
	//} else if diff != "" {
	//	d.SetNewComputed(KeyApplyOutput)
	//}

	return diff, nil
}

// Until https://github.com/roboll/helmfile/pull/1383 and Helmfile v0.125.1,
// various helmfile command was running `helm repo up` to update Helm chart repositories before diff/template/apply.
// `helm repo up` seems to update repositories concnurrently, in an nondeterministic order, which makes the stdout printed by the command
// nondeterministic.
//
// This provider uses output of `helmfile template` to calculate the hash key of the helmfile-diff cache, which is used
// to make originally non-determinisitc `helmfile-diff` result to be determinisitc.
// IN `removeNondeterministicLogLines`, we remove non-deterministic part of `helm repo up`, so that the provider reliably
// work with older versions of Helmfile.
func removeNondeterministicLogLines(s string) (string, error) {
	buf := &bytes.Buffer{}
	w := bufio.NewWriter(buf)

	b := bufio.NewScanner(strings.NewReader(s))
	for b.Scan() {
		l := b.Text()
		if !strings.HasPrefix(l, "...Successfully got an update from the \"") {
			if _, err := w.WriteString(l); err != nil {
				return "", err
			}
			if _, err := w.WriteRune('\n'); err != nil {
				return "", err
			}
		}
	}
	if err := w.Flush(); err != nil {
		return "", fmt.Errorf("filtering helmfile diff output: %v", err)
	}

	return buf.String(), nil
}

func UpdateReleaseSet(fs *ReleaseSet, d ResourceReadWrite) error {
	diffFile, err := getDiffFile(fs)
	if err != nil {
		return err
	}

	defer func() {
		if _, err := os.Stat(diffFile); err == nil {
			if err := os.Remove(diffFile); err != nil {
				logf("Failed cleaning diff file: %v", err)
			}
		}
	}()

	logf("[DEBUG] Updating release set resource...")

	d.Set(KeyDirty, false)

	var plannedDiffOutput string
	if v := d.Get(KeyDiffOutput); v != nil {
		plannedDiffOutput = v.(string)
	}

	if plannedDiffOutput == "" {
		return nil
	}

	args := []string{
		"apply",
		"--concurrency", strconv.Itoa(fs.Concurrency),
		"--suppress-secrets",
	}

	cmd, err := NewCommand(fs, args...)
	if err != nil {
		return err
	}

	//obtain exclusive lock
	mutexKV.Lock(fs.WorkingDirectory)
	defer mutexKV.Unlock(fs.WorkingDirectory)

	state := NewState()
	st, err := runCommand(cmd, state, false)
	if err != nil {
		return err
	}

	d.Set(KeyApplyOutput, st.Output)

	return nil
}

func DeleteReleaseSet(fs *ReleaseSet, d ResourceReadWrite) error {
	logf("[DEBUG] Deleting release set resource...")
	cmd, err := NewCommand(fs, "destroy")
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

	//SetDiffOutput(d, "")

	return nil
}
