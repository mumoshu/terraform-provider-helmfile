package helmfile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/mitchellh/go-linereader"
)

// State is a wrapper around both the input and output attributes that are relavent for updates
type State struct {
	Output string
}

// NewState is the constructor for State
func NewState() *State {
	return &State{}
}

func readEnvironmentVariables(ev map[string]interface{}, exclude string) []string {
	var variables []string
	if ev != nil {
		for k, v := range ev {
			if k == exclude {
				continue
			}
			variables = append(variables, k+"="+v.(string))
		}
	}
	return variables
}

type outputter struct{}

func (o outputter) Output(_ string) {

}

func runCommand(cmd *exec.Cmd, state *State, diffMode bool) (*State, error) {
	const maxBufSize = 8 * 1024
	// Setup the command
	input, _ := json.Marshal(state.Output)
	stdin := bytes.NewReader(input)
	cmd.Stdin = stdin
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = pw
	cmd.Stdout = pw

	output := &bytes.Buffer{}

	// Write everything we read from the pipe to the output buffer too
	tee := io.TeeReader(pr, output)

	// copy the teed output to the UI output
	copyDoneCh := make(chan struct{})
	//o := ctx.Value(schema.ProvOutputKey).(terraform.UIOutput)
	go copyOutput(outputter{}, tee, copyDoneCh)

	// Run the command to completion
	runErr := cmd.Run()

	if err := pw.Close(); err != nil {
		return nil, err
	}

	select {
	case <-copyDoneCh:
		//case <-ctx.Done():
	}

	out := output.String()
	log.Printf("[DEBUG] helmfile command output: \"%s\"", out)
	var exitStatus int
	if runErr != nil {
		switch ee := runErr.(type) {
		case *exec.ExitError:
			// Propagate any non-zero exit status from the external command, rather than throwing it away,
			// so that helmfile could return its own exit code accordingly
			waitStatus := ee.Sys().(syscall.WaitStatus)
			exitStatus = waitStatus.ExitStatus()
			if exitStatus != 2 {
				return nil, fmt.Errorf("%s: %v\n%s", cmd.Path, runErr, out)
			}
		default:
			return nil, fmt.Errorf("%s: %v\n%s", cmd.Path, runErr, out)
		}
	}

	newState := NewState()
	if diffMode && exitStatus == 0 {
		newState.Output = ""
	} else {
		newState.Output = out
	}
	log.Printf("[DEBUG] helmfile command new state: \"%v\"", newState)
	return newState, nil
}

type Outputter interface {
	Output(string)
}

func copyOutput(o Outputter, r io.Reader, doneCh chan<- struct{}) {
	defer close(doneCh)
	lr := linereader.New(r)
	for line := range lr.Ch {
		o.Output(line)
	}
}
