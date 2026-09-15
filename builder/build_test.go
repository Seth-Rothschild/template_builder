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

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPreambleDeclaresRequiredPackages(t *testing.T) {
	expected := "\\usepackage{graphicx}\n" +
		"\\usepackage{float}\n" +
		"\\usepackage{tabularray}\n" +
		"\\usepackage{xcolor}\n" +
		"\\providecommand{\\tightlist}{%\n" +
		"  \\setlength{\\itemsep}{0pt}\\setlength{\\parskip}{0pt}}\n" +
		"\\providecommand{\\pandocbounded}[1]{#1}\n" +
		"\\providecolor{accent}{HTML}{2C6E91}\n"
	actual := preamble()

	assertEqual(t, expected, actual)
}

func TestTemplateVersion(t *testing.T) {
	t.Run("returns template version from metadata", func(t *testing.T) {
		path := "testdata/slick-template.md"
		actual, err := getVersion(path)
		assertNoError(t, err)
		assertEqual(t, "slick", actual)
	})

	t.Run("returns empty when no metadata", func(t *testing.T) {
		path := "testdata/nometadata.md"
		actual, err := getVersion(path)
		assertNoError(t, err)
		assertEqual(t, "", actual)
	})
}

func TestRunPandoc(t *testing.T) {
	t.Run("returns error for nonexistent file", func(t *testing.T) {
		path := "testdata/does-not-exist.md"

		_, err := runPandoc(path)

		if !os.IsNotExist(err) {
			t.Errorf("expected a not-exist error, got %v", err)
		}
	})

	t.Run("uses default template when none given", func(t *testing.T) {
		path := "testdata/minimal.md"

		actual, err := runPandoc(path)

		assertNoError(t, err)

		if !strings.Contains(actual, "DEFAULT-TEMPLATE") {
			t.Errorf("expected output to be built from the default template, got %q", actual)
		}
	})

	t.Run("uses template named by metadata", func(t *testing.T) {
		path := "testdata/slick-template.md"

		actual, err := runPandoc(path)

		assertNoError(t, err)

		if !strings.Contains(actual, "\\setstretch{1.15}") {
			t.Errorf("expected output to be built from the slick template, got %q", actual)
		}
	})

	t.Run("gfm treats list without blank line above as list", func(t *testing.T) {
		path := "testdata/list-no-blank-line.md"

		actual, err := runPandoc(path)

		assertNoError(t, err)

		if !strings.Contains(actual, "\\begin{itemize}") {
			t.Errorf("expected list immediately following text to be treated as a list, got %q", actual)
		}
	})

	t.Run("uses tabularray instead of pandoc's longtable for tables", func(t *testing.T) {
		path := "testdata/table.md"

		actual, err := runPandoc(path)

		assertNoError(t, err)

		if !strings.Contains(actual, "\\begin{longtblr}") {
			t.Errorf("expected table to use longtblr, got %q", actual)
		}
		if strings.Contains(actual, "\\begin{longtable}") {
			t.Errorf("expected table not to use pandoc's longtable, got %q", actual)
		}
	})

	t.Run("wraps long cell text and breaks across pages", func(t *testing.T) {
		path := "testdata/llm-pros-cons.md"

		actual, err := runPandoc(path)

		assertNoError(t, err)

		if !strings.Contains(actual, "colspec = {X[l]X[l]X[l]}") {
			t.Errorf("expected columns to use wrapping X specs, got %q", actual)
		}
	})

	t.Run("includes an image found next to the markdown file", func(t *testing.T) {
		path := "testdata/image.md"

		actual, err := runPandoc(path)

		assertNoError(t, err)

		if !strings.Contains(actual, "\\includegraphics[width=\\linewidth,height=\\textheight,keepaspectratio]{image.png}") {
			t.Errorf("expected image to be included, got %q", actual)
		}
		if !strings.Contains(actual, "\\caption{a test image}") {
			t.Errorf("expected image caption, got %q", actual)
		}
	})

	t.Run("marks a missing image instead of including it", func(t *testing.T) {
		path := "testdata/missing-image.md"

		actual, err := runPandoc(path)

		assertNoError(t, err)

		if !strings.Contains(actual, "(missing image): a missing image") {
			t.Errorf("expected missing image placeholder, got %q", actual)
		}
		if strings.Contains(actual, "\\includegraphics") {
			t.Errorf("expected missing image not to be included, got %q", actual)
		}
	})
}

func TestRunTectonic(t *testing.T) {
	texPath := "testdata/sample.tex"
	outDir := t.TempDir()

	pdfPath, err := runTectonic(texPath, outDir)
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

	pdfPath, err := Build(mdPath)
	assertNoError(t, err)

	expectedPdfPath := filepath.Join(filepath.Dir(mdPath), "minimal.pdf")
	assertEqual(t, expectedPdfPath, pdfPath)
	if _, statErr := os.Stat(pdfPath); statErr != nil {
		t.Errorf("expected pdf file to exist at %q, got %v", pdfPath, statErr)
	}

	assertEqual(t, "minimal.pdf", filepath.Base(pdfPath))
}
