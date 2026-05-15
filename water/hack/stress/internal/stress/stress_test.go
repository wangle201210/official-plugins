package stress

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

// Do executes the test round trip function.
func (f roundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

// testLogger captures failure log formats for assertions.
type testLogger struct {
	mu      sync.Mutex
	entries []string
}

// Printf records one failure log format string.
func (l *testLogger) Printf(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, format)
}

// TestNormalizeConfigRequiresDataOrImage verifies a request body source is required.
func TestNormalizeConfigRequiresDataOrImage(t *testing.T) {
	_, err := NormalizeConfig(Config{URL: "http://127.0.0.1:8080/api/v1/water/preview"})
	if err == nil {
		t.Fatal("expected missing data error")
	}
	if !strings.Contains(err.Error(), "data or image path is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestNormalizeConfigRejectsInvalidJSON verifies JSON request bodies are validated.
func TestNormalizeConfigRejectsInvalidJSON(t *testing.T) {
	_, err := NormalizeConfig(Config{
		URL:  "http://127.0.0.1:8080/api/v1/water/preview",
		Data: "{bad-json",
	})
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
	if !strings.Contains(err.Error(), "valid JSON") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRunAggregatesConcurrentRequests verifies concurrent workers aggregate request metrics.
func TestRunAggregatesConcurrentRequests(t *testing.T) {
	var count int
	var mu sync.Mutex
	client := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("Authorization") != "Bearer token-a" {
			t.Fatalf("expected bearer token header, got %q", req.Header.Get("Authorization"))
		}
		mu.Lock()
		count++
		mu.Unlock()
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"success":true}`)),
		}, nil
	})

	result, err := Run(context.Background(), Config{
		URL:            "http://127.0.0.1:8080/api/v1/water/preview",
		BearerToken:    "token-a",
		Data:           `{"tenant":"tenant-a","image":"x"}`,
		Concurrency:    4,
		TotalRequests:  12,
		RequestTimeout: time.Second,
	}, client, nil)
	if err != nil {
		t.Fatalf("run stress test: %v", err)
	}
	if result.TotalRequests != 12 || result.SuccessRequests != 12 || result.FailedRequests != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.AvgResponseTime <= 0 {
		t.Fatalf("expected positive average response time, got %s", result.AvgResponseTime)
	}
	mu.Lock()
	defer mu.Unlock()
	if count != 12 {
		t.Fatalf("expected 12 requests, got %d", count)
	}
}

// TestMakeRequestLogsFailedStatus verifies non-2xx responses are counted as failures.
func TestMakeRequestLogsFailedStatus(t *testing.T) {
	logger := &testLogger{}
	client := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("failed")),
		}, nil
	})
	_, success := MakeRequest(context.Background(), Config{
		URL:             "http://127.0.0.1:8080/api/v1/water/preview",
		Data:            `{"tenant":"tenant-a","image":"x"}`,
		ResponsePreview: 3,
	}, client, logger)
	if success {
		t.Fatal("expected request failure")
	}
	if len(logger.entries) != 1 {
		t.Fatalf("expected one failure log, got %d", len(logger.entries))
	}
}

// TestDefaultURLSubmitEscapesPathValues verifies submit URLs are path-safe.
func TestDefaultURLSubmitEscapesPathValues(t *testing.T) {
	got := DefaultURL("http://127.0.0.1:8080/api/v1/", ModeSubmit, "gb/type", "device id")
	want := "http://127.0.0.1:8080/api/v1/water/snaps/gb%2Ftype/device%20id"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
