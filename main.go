package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"template_builder/builder"
)

func port() string {
	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}
	return "8080"
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func buildHandler(w http.ResponseWriter, r *http.Request) {
	markdown, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tempDir, err := os.MkdirTemp("", "template_builder")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	mdPath := filepath.Join(tempDir, "input.md")
	if err := os.WriteFile(mdPath, markdown, 0644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pdfPath, err := builder.Build(mdPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pdf, err := os.ReadFile(pdfPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Write(pdf)
}

func oneshot(mdPath string) {
	pdfPath, err := builder.Build(mdPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("wrote " + pdfPath)
}

func main() {
	if len(os.Args) > 1 {
		oneshot(os.Args[1])
		return
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/build", buildHandler)
	addr := ":" + port()
	log.Println("listening on " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
