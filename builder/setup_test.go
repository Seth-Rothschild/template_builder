package builder

import (
	"os"
	"testing"
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
