package builder

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

var requiredPrograms = []string{"pandoc", "tectonic"}

var requiredFiles = []string{"templates/default.tex", "filters/tables.lua", "filters/images.lua"}

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
