package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func assertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestHealthHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	healthHandler(recorder, request)

	assertEqual(t, http.StatusOK, recorder.Code)
}

func TestPort(t *testing.T) {
	t.Run("defaults when env var not set", func(t *testing.T) {
		os.Unsetenv("PORT")

		expected := "8080"
		actual := port()

		assertEqual(t, expected, actual)
	})

	t.Run("uses env var when set", func(t *testing.T) {
		os.Setenv("PORT", "9090")
		defer os.Unsetenv("PORT")

		expected := "9090"
		actual := port()

		assertEqual(t, expected, actual)
	})
}

func TestTimeout(t *testing.T) {
	t.Run("defaults when env var not set", func(t *testing.T) {
		os.Unsetenv("TIMEOUT")

		expected := 60
		actual := timeout()
		assertEqual(t, expected, actual)
	})

	t.Run("uses env var when set", func(t *testing.T) {
		os.Setenv("TIMEOUT", "45")
		defer os.Unsetenv("TIMEOUT")

		expected := 45
		actual := timeout()

		assertEqual(t, expected, actual)
	})
}

func TestBuildHandler(t *testing.T) {
	markdown := "# Heading\n"
	request := httptest.NewRequest(http.MethodPost, "/build", strings.NewReader(markdown))
	recorder := httptest.NewRecorder()

	buildHandler(recorder, request)

	assertEqual(t, http.StatusOK, recorder.Code)
	assertEqual(t, "application/pdf", recorder.Header().Get("Content-Type"))

	body := recorder.Body.Bytes()
	if len(body) < 4 || string(body[:4]) != "%PDF" {
		t.Errorf("expected response body to be a pdf, got %q", string(body[:min(len(body), 20)]))
	}
}

func TestBuildHandlerErrors(t *testing.T) {
	t.Run("returns 500 for a mistake in the document", func(t *testing.T) {
		markdown := "---\ntitle: [unclosed\n---\n\nSome body text.\n"
		request := httptest.NewRequest(http.MethodPost, "/build", strings.NewReader(markdown))
		recorder := httptest.NewRecorder()

		buildHandler(recorder, request)

		assertEqual(t, http.StatusInternalServerError, recorder.Code)
		expected := "invalid metadata in template metadata: while parsing a flow sequence:\n" +
			"did not find expected ',' or ']'\n"
		assertEqual(t, expected, recorder.Body.String())
	})

	t.Run("returns 500 when a temp directory cannot be created", func(t *testing.T) {
		t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "missing"))
		markdown := "# Heading\n"
		request := httptest.NewRequest(http.MethodPost, "/build", strings.NewReader(markdown))
		recorder := httptest.NewRecorder()

		buildHandler(recorder, request)

		assertEqual(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("returns 400 when the request cannot be parsed", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/build", errReader{})
		recorder := httptest.NewRecorder()

		buildHandler(recorder, request)

		assertEqual(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("returns 500 when the server cannot build", func(t *testing.T) {
		t.Setenv("PATH", "")
		markdown := "# Heading\n"
		request := httptest.NewRequest(http.MethodPost, "/build", strings.NewReader(markdown))
		recorder := httptest.NewRecorder()

		buildHandler(recorder, request)

		assertEqual(t, http.StatusInternalServerError, recorder.Code)
		expected := "could not run pandoc: exec: \"pandoc\": executable file not found in $PATH\n"
		assertEqual(t, expected, recorder.Body.String())
	})
}

func TestNewServer(t *testing.T) {
	t.Run("uses the configured port and timeout", func(t *testing.T) {
		t.Setenv("PORT", "9090")
		t.Setenv("TIMEOUT", "45")

		server := newServer()

		assertEqual(t, ":9090", server.Addr)
		expectedTimeout := 45 * time.Second
		assertEqual(t, expectedTimeout, server.ReadTimeout)
		assertEqual(t, expectedTimeout, server.WriteTimeout)
		assertEqual(t, expectedTimeout, server.ReadHeaderTimeout)
		assertEqual(t, expectedTimeout, server.IdleTimeout)
	})

	t.Run("registers the health and build handlers", func(t *testing.T) {
		server := newServer()
		mux := server.Handler.(*http.ServeMux)

		healthRequest := httptest.NewRequest(http.MethodGet, "/health", nil)
		registeredHealthHandler, healthPattern := mux.Handler(healthRequest)
		assertEqual(t, "/health", healthPattern)
		assertEqual(t, reflect.ValueOf(healthHandler).Pointer(), reflect.ValueOf(registeredHealthHandler).Pointer())

		buildRequest := httptest.NewRequest(http.MethodPost, "/build", nil)
		registeredBuildHandler, buildPattern := mux.Handler(buildRequest)
		assertEqual(t, "/build", buildPattern)
		assertEqual(t, reflect.ValueOf(buildHandler).Pointer(), reflect.ValueOf(registeredBuildHandler).Pointer())
	})
}
