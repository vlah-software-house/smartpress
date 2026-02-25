package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(3, 1*time.Second)
	defer rl.Stop()

	// First 3 requests should be allowed.
	for i := 0; i < 3; i++ {
		if !rl.allow("test-ip") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied.
	if rl.allow("test-ip") {
		t.Error("4th request should be rate-limited")
	}

	// Different IP should still be allowed.
	if !rl.allow("other-ip") {
		t.Error("different IP should be allowed")
	}
}

func TestRateLimiterWindowExpiry(t *testing.T) {
	rl := NewRateLimiter(2, 100*time.Millisecond)
	defer rl.Stop()

	// Use up the limit.
	rl.allow("test-ip")
	rl.allow("test-ip")

	if rl.allow("test-ip") {
		t.Error("should be rate-limited")
	}

	// Wait for the window to expire.
	time.Sleep(150 * time.Millisecond)

	if !rl.allow("test-ip") {
		t.Error("should be allowed after window expires")
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	rl := NewRateLimiter(2, 1*time.Second)
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First 2 requests should succeed.
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: got status %d, want 200", i+1, rr.Code)
		}
	}

	// 3rd request should be rate-limited.
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("got status %d, want 429", rr.Code)
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		xff        string
		xri        string
		remoteAddr string
		want       string
	}{
		{
			name:       "x-forwarded-for single",
			xff:        "10.0.0.1",
			remoteAddr: "192.168.1.1:1234",
			want:       "10.0.0.1",
		},
		{
			name:       "x-forwarded-for multiple",
			xff:        "10.0.0.1, 172.16.0.1, 192.168.1.1",
			remoteAddr: "192.168.1.1:1234",
			want:       "10.0.0.1",
		},
		{
			name:       "x-real-ip",
			xri:        "10.0.0.2",
			remoteAddr: "192.168.1.1:1234",
			want:       "10.0.0.2",
		},
		{
			name:       "remote addr only",
			remoteAddr: "192.168.1.1:1234",
			want:       "192.168.1.1",
		},
		{
			name:       "remote addr no port",
			remoteAddr: "192.168.1.1",
			want:       "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}
			got := clientIP(req)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	rl := NewRateLimiter(5, 50*time.Millisecond)
	defer rl.Stop()

	// Add some entries.
	rl.allow("ip1")
	rl.allow("ip2")

	// Wait for entries to expire.
	time.Sleep(100 * time.Millisecond)

	rl.cleanup()

	rl.mu.RLock()
	count := len(rl.clients)
	rl.mu.RUnlock()

	if count != 0 {
		t.Errorf("cleanup should remove expired entries, got %d", count)
	}
}

// TestRateLimiterCleanupRetainsRecentEntries verifies that cleanup keeps
// entries that still have recent (non-expired) timestamps.
func TestRateLimiterCleanupRetainsRecentEntries(t *testing.T) {
	rl := NewRateLimiter(10, 200*time.Millisecond)
	defer rl.Stop()

	// Add entries for two IPs.
	rl.allow("ip-old")
	rl.allow("ip-fresh")

	// Wait long enough for "ip-old" to expire.
	time.Sleep(250 * time.Millisecond)

	// Add a new entry for "ip-fresh" so it has a recent timestamp.
	rl.allow("ip-fresh")

	rl.cleanup()

	rl.mu.RLock()
	_, oldExists := rl.clients["ip-old"]
	_, freshExists := rl.clients["ip-fresh"]
	count := len(rl.clients)
	rl.mu.RUnlock()

	if oldExists {
		t.Error("ip-old should have been cleaned up (all timestamps expired)")
	}
	if !freshExists {
		t.Error("ip-fresh should still exist (has recent timestamp)")
	}
	if count != 1 {
		t.Errorf("expected 1 remaining client, got %d", count)
	}
}
