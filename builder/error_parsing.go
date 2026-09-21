package builder

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type errorPattern struct {
	pattern                 *regexp.Regexp
	userFriendlyExplanation string
}

var pandocErrors = []errorPattern{
	{
		pattern:                 regexp.MustCompile(`(?s)\(line \d+, column \d+\):\nYAML parse exception at line \d+, column \d+[,:]\n(.*)`),
		userFriendlyExplanation: "invalid metadata in template metadata: %s",
	},
}

var tectonicErrors = []errorPattern{
	{
		pattern:                 regexp.MustCompile("File `([^']+)\\.sty' not found"),
		userFriendlyExplanation: "could not render pdf: the LaTeX package %q could not be found. Check the \\usepackage lines in the template for a typo.",
	},
	{
		pattern:                 regexp.MustCompile(`Unable to load picture or PDF file '([^']+)'`),
		userFriendlyExplanation: "could not render pdf: could not load image %q. Make sure it is a real PNG, JPEG, or PDF file, not another format renamed to one of those extensions.",
	},
	{
		pattern:                 regexp.MustCompile(`(?m)^l\.\d+ .*(\\[A-Za-z@]+)\s*$`),
		userFriendlyExplanation: "could not render pdf: the LaTeX command %s does not exist. Check the template for a typo, or for a \\usepackage line it is missing.",
	},
	{
		pattern:                 regexp.MustCompile(`(?m)^error: \S+\.tex:\d+: (.*)$`),
		userFriendlyExplanation: "could not render pdf: %s",
	},
}

func parseError(stderr string, patterns []errorPattern) error {
	for _, entry := range patterns {
		parts := entry.pattern.FindStringSubmatch(stderr)
		if parts != nil {
			capturedText := strings.TrimSpace(parts[1])
			return fmt.Errorf(entry.userFriendlyExplanation, capturedText)
		}
	}
	return errors.New(stderr)
}

func parsePandocError(stderr string) error {
	return parseError(stderr, pandocErrors)
}

func parseTectonicError(stderr string) error {
	return parseError(stderr, tectonicErrors)
}
