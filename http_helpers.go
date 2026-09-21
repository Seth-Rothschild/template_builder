package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func getRequestType(r *http.Request) string {
	contentType := r.Header.Get("Content-Type")
	mediaType, _, _ := strings.Cut(contentType, ";")
	return strings.TrimSpace(mediaType)
}

func writeMarkdown(dir string, markdown string) (string, error) {
	mdPath := filepath.Join(dir, "input.md")
	if err := os.WriteFile(mdPath, []byte(markdown), 0644); err != nil {
		return "", err
	}
	return mdPath, nil
}

func writeMultipartImage(dir string, header *multipart.FileHeader) error {
	image, err := header.Open()
	if err != nil {
		return err
	}
	defer image.Close()

	content, err := io.ReadAll(image)
	if err != nil {
		return err
	}

	imagePath := filepath.Join(dir, header.Filename)
	return os.WriteFile(imagePath, content, 0644)
}

func writeBase64Image(dir string, filename string, data string) error {
	content, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("image %q: %w", filename, err)
	}
	imagePath := filepath.Join(dir, filename)
	return os.WriteFile(imagePath, content, 0644)
}

func parseMultipart(r *http.Request, dir string) (string, error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return "", err
	}

	markdown := r.FormValue("markdown")
	mdPath, err := writeMarkdown(dir, markdown)
	if err != nil {
		return "", err
	}

	for _, header := range r.MultipartForm.File["images"] {
		if err := writeMultipartImage(dir, header); err != nil {
			return "", err
		}
	}

	return mdPath, nil
}

func parseJSON(r *http.Request, dir string) (string, error) {
	var body struct {
		Markdown string `json:"markdown"`
		Images   []struct {
			Filename string `json:"filename"`
			Data     string `json:"data"`
		} `json:"images"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return "", err
	}

	mdPath, err := writeMarkdown(dir, body.Markdown)
	if err != nil {
		return "", err
	}

	for _, image := range body.Images {
		err := writeBase64Image(dir, image.Filename, image.Data)
		if err != nil {
			return "", err
		}
	}

	return mdPath, nil
}

func parsePlainMarkdown(r *http.Request, dir string) (string, error) {
	markdown, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return writeMarkdown(dir, string(markdown))
}

func parseRequest(r *http.Request, dir string) (string, error) {
	switch getRequestType(r) {
	case "multipart/form-data":
		return parseMultipart(r, dir)
	case "application/json":
		return parseJSON(r, dir)
	default:
		return parsePlainMarkdown(r, dir)
	}
}
