package builder

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var pandocErrors = []struct {
	pattern     *regexp.Regexp
	explanation string
}{
	{
		pattern:     regexp.MustCompile(`(?s)\(line (\d+), column \d+\):\nYAML parse exception at line (\d+), column \d+[,:]\n(.*)`),
		explanation: "invalid metadata in template metadata",
	},
}

var tectonicErrors = []struct {
	pattern     *regexp.Regexp
	explanation string
}{
	{
		pattern:     regexp.MustCompile("File `([^']+)\\.sty' not found"),
		explanation: "could not render pdf: the LaTeX package %q could not be found. Check the \\usepackage lines in the template for a typo.",
	},
	{
		pattern:     regexp.MustCompile(`Unable to load picture or PDF file '([^']+)'`),
		explanation: "could not render pdf: could not load image %q. Make sure it is a real PNG, JPEG, or PDF file, not another format renamed to one of those extensions.",
	},
	{
		pattern:     regexp.MustCompile(`(?m)^l\.\d+ .*(\\[A-Za-z@]+)\s*$`),
		explanation: "could not render pdf: the LaTeX command %s does not exist. Check the template for a typo, or for a \\usepackage line it is missing.",
	},
	{
		pattern:     regexp.MustCompile(`(?m)^error: \S+\.tex:\d+: (.*)$`),
		explanation: "could not render pdf: %s",
	},
}

func parsePandocError(stderr string) error {
	for _, pandocError := range pandocErrors {
		parts := pandocError.pattern.FindStringSubmatch(stderr)
		if parts != nil {
			return fmt.Errorf("%s: %s", pandocError.explanation, strings.TrimSpace(parts[3]))
		}
	}
	return errors.New(stderr)
}

func parseTectonicError(stderr string) error {
	for _, tectonicError := range tectonicErrors {
		parts := tectonicError.pattern.FindStringSubmatch(stderr)
		if parts != nil {
			return fmt.Errorf(tectonicError.explanation, parts[1])
		}
	}
	return errors.New(stderr)
}
