// Package builder drives Pandoc and Tectonic to turn a Markdown file into a PDF.
package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func preamble() string {
	lines := []string{
		"\\usepackage{graphicx}",
		"\\usepackage{float}",
		"\\usepackage{tikz}",
		"\\usetikzlibrary{arrows.meta, positioning, calc, shapes.geometric}",
		"\\usepackage{tabularray}",
		"\\usepackage{xcolor}",
		"\\usepackage{soul}",
		"\\usepackage{fvextra}",
		"\\usepackage{hyperref}",
		"\\fvset{breaklines=true, breakanywhere=true}",
		"\\setlength{\\emergencystretch}{3em}",
		"\\providecommand{\\tightlist}{%\n  \\setlength{\\itemsep}{0pt}\\setlength{\\parskip}{0pt}}",
		"\\providecommand{\\pandocbounded}[1]{#1}",
		"\\providecolor{accent}{HTML}{2C6E91}",
	}
	return strings.Join(lines, "\n") + "\n"
}

func inputFormat() string {
	extensions := []string{
		"+lists_without_preceding_blankline",
		"+gfm_auto_identifiers",
	}
	return "markdown" + strings.Join(extensions, "")
}

func getVersion(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) < 2 {
		return "", nil
	}
	firstLine := lines[0]
	secondLine := lines[1]
	if firstLine != "---" {
		return "", nil
	}

	key, value, _ := strings.Cut(secondLine, ":")
	if strings.TrimSpace(key) != "template_version" {
		return "", nil
	}
	version := strings.TrimSpace(value)

	return version, nil
}

func runPandoc(path string) (string, error) {
	version, err := getVersion(path)
	if err != nil {
		return "", err
	}
	template := "templates/default.tex"
	if version != "" {
		template = filepath.Join("templates", version+".tex")
	}
	if _, err := os.Stat(template); err != nil {
		return "", fmt.Errorf("could not find template %q for template_version %q", template, version)
	}

	args := []string{
		"-f", inputFormat(),
		"-t", "latex",
		"--template", template,
		"--lua-filter", "filters/tables.lua",
		"--lua-filter", "filters/images.lua",
		"-V", "header-includes=" + preamble(),
		path,
	}

	latex, stderr, err := RunCommand("pandoc", args...)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("could not run pandoc: %w", err)
	}
	if err != nil {
		return "", parsePandocError(stderr)
	}
	return latex, nil
}

func runTectonic(texPath string, outDir string) (string, error) {
	args := []string{
		texPath,
		"--outdir", outDir,
		"-Z", "search-path=templates",
	}
	_, stderr, err := RunCommand("tectonic", args...)
	if err != nil && stderr == "" {
		return "", fmt.Errorf("could not run tectonic: %w", err)
	}
	if err != nil {
		return "", parseTectonicError(stderr)
	}

	base := filepath.Base(texPath)
	pdfName := strings.TrimSuffix(base, filepath.Ext(base)) + ".pdf"
	return filepath.Join(outDir, pdfName), nil
}

// Build renders the Markdown file at mdPath into a PDF using Pandoc and
// Tectonic, and returns the path to the generated PDF.
func Build(mdPath string) (string, error) {
	latex, err := runPandoc(mdPath)
	if err != nil {
		return "", err
	}

	outDir := filepath.Dir(mdPath)
	base := filepath.Base(mdPath)
	texName := strings.TrimSuffix(base, filepath.Ext(base)) + ".tex"
	texPath := filepath.Join(outDir, texName)
	if err := os.WriteFile(texPath, []byte(latex), 0644); err != nil {
		return "", err
	}
	return runTectonic(texPath, outDir)
}
