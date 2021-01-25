package helmfile

import (
	"github.com/mumoshu/terraform-provider-eksctl/pkg/sdk"
	"log"
	"os/exec"
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

func runCommand(ctx *sdk.Context, cmd *exec.Cmd, state *State, diffMode bool) (*State, error) {
	res, err := ctx.Run(cmd)
	if err != nil {
		return nil, err
	}

	newState := NewState()
	if diffMode && res.ExitStatus == 0 {
		newState.Output = ""
	} else {
		newState.Output = res.Output
	}

	log.Printf("[DEBUG] helmfile command new state: \"%v\"", newState)

	return newState, nil
}
