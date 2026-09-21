package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"template_builder/builder"
)

func port() string {
	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}
	return "8080"
}

func timeout() int {
	if envTimeout := os.Getenv("TIMEOUT"); envTimeout != "" {
		if t, err := strconv.Atoi(envTimeout); err == nil {
			return t
		}
	}
	return 60
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func buildHandler(w http.ResponseWriter, r *http.Request) {
	tempDir, err := os.MkdirTemp("", "template_builder")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	mdPath, err := parseRequest(r, tempDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pdfPath, err := builder.Build(mdPath)
	if builder.IsUserError(err) {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
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

func main() {
	if err := builder.CheckSetup(); err != nil {
		log.Fatal(err)
	}

	calledAsCommand := len(os.Args) > 1
	if calledAsCommand {
		pdfPath, err := builder.Build(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		log.Println("wrote " + pdfPath)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/build", buildHandler)
	addr := ":" + port()
	timeout := time.Duration(timeout()) * time.Second

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadTimeout:       timeout,
		WriteTimeout:      timeout,
		ReadHeaderTimeout: timeout,
		IdleTimeout:       timeout,
	}

	log.Println("listening on " + addr)
	log.Fatal(server.ListenAndServe())
}
