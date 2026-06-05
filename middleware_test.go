package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRateMW verifies that a client may burst up to rateBurst requests
// before being throttled with 429, and that a different client IP is
// unaffected by another's limit.
func TestRateMW(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := rateMW(ok)

	get := func(ip string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip + ":12345"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	// Burst allowance passes, the next request is throttled.
	const ip = "192.0.2.10"
	for i := 1; i <= rateBurst; i++ {
		if code := get(ip); code != http.StatusOK {
			t.Fatalf("request %d: got %d, want %d", i, code, http.StatusOK)
		}
	}
	if code := get(ip); code != http.StatusTooManyRequests {
		t.Fatalf("request %d: got %d, want %d",
			rateBurst+1, code, http.StatusTooManyRequests)
	}

	// A separate IP has its own bucket and is not throttled.
	if code := get("192.0.2.11"); code != http.StatusOK {
		t.Fatalf("other ip: got %d, want %d", code, http.StatusOK)
	}
}
