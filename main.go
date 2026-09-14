package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func port() string {
	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}
	return "8080"
}

func convertMarkdown(path string) (string, error) {
	output, err := exec.Command("pandoc", "-f", "gfm+implicit_figures", "-t", "latex", "--wrap=none", "--lua-filter=filters/tables.lua", path).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func readTemplateName(mdPath string) (string, error) {
	content, err := os.ReadFile(mdPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) < 2 || lines[0] != "---" {
		return "", nil
	}

	key, value, found := strings.Cut(lines[1], ":")
	if !found || strings.TrimSpace(key) != "template_name" {
		return "", nil
	}

	return strings.TrimSpace(value), nil
}

func preamble() string {
	return "\\usepackage{graphicx}\n" +
		"\\usepackage{tabularray}\n" +
		"\\usepackage{xcolor}\n" +
		"\\providecommand{\\tightlist}{%\n" +
		"  \\setlength{\\itemsep}{0pt}\\setlength{\\parskip}{0pt}}\n" +
		"\\providecommand{\\pandocbounded}[1]{#1}\n" +
		"\\providecolor{accent}{HTML}{2C6E91}\n"
}

func buildPDF(mdPath string, outDir string, templatePath string) (string, error) {
	if templatePath == "" {
		templatePath = "templates/default.tex"
	}

	headerPath := filepath.Join(outDir, "preamble.tex")
	if err := os.WriteFile(headerPath, []byte(preamble()), 0644); err != nil {
		return "", err
	}

	texPath := filepath.Join(outDir, "output.tex")

	pandocCmd := exec.Command("pandoc",
		"-f", "gfm+implicit_figures",
		"-t", "latex",
		"-s",
		"--template="+templatePath,
		"--include-in-header="+headerPath,
		"--wrap=none",
		"--lua-filter=filters/tables.lua",
		"-o", texPath,
		mdPath,
	)
	if output, err := pandocCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("pandoc failed: %w: %s", err, output)
	}

	cmd := exec.Command("tectonic", texPath, "--outdir", outDir, "--print")
	output, err := cmd.CombinedOutput()
	if writeErr := os.WriteFile(filepath.Join(outDir, "output.log"), output, 0644); writeErr != nil {
		return "", writeErr
	}
	if err != nil {
		return "", fmt.Errorf("tectonic failed: %w: %s", err, output)
	}

	return filepath.Join(outDir, "output.pdf"), nil
}

func saveUploadedFile(dst string, src io.Reader) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func buildHandler(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	workDir, err := os.MkdirTemp("", "build")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(workDir)

	mdPath := filepath.Join(workDir, "input.md")
	if err := saveUploadedFile(mdPath, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templateName, err := readTemplateName(mdPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var templatePath string
	if templateName != "" {
		templatePath = filepath.Join("templates", templateName+".tex")
		if _, err := os.Stat(templatePath); err != nil {
			http.Error(w, "template not found: "+templateName, http.StatusBadRequest)
			return
		}
	}

	for _, imageHeader := range r.MultipartForm.File["images"] {
		imageFile, err := imageHeader.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		imagePath := filepath.Join(workDir, imageHeader.Filename)
		err = saveUploadedFile(imagePath, imageFile)
		imageFile.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	pdfPath, err := buildPDF(mdPath, workDir, templatePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Write(pdfBytes)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to template builder!")
}

func main() {
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/build", buildHandler)

	addr := ":" + port()
	log.Println("listening on " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
