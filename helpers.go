package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func readMarkdown(r *http.Request) ([]byte, error) {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return io.ReadAll(r.Body)
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	markdown := r.FormValue("markdown")
	return []byte(markdown), nil
}

func writeImages(r *http.Request, dir string) error {
	if r.MultipartForm == nil {
		return nil
	}

	for _, header := range r.MultipartForm.File["images"] {
		image, err := header.Open()
		if err != nil {
			return err
		}
		content, err := io.ReadAll(image)
		image.Close()
		if err != nil {
			return err
		}
		imagePath := filepath.Join(dir, header.Filename)
		if err := os.WriteFile(imagePath, content, 0644); err != nil {
			return err
		}
	}
	return nil
}
