package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreambleDeclaresRequiredPackages(t *testing.T) {
	expected := "\\usepackage{graphicx}\n" +
		"\\usepackage{tabularray}\n" +
		"\\usepackage{xcolor}\n" +
		"\\providecommand{\\tightlist}{%\n" +
		"  \\setlength{\\itemsep}{0pt}\\setlength{\\parskip}{0pt}}\n" +
		"\\providecommand{\\pandocbounded}[1]{#1}\n" +
		"\\providecolor{accent}{HTML}{2C6E91}\n"
	actual := preamble()

	assertEqual(t, expected, actual)
}

func TestReadTemplateNameReturnsValueFromFrontMatter(t *testing.T) {
	path := "testdata/nonexistent-template.md"

	expected := "does-not-exist"
	actual, err := readTemplateName(path)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertEqual(t, expected, actual)
}

func TestReadTemplateNameReturnsEmptyWhenNoFrontMatter(t *testing.T) {
	path := "testdata/minimal.md"

	expected := ""
	actual, err := readTemplateName(path)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertEqual(t, expected, actual)
}

func assertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestConvertMarkdownNonexistentFileReturnsError(t *testing.T) {
	path := "testdata/does-not-exist.md"

	_, err := convertMarkdown(path)

	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}

func TestConvertMarkdownReturnsLatex(t *testing.T) {
	path := "testdata/minimal.md"

	expected := "\\section{Heading}\\label{heading}\n"
	actual, err := convertMarkdown(path)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertEqual(t, expected, actual)
}

func TestConvertMarkdownPreservesUnicode(t *testing.T) {
	path := "testdata/unicode.md"

	expected := "\\section{Heading}\\label{heading}\n\nCafé naïve résumé 日本語\n"
	actual, err := convertMarkdown(path)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertEqual(t, expected, actual)
}

func TestConvertMarkdownHandlesListWithoutBlankLineAbove(t *testing.T) {
	path := "testdata/list-no-blank-line.md"

	expected := "Some text\n\n\\begin{itemize}\n\\tightlist\n\\item\n  first item\n\\item\n  second item\n\\end{itemize}\n"
	actual, err := convertMarkdown(path)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertEqual(t, expected, actual)
}

func TestConvertMarkdownEscapesBackslashes(t *testing.T) {
	path := "testdata/backslash.md"

	expected := "Use \\textbackslash newline to break lines.\n"
	actual, err := convertMarkdown(path)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertEqual(t, expected, actual)
}

func TestConvertMarkdownHandlesTable(t *testing.T) {
	path := "testdata/table.md"

	expected := "\\begin{center}\n\\begin{tblr}{\n  colspec = {ll},\n  row{1} = {bg=accent, fg=white, font=\\bfseries},\n  row{even} = {bg=gray!10},\n  hline{1,2,Z} = {solid, 0.8pt, accent},\n}\nName & Age \\\\\nAlice & 30 \\\\\nBob & 25 \\\\\n\\end{tblr}\n\\end{center}\n"
	actual, err := convertMarkdown(path)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertEqual(t, expected, actual)
}

func TestConvertMarkdownWrapsStandaloneImageInCaptionedFigure(t *testing.T) {
	path := "testdata/image.md"

	expected := "\\begin{figure}\n\\centering\n\\pandocbounded{\\includegraphics[keepaspectratio,alt={a test image}]{image.png}}\n\\caption{a test image}\n\\end{figure}\n"
	actual, err := convertMarkdown(path)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertEqual(t, expected, actual)
}

func TestConvertMarkdownUsesTabularrayForTables(t *testing.T) {
	path := "testdata/table.md"

	actual, err := convertMarkdown(path)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(actual, "\\begin{tblr}") {
		t.Errorf("expected output to contain a tblr environment, got %q", actual)
	}
}

func TestBuildPDFProducesPDFFileForListContent(t *testing.T) {
	mdPath := "testdata/list-no-blank-line.md"
	outDir := t.TempDir()

	pdfPath, err := buildPDF(mdPath, outDir, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, statErr := os.Stat(pdfPath); statErr != nil {
		t.Errorf("expected pdf file to exist at %q, got %v", pdfPath, statErr)
	}
}

func TestBuildPDFProducesPDFFileForTableContent(t *testing.T) {
	mdPath := "testdata/table.md"
	outDir := t.TempDir()

	pdfPath, err := buildPDF(mdPath, outDir, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, statErr := os.Stat(pdfPath); statErr != nil {
		t.Errorf("expected pdf file to exist at %q, got %v", pdfPath, statErr)
	}
}

func TestBuildPDFUsesGivenTemplate(t *testing.T) {
	mdPath := "testdata/minimal.md"
	outDir := t.TempDir()

	_, err := buildPDF(mdPath, outDir, "testdata/marker-template.tex")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	texBytes, err := os.ReadFile(filepath.Join(outDir, "output.tex"))
	if err != nil {
		t.Fatalf("failed to read generated tex file: %v", err)
	}

	if !strings.Contains(string(texBytes), "MARKER-TEMPLATE") {
		t.Errorf("expected tex to be built from the given template, got %q", string(texBytes))
	}
}

func TestBuildPDFProducesPDFFileWithSlickTemplate(t *testing.T) {
	mdPath := "testdata/metadata.md"
	outDir := t.TempDir()

	pdfPath, err := buildPDF(mdPath, outDir, "templates/slick.tex")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, statErr := os.Stat(pdfPath); statErr != nil {
		t.Errorf("expected pdf file to exist at %q, got %v", pdfPath, statErr)
	}
}

func TestBuildPDFSlickTemplateDoesNotFallBackFonts(t *testing.T) {
	mdPath := "testdata/metadata.md"
	outDir := t.TempDir()

	_, err := buildPDF(mdPath, outDir, "templates/slick.tex")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	logBytes, err := os.ReadFile(filepath.Join(outDir, "output.log"))
	if err != nil {
		t.Fatalf("failed to read tectonic log: %v", err)
	}

	if strings.Contains(string(logBytes), "Font Warning") {
		t.Errorf("expected no font fallback warnings, got log:\n%s", logBytes)
	}
}

func TestBuildPDFUsesMetadataForTitle(t *testing.T) {
	mdPath := "testdata/metadata.md"
	outDir := t.TempDir()

	_, err := buildPDF(mdPath, outDir, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	texBytes, err := os.ReadFile(filepath.Join(outDir, "output.tex"))
	if err != nil {
		t.Fatalf("failed to read generated tex file: %v", err)
	}
	tex := string(texBytes)

	if !strings.Contains(tex, "\\title{Test Title}") {
		t.Errorf("expected tex to contain title, got %q", tex)
	}
	if !strings.Contains(tex, "\\author{Test Author}") {
		t.Errorf("expected tex to contain author, got %q", tex)
	}
	if !strings.Contains(tex, "\\maketitle") {
		t.Errorf("expected tex to contain \\maketitle, got %q", tex)
	}
}

func TestBuildPDFProducesPDFFile(t *testing.T) {
	mdPath := "testdata/minimal.md"
	outDir := t.TempDir()

	pdfPath, err := buildPDF(mdPath, outDir, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, statErr := os.Stat(pdfPath); statErr != nil {
		t.Errorf("expected pdf file to exist at %q, got %v", pdfPath, statErr)
	}
	if filepath.Ext(pdfPath) != ".pdf" {
		t.Errorf("expected a .pdf file, got %q", pdfPath)
	}
}

func TestPortDefaultsWhenEnvVarNotSet(t *testing.T) {
	os.Unsetenv("PORT")

	expected := "8080"
	actual := port()

	assertEqual(t, expected, actual)
}

func TestPortUsesEnvVarWhenSet(t *testing.T) {
	os.Setenv("PORT", "9090")
	defer os.Unsetenv("PORT")

	expected := "9090"
	actual := port()

	assertEqual(t, expected, actual)
}

func TestBuildHandlerReturnsBadRequestWhenNoFileProvided(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/build", nil)
	recorder := httptest.NewRecorder()

	buildHandler(recorder, request)

	response := recorder.Result()
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, response.StatusCode)
	}
}

func TestBuildHandlerReturnsBadRequestWhenTemplateNotFound(t *testing.T) {
	mdBytes, err := os.ReadFile("testdata/nonexistent-template.md")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "nonexistent-template.md")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write(mdBytes)
	writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/build", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()

	buildHandler(recorder, request)

	response := recorder.Result()
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, response.StatusCode)
	}

	expectedBody := "template not found: does-not-exist\n"
	assertEqual(t, expectedBody, recorder.Body.String())
}

func TestBuildHandlerReturnsPDF(t *testing.T) {
	mdBytes, err := os.ReadFile("testdata/minimal.md")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "minimal.md")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write(mdBytes)
	writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/build", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()

	buildHandler(recorder, request)

	response := recorder.Result()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}

	expectedContentType := "application/pdf"
	actualContentType := response.Header.Get("Content-Type")
	assertEqual(t, expectedContentType, actualContentType)

	if recorder.Body.Len() == 0 {
		t.Errorf("expected a non-empty pdf body")
	}
}

func TestBuildHandlerAcceptsMarkdownWithImage(t *testing.T) {
	mdBytes, err := os.ReadFile("testdata/image.md")
	if err != nil {
		t.Fatalf("failed to read markdown fixture: %v", err)
	}
	imageBytes, err := os.ReadFile("testdata/image.png")
	if err != nil {
		t.Fatalf("failed to read image fixture: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	mdPart, err := writer.CreateFormFile("file", "image.md")
	if err != nil {
		t.Fatalf("failed to create markdown form file: %v", err)
	}
	mdPart.Write(mdBytes)

	imagePart, err := writer.CreateFormFile("images", "image.png")
	if err != nil {
		t.Fatalf("failed to create image form file: %v", err)
	}
	imagePart.Write(imageBytes)

	writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/build", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()

	buildHandler(recorder, request)

	response := recorder.Result()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.StatusCode, recorder.Body.String())
	}

	if recorder.Body.Len() == 0 {
		t.Errorf("expected a non-empty pdf body")
	}
}

func TestHelloHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	helloHandler(recorder, request)

	response := recorder.Result()
	if response.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}

	body := recorder.Body.String()
	expectedBody := "Welcome to template builder!\n"
	if body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, body)
	}
}
