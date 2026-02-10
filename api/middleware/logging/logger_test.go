package logging

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware_Success(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrappedHandler := LoggingMiddleware(handler)

	req := httptest.NewRequest("GET", "/weather?city=London", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	logOutput := buf.String()

	assert.Contains(t, logOutput, "GET", "Log should contain HTTP method 'GET'")
	assert.Contains(t, logOutput, "/weather?city=London", "Log should contain path '/weather?city=London'")
	assert.Contains(t, logOutput, "200", "Log should contain status code '200'")
}

func TestLoggingMiddleware_NotFound(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	})

	wrappedHandler := LoggingMiddleware(handler)

	req := httptest.NewRequest("GET", "/unknown", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	logOutput := buf.String()

	assert.Contains(t, logOutput, "404", "Log should contain status code '404'")
}

func TestLoggingMiddleware_POST(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Created"))
	})

	wrappedHandler := LoggingMiddleware(handler)

	req := httptest.NewRequest("POST", "/weather", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	logOutput := buf.String()

	assert.Contains(t, logOutput, "POST", "Log should contain HTTP method 'POST'")
	assert.Contains(t, logOutput, "201", "Log should contain status code '201'")
}

func TestLoggingMiddleware_RateLimitExceeded(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Rate limit exceeded"))
	})

	wrappedHandler := LoggingMiddleware(handler)

	req := httptest.NewRequest("GET", "/weather?city=Paris", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	logOutput := buf.String()

	assert.Contains(t, logOutput, "429", "Log should contain status code '429'")
}
