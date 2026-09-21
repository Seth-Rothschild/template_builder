package builder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

var requiredPrograms = []string{"pandoc", "tectonic"}
var requiredFiles = []string{"templates/default.tex", "filters/tables.lua", "filters/images.lua"}
var commandTimeout = 30 * time.Second

// CheckSetup verifies that the external programs and repository files Build
// depends on are available, returning a joined error listing everything
// that's missing.
func CheckSetup() error {
	problems := []error{}

	for _, program := range requiredPrograms {
		if _, err := exec.LookPath(program); err != nil {
			problem := fmt.Errorf("setup problem: %s is not installed or not on PATH", program)
			problems = append(problems, problem)
		}
	}

	workingDir, _ := os.Getwd()
	for _, path := range requiredFiles {
		if _, err := os.Stat(path); err != nil {
			problem := fmt.Errorf("setup problem: %s not found in %s, run template_builder from the repository root", path, workingDir)
			problems = append(problems, problem)
		}
	}

	return errors.Join(problems...)
}

// RunCommand runs an external command with a bounded timeout, returning its
// stdout and stderr separately.
func RunCommand(name string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		return "", "", ctx.Err()
	}
	return stdout.String(), stderr.String(), err
}
