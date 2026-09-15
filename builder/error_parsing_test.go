package builder

import (
	"os"
	"testing"
)

func readStderr(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile("testdata/stderr/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestParsePandocError(t *testing.T) {
	t.Run("explains an unclosed yaml list", func(t *testing.T) {
		stderr := readStderr(t, "pandoc-yaml-flow-sequence.txt")

		err := parsePandocError(stderr)

		expected := "invalid metadata near line 4: a list starting with \"[\" is never closed. " +
			"Add the missing \"]\", or put the value in quotes if the \"[\" is part of the text."
		assertEqual(t, expected, err.Error())
	})

	t.Run("explains an unexpected yaml colon", func(t *testing.T) {
		stderr := readStderr(t, "pandoc-yaml-mapping-value.txt")

		err := parsePandocError(stderr)

		expected := "invalid metadata near line 2: a value contains a colon or is indented too far. " +
			"Put the value in quotes, like title: \"Part 1: Setup\", or remove the extra indentation."
		assertEqual(t, expected, err.Error())
	})

	t.Run("explains an unclosed yaml quote", func(t *testing.T) {
		stderr := readStderr(t, "pandoc-yaml-unclosed-quote.txt")

		err := parsePandocError(stderr)

		expected := "invalid metadata near line 3: a quoted value is never closed. " +
			"Add the missing closing quote."
		assertEqual(t, expected, err.Error())
	})

	t.Run("passes through an unrecognized error", func(t *testing.T) {
		stderr := readStderr(t, "pandoc-unrecognized.txt")

		err := parsePandocError(stderr)

		assertEqual(t, stderr, err.Error())
	})
}

func TestParseTectonicError(t *testing.T) {
	t.Run("explains an image that cannot be loaded", func(t *testing.T) {
		stderr := readStderr(t, "tectonic-bad-image.txt")

		err := parseTectonicError(stderr)

		expected := "could not render pdf: could not load image \"bad.png\". " +
			"Make sure it is a real PNG, JPEG, or PDF file, not another format renamed to one of those extensions."
		assertEqual(t, expected, err.Error())
	})

	t.Run("explains an undefined latex command", func(t *testing.T) {
		stderr := readStderr(t, "tectonic-undefined-command.txt")

		err := parseTectonicError(stderr)

		expected := "could not render pdf: the LaTeX command \\notacommand does not exist. " +
			"Check the template for a typo, or for a \\usepackage line it is missing."
		assertEqual(t, expected, err.Error())
	})

	t.Run("explains a missing latex package", func(t *testing.T) {
		stderr := readStderr(t, "tectonic-missing-package.txt")

		err := parseTectonicError(stderr)

		expected := "could not render pdf: the LaTeX package \"notarealpackage\" could not be found. " +
			"Check the \\usepackage lines in the template for a typo."
		assertEqual(t, expected, err.Error())
	})

	t.Run("reports any other latex error without the log", func(t *testing.T) {
		stderr := "error: input.tex:5: Something else went wrong\nerror: something bad happened inside XeTeX\n"

		err := parseTectonicError(stderr)

		expected := "could not render pdf: Something else went wrong"
		assertEqual(t, expected, err.Error())
	})

	t.Run("passes through an unrecognized error", func(t *testing.T) {
		stderr := "tectonic: something unexpected\n"

		err := parseTectonicError(stderr)

		assertEqual(t, stderr, err.Error())
	})
}
