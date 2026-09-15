package builder

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var yamlErrorPattern = regexp.MustCompile(`(?s)\(line (\d+), column \d+\):\nYAML parse exception at line (\d+), column \d+[,:]\n(.*)`)

var yamlExplanations = map[string]string{
	"did not find expected ',' or ']'":               `a list starting with "[" is never closed. Add the missing "]", or put the value in quotes if the "[" is part of the text.`,
	"mapping values are not allowed in this context": `a value contains a colon or is indented too far. Put the value in quotes, like title: "Part 1: Setup", or remove the extra indentation.`,
	"found unexpected document indicator":            `a quoted value is never closed. Add the missing closing quote.`,
}

var missingPackagePattern = regexp.MustCompile("File `([^']+)\\.sty' not found")

var badImagePattern = regexp.MustCompile(`Unable to load picture or PDF file '([^']+)'`)

var undefinedCommandPattern = regexp.MustCompile(`(?m)^l\.\d+ .*(\\[A-Za-z@]+)\s*$`)

var latexErrorPattern = regexp.MustCompile(`(?m)^error: \S+\.tex:\d+: (.*)$`)

func parsePandocError(stderr string) error {
	parts := yamlErrorPattern.FindStringSubmatch(stderr)
	if parts == nil {
		return errors.New(stderr)
	}

	metadataStartLine, _ := strconv.Atoi(parts[1])
	lineInMetadata, _ := strconv.Atoi(parts[2])
	line := metadataStartLine + lineInMetadata

	reason := strings.TrimSpace(parts[3])
	reason = strings.ReplaceAll(reason, "\n", " ")
	for yamlMessage, explanation := range yamlExplanations {
		if strings.Contains(reason, yamlMessage) {
			reason = explanation
		}
	}
	return userError("invalid metadata near line %d: %s", line, reason)
}

func parseTectonicError(stderr string) error {
	missingPackage := missingPackagePattern.FindStringSubmatch(stderr)
	if missingPackage != nil {
		packageName := missingPackage[1]
		return fmt.Errorf("could not render pdf: the LaTeX package %q could not be found. Check the \\usepackage lines in the template for a typo.", packageName)
	}

	badImage := badImagePattern.FindStringSubmatch(stderr)
	if badImage != nil {
		imageName := badImage[1]
		return userError("could not render pdf: could not load image %q. Make sure it is a real PNG, JPEG, or PDF file, not another format renamed to one of those extensions.", imageName)
	}

	undefinedCommand := undefinedCommandPattern.FindStringSubmatch(stderr)
	if strings.Contains(stderr, "Undefined control sequence") && undefinedCommand != nil {
		command := undefinedCommand[1]
		return fmt.Errorf("could not render pdf: the LaTeX command %s does not exist. Check the template for a typo, or for a \\usepackage line it is missing.", command)
	}

	latexError := latexErrorPattern.FindStringSubmatch(stderr)
	if latexError != nil {
		reason := latexError[1]
		return fmt.Errorf("could not render pdf: %s", reason)
	}

	return errors.New(stderr)
}

type UserError struct {
	message string
}

func (e UserError) Error() string {
	return e.message
}

func userError(format string, args ...any) error {
	message := fmt.Sprintf(format, args...)
	return UserError{message: message}
}

func IsUserError(err error) bool {
	var target UserError
	return errors.As(err, &target)
}
