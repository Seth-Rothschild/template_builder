package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
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
