package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func preamble() string {
	lines := []string{}
	lines = append(lines, "\\usepackage{graphicx}")
	lines = append(lines, "\\usepackage{float}")
	lines = append(lines, "\\usepackage{tabularray}")
	lines = append(lines, "\\usepackage{xcolor}")
	lines = append(lines, "\\providecommand{\\tightlist}{%\n  \\setlength{\\itemsep}{0pt}\\setlength{\\parskip}{0pt}}")
	lines = append(lines, "\\providecommand{\\pandocbounded}[1]{#1}")
	lines = append(lines, "\\providecolor{accent}{HTML}{2C6E91}")
	return strings.Join(lines, "\n") + "\n"
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
		return "", userError("unknown template_version %q", version)
	}

	args := []string{}
	args = append(args, "-f", "gfm+implicit_figures")
	args = append(args, "-t", "latex")
	args = append(args, "--template", template)
	args = append(args, "--lua-filter", "filters/tables.lua")
	args = append(args, "--lua-filter", "filters/images.lua")
	args = append(args, "-V", "header-includes="+preamble())
	args = append(args, path)

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
	_, stderr, err := RunCommand("tectonic", texPath, "--outdir", outDir)
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
