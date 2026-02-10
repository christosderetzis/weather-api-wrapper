package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupRateLimitedHandler(requestsPerMinute int) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	rateLimiter := NewRateLimiter(requestsPerMinute)
	return rateLimiter.Middleware(handler)
}

func TestRateLimiter_Allow(t *testing.T) {
	wrappedHandler := setupRateLimitedHandler(30)

	req := httptest.NewRequest("GET", "/weather?city=London", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	for i := 0; i < 30; i++ {
		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Request %d: expected status 200", i+1)
	}
}

func TestRateLimiter_Exceed(t *testing.T) {
	wrappedHandler := setupRateLimitedHandler(5)

	req := httptest.NewRequest("GET", "/weather?city=London", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Request %d: expected status 200", i+1)
	}

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code, "Expected status 429 (Too Many Requests)")
}

func TestRateLimiter_MultipleIPs(t *testing.T) {
	wrappedHandler := setupRateLimitedHandler(5)

	req1 := httptest.NewRequest("GET", "/weather?city=London", nil)
	req1.RemoteAddr = "192.168.1.1:12345"

	req2 := httptest.NewRequest("GET", "/weather?city=Paris", nil)
	req2.RemoteAddr = "192.168.1.2:12345"

	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req1)
		assert.Equal(t, http.StatusOK, rr.Code, "IP1 Request %d: expected status 200", i+1)
	}

	rr1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusTooManyRequests, rr1.Code, "IP1: Expected rate limit exceeded")

	rr2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code, "IP2: Expected status 200 (different IP should not be rate limited)")
}

func TestRateLimiter_Recovery(t *testing.T) {
	wrappedHandler := setupRateLimitedHandler(60)

	req := httptest.NewRequest("GET", "/weather?city=London", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	for i := 0; i < 60; i++ {
		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, "Request %d: expected status 200", i+1)
	}

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code, "Expected rate limit exceeded")

	time.Sleep(2 * time.Second)

	rr = httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "After waiting, expected status 200")
}
