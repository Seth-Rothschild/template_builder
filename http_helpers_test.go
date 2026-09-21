package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", substr)
	}
	if !strings.Contains(err.Error(), substr) {
		t.Errorf("expected error containing %q, got %q", substr, err.Error())
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errors.New("boom")
}

func newMultipartRequest(t *testing.T, markdown string, images map[string][]byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("markdown", markdown); err != nil {
		t.Fatalf("writing markdown field: %v", err)
	}

	for filename, content := range images {
		part, err := writer.CreateFormFile("images", filename)
		if err != nil {
			t.Fatalf("creating form file: %v", err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatalf("writing form file content: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("closing multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/build", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func multipartFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()
	request := newMultipartRequest(t, "", map[string][]byte{filename: content})
	if err := request.ParseMultipartForm(32 << 20); err != nil {
		t.Fatalf("parsing multipart form: %v", err)
	}
	return request.MultipartForm.File["images"][0]
}

func newJSONRequest(t *testing.T, markdown string, images map[string][]byte) *http.Request {
	t.Helper()
	type image struct {
		Filename string `json:"filename"`
		Data     string `json:"data"`
	}
	body := struct {
		Markdown string  `json:"markdown"`
		Images   []image `json:"images"`
	}{Markdown: markdown}

	for filename, content := range images {
		body.Images = append(body.Images, image{
			Filename: filename,
			Data:     base64.StdEncoding.EncodeToString(content),
		})
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshaling json body: %v", err)
	}
	return httptest.NewRequest(http.MethodPost, "/build", bytes.NewReader(encoded))
}

func TestGetRequestType(t *testing.T) {
	t.Run("returns the media type from the content type header", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/build", nil)
		request.Header.Set("Content-Type", "application/json; charset=utf-8")

		assertEqual(t, "application/json", getRequestType(request))
	})

	t.Run("returns an empty string when the header is missing", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/build", nil)

		assertEqual(t, "", getRequestType(request))
	})
}

func TestWriteMarkdown(t *testing.T) {
	t.Run("writes the markdown to disk", func(t *testing.T) {
		dir := t.TempDir()

		mdPath, err := writeMarkdown(dir, "# Heading\n")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertEqual(t, filepath.Join(dir, "input.md"), mdPath)
		content, err := os.ReadFile(mdPath)
		if err != nil {
			t.Fatalf("reading written markdown: %v", err)
		}
		assertEqual(t, "# Heading\n", string(content))
	})

	t.Run("wraps the error when the file cannot be written", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "missing")

		_, err := writeMarkdown(dir, "# Heading\n")

		assertErrorContains(t, err, "writing markdown:")
	})
}

func TestWriteMultipartImage(t *testing.T) {
	t.Run("writes the uploaded image to disk", func(t *testing.T) {
		dir := t.TempDir()
		header := multipartFileHeader(t, "photo.png", []byte("fake image bytes"))

		err := writeMultipartImage(dir, header)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		content, err := os.ReadFile(filepath.Join(dir, "photo.png"))
		if err != nil {
			t.Fatalf("reading written image: %v", err)
		}
		assertEqual(t, "fake image bytes", string(content))
	})

	t.Run("wraps the error when the image cannot be written", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "missing")
		header := multipartFileHeader(t, "photo.png", []byte("fake image bytes"))

		err := writeMultipartImage(dir, header)

		assertErrorContains(t, err, `writing image "photo.png":`)
	})

	t.Run("wraps the error when the image cannot be opened", func(t *testing.T) {
		dir := t.TempDir()
		header := &multipart.FileHeader{Filename: "photo.png"}

		err := writeMultipartImage(dir, header)

		assertErrorContains(t, err, `opening image "photo.png":`)
	})
}

func TestWriteBase64Image(t *testing.T) {
	t.Run("writes the decoded image to disk", func(t *testing.T) {
		dir := t.TempDir()
		data := base64.StdEncoding.EncodeToString([]byte("fake image bytes"))

		err := writeBase64Image(dir, "photo.png", data)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		content, err := os.ReadFile(filepath.Join(dir, "photo.png"))
		if err != nil {
			t.Fatalf("reading written image: %v", err)
		}
		assertEqual(t, "fake image bytes", string(content))
	})

	t.Run("wraps the error when the data is not valid base64", func(t *testing.T) {
		dir := t.TempDir()

		err := writeBase64Image(dir, "photo.png", "!!!")

		assertErrorContains(t, err, `decoding image "photo.png":`)
	})

	t.Run("wraps the error when the image cannot be written", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "missing")
		data := base64.StdEncoding.EncodeToString([]byte("fake image bytes"))

		err := writeBase64Image(dir, "photo.png", data)

		assertErrorContains(t, err, `writing image "photo.png":`)
	})
}

func TestParseMultipart(t *testing.T) {
	t.Run("writes markdown and images from the form", func(t *testing.T) {
		dir := t.TempDir()
		request := newMultipartRequest(t, "# Heading\n", map[string][]byte{"photo.png": []byte("fake image bytes")})

		mdPath, err := parseMultipart(request, dir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		markdown, err := os.ReadFile(mdPath)
		if err != nil {
			t.Fatalf("reading written markdown: %v", err)
		}
		assertEqual(t, "# Heading\n", string(markdown))
		image, err := os.ReadFile(filepath.Join(dir, "photo.png"))
		if err != nil {
			t.Fatalf("reading written image: %v", err)
		}
		assertEqual(t, "fake image bytes", string(image))
	})

	t.Run("wraps the error when the form cannot be parsed", func(t *testing.T) {
		dir := t.TempDir()
		request := httptest.NewRequest(http.MethodPost, "/build", strings.NewReader(""))
		request.Header.Set("Content-Type", "multipart/form-data")

		_, err := parseMultipart(request, dir)

		assertErrorContains(t, err, "parsing multipart form:")
	})

	t.Run("returns the error when the markdown cannot be written", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "missing")
		request := newMultipartRequest(t, "# Heading\n", nil)

		_, err := parseMultipart(request, dir)

		assertErrorContains(t, err, "writing markdown:")
	})

	t.Run("returns the error when an image cannot be written", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "photo.png"), 0755); err != nil {
			t.Fatalf("creating colliding directory: %v", err)
		}
		request := newMultipartRequest(t, "# Heading\n", map[string][]byte{"photo.png": []byte("fake image bytes")})

		_, err := parseMultipart(request, dir)

		assertErrorContains(t, err, `writing image "photo.png":`)
	})
}

func TestParseJSON(t *testing.T) {
	t.Run("writes markdown and images from the body", func(t *testing.T) {
		dir := t.TempDir()
		request := newJSONRequest(t, "# Heading\n", map[string][]byte{"photo.png": []byte("fake image bytes")})

		mdPath, err := parseJSON(request, dir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		markdown, err := os.ReadFile(mdPath)
		if err != nil {
			t.Fatalf("reading written markdown: %v", err)
		}
		assertEqual(t, "# Heading\n", string(markdown))
		image, err := os.ReadFile(filepath.Join(dir, "photo.png"))
		if err != nil {
			t.Fatalf("reading written image: %v", err)
		}
		assertEqual(t, "fake image bytes", string(image))
	})

	t.Run("wraps the error when the body is not valid json", func(t *testing.T) {
		dir := t.TempDir()
		request := httptest.NewRequest(http.MethodPost, "/build", strings.NewReader("{"))

		_, err := parseJSON(request, dir)

		assertErrorContains(t, err, "decoding json body:")
	})

	t.Run("returns the error when the markdown cannot be written", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "missing")
		request := newJSONRequest(t, "# Heading\n", nil)

		_, err := parseJSON(request, dir)

		assertErrorContains(t, err, "writing markdown:")
	})

	t.Run("returns the error when an image is not valid base64", func(t *testing.T) {
		dir := t.TempDir()
		body := `{"markdown":"# Heading\n","images":[{"filename":"photo.png","data":"!!!"}]}`
		request := httptest.NewRequest(http.MethodPost, "/build", strings.NewReader(body))

		_, err := parseJSON(request, dir)

		assertErrorContains(t, err, `decoding image "photo.png":`)
	})
}

func TestParsePlainMarkdown(t *testing.T) {
	t.Run("writes the body to disk as markdown", func(t *testing.T) {
		dir := t.TempDir()
		request := httptest.NewRequest(http.MethodPost, "/build", strings.NewReader("# Heading\n"))

		mdPath, err := parsePlainMarkdown(request, dir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		markdown, err := os.ReadFile(mdPath)
		if err != nil {
			t.Fatalf("reading written markdown: %v", err)
		}
		assertEqual(t, "# Heading\n", string(markdown))
	})

	t.Run("wraps the error when the body cannot be read", func(t *testing.T) {
		dir := t.TempDir()
		request := httptest.NewRequest(http.MethodPost, "/build", errReader{})

		_, err := parsePlainMarkdown(request, dir)

		assertErrorContains(t, err, "reading request body:")
	})

	t.Run("returns the error when the markdown cannot be written", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "missing")
		request := httptest.NewRequest(http.MethodPost, "/build", strings.NewReader("# Heading\n"))

		_, err := parsePlainMarkdown(request, dir)

		assertErrorContains(t, err, "writing markdown:")
	})
}

func TestParseRequest(t *testing.T) {
	t.Run("dispatches multipart requests to parseMultipart", func(t *testing.T) {
		dir := t.TempDir()
		request := newMultipartRequest(t, "# Heading\n", nil)

		mdPath, err := parseRequest(request, dir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		markdown, err := os.ReadFile(mdPath)
		if err != nil {
			t.Fatalf("reading written markdown: %v", err)
		}
		assertEqual(t, "# Heading\n", string(markdown))
	})

	t.Run("dispatches json requests to parseJSON", func(t *testing.T) {
		dir := t.TempDir()
		request := newJSONRequest(t, "# Heading\n", nil)
		request.Header.Set("Content-Type", "application/json")

		mdPath, err := parseRequest(request, dir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		markdown, err := os.ReadFile(mdPath)
		if err != nil {
			t.Fatalf("reading written markdown: %v", err)
		}
		assertEqual(t, "# Heading\n", string(markdown))
	})

	t.Run("dispatches everything else to parsePlainMarkdown", func(t *testing.T) {
		dir := t.TempDir()
		request := httptest.NewRequest(http.MethodPost, "/build", strings.NewReader("# Heading\n"))

		mdPath, err := parseRequest(request, dir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		markdown, err := os.ReadFile(mdPath)
		if err != nil {
			t.Fatalf("reading written markdown: %v", err)
		}
		assertEqual(t, "# Heading\n", string(markdown))
	})
}
