package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	err := os.Chdir("..")
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func copyFile(src string, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0644)
}

func assertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPandocConvertMarkdownToLatex(t *testing.T) {
	t.Run("returns error for nonexistent file", func(t *testing.T) {
		path := "testdata/does-not-exist.md"
		template := ""

		_, err := pandocConvertMarkdownToLatex(path, template)

		assertError(t, err)
		if !os.IsNotExist(err) {
			t.Errorf("expected a not-exist error, got %v", err)
		}
	})

	t.Run("uses default template when none given", func(t *testing.T) {
		path := "testdata/minimal.md"
		template := ""

		actual, err := pandocConvertMarkdownToLatex(path, template)

		assertNoError(t, err)

		if !strings.Contains(actual, "DEFAULT-TEMPLATE") {
			t.Errorf("expected output to be built from the default template, got %q", actual)
		}
	})

	t.Run("gfm treats list without blank line above as list", func(t *testing.T) {
		path := "testdata/list-no-blank-line.md"
		template := ""

		actual, err := pandocConvertMarkdownToLatex(path, template)

		assertNoError(t, err)

		if !strings.Contains(actual, "\\begin{itemize}") {
			t.Errorf("expected list immediately following text to be treated as a list, got %q", actual)
		}
	})

	t.Run("uses tabularray instead of longtable for tables", func(t *testing.T) {
		path := "testdata/table.md"
		template := ""

		actual, err := pandocConvertMarkdownToLatex(path, template)

		assertNoError(t, err)

		if !strings.Contains(actual, "\\begin{tblr}") {
			t.Errorf("expected table to use tblr, got %q", actual)
		}
		if strings.Contains(actual, "\\begin{longtable}") {
			t.Errorf("expected table not to use longtable, got %q", actual)
		}
	})
}

func TestPreambleDeclaresRequiredPackages(t *testing.T) {
	expected := "\\usepackage{graphicx}\n" +
		"\\usepackage{tabularray}\n" +
		"\\usepackage{xcolor}\n" +
		"\\providecommand{\\tightlist}{%\n" +
		"  \\setlength{\\itemsep}{0pt}\\setlength{\\parskip}{0pt}}\n" +
		"\\providecommand{\\pandocbounded}[1]{#1}\n" +
		"\\providecolor{accent}{HTML}{2C6E91}\n" +
		"\\usepackage{xeCJK}\n" +
		"\\setCJKmainfont{FandolHei-Regular.otf}\n"
	actual := preamble()

	assertEqual(t, expected, actual)
}

func TestParseTemplateVersion(t *testing.T) {
	t.Run("returns template version from metadata", func(t *testing.T) {
		path := "testdata/slick-template.md"
		actual, err := parseTemplateVersion(path)
		assertNoError(t, err)
		assertEqual(t, "slick", actual)
	})

	t.Run("returns empty when no metadata", func(t *testing.T) {
		path := "testdata/nometadata.md"
		actual, err := parseTemplateVersion(path)
		assertNoError(t, err)
		assertEqual(t, "", actual)
	})
}

func TestBuildLatex(t *testing.T) {
	texPath := "testdata/sample.tex"
	outDir := t.TempDir()

	pdfPath, err := buildLatex(texPath, outDir)
	assertNoError(t, err)

	if _, statErr := os.Stat(pdfPath); statErr != nil {
		t.Errorf("expected pdf file to exist at %q, got %v", pdfPath, statErr)
	}
}

func TestBuild(t *testing.T) {
	mdPath := filepath.Join(t.TempDir(), "minimal.md")
	if err := copyFile("testdata/minimal.md", mdPath); err != nil {
		t.Fatal(err)
	}

	pdfPath, err := Build(mdPath, "")
	assertNoError(t, err)

	expectedPdfPath := filepath.Join(filepath.Dir(mdPath), "minimal.pdf")
	assertEqual(t, expectedPdfPath, pdfPath)
	if _, statErr := os.Stat(pdfPath); statErr != nil {
		t.Errorf("expected pdf file to exist at %q, got %v", pdfPath, statErr)
	}

	assertEqual(t, "minimal.pdf", filepath.Base(pdfPath))
}
