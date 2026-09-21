package builder

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCheckSetup(t *testing.T) {
	t.Run("passes in a working repository", func(t *testing.T) {
		err := CheckSetup()

		assertNoError(t, err)
	})

	t.Run("reports programs that are not installed", func(t *testing.T) {
		t.Setenv("PATH", "")

		err := CheckSetup()

		expected := "setup problem: pandoc is not installed or not on PATH\n" +
			"setup problem: tectonic is not installed or not on PATH"
		assertEqual(t, expected, err.Error())
	})

	t.Run("reports repository files that cannot be found", func(t *testing.T) {
		t.Chdir(t.TempDir())
		workingDir, _ := os.Getwd()

		err := CheckSetup()

		expected := "setup problem: templates/default.tex not found in " + workingDir + ", run template_builder from the repository root\n" +
			"setup problem: filters/tables.lua not found in " + workingDir + ", run template_builder from the repository root\n" +
			"setup problem: filters/images.lua not found in " + workingDir + ", run template_builder from the repository root"
		assertEqual(t, expected, err.Error())
	})
}

func ExampleCheckSetup() {
	if err := CheckSetup(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ready")
	// Output:
	// ready
}

func ExampleRunCommand() {
	stdout, _, err := RunCommand("pandoc", "--version")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(strings.HasPrefix(stdout, "pandoc"))
	// Output:
	// true
}

func TestRunCommandTimeout(t *testing.T) {
	commandTimeout = 0
	t.Cleanup(func() { commandTimeout = 30 * time.Second })

	_, _, err := RunCommand("pandoc", "--version")

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected a deadline exceeded error, got %v", err)
	}
}
