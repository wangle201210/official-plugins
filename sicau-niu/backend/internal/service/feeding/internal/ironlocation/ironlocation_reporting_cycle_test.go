// ironlocation_reporting_cycle_test.go covers the v1.1 reporting-cycle command:
// exact token/setConfig requests, per-device success validation, failure
// propagation and deferred missing-configuration behavior.

package ironlocation

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestIOTReportingCycleUpdaterSendsTenSecondCommand(t *testing.T) {
	requestCount := 0
	client := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requestCount++
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		switch req.URL.Path {
		case "/iot/open/device/getToken":
			if string(body) != `{"key":"key-1","secret":"secret-1"}` {
				t.Fatalf("unexpected token body: %s", body)
			}
			return jsonResponse(`{"code":200,"msg":"成功","data":{"data":"TOKEN-1"}}`), nil
		case "/iot/open/device/setConfig":
			if req.Header.Get("Authorization") != "TOKEN-1" {
				t.Fatalf("expected Authorization TOKEN-1, got %q", req.Header.Get("Authorization"))
			}
			want := `{"deviceNoList":["50275156712"],"sign":"LOCATOR","configCode":"REPORTING_CYCLE_CONFIG","template":[{"params":"cycle","value":"10"}]}`
			if string(body) != want {
				t.Fatalf("unexpected setConfig body: %s", body)
			}
			return jsonResponse(`{"code":200,"msg":"成功","data":{"count":1,"data":[{"name":"50275156712","value":null,"status":0}]}}`), nil
		default:
			t.Fatalf("unexpected path %s", req.URL.Path)
			return nil, nil
		}
	})

	updater, err := NewIOTReportingCycleUpdater(Config{
		BaseURL: "https://example.test/iot",
		Key:     "key-1",
		Secret:  "secret-1",
	}, client)
	if err != nil {
		t.Fatalf("NewIOTReportingCycleUpdater returned error: %v", err)
	}
	if err = updater.SetReportingCycle(context.Background(), " 50275156712 ", 10*time.Second); err != nil {
		t.Fatalf("SetReportingCycle returned error: %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("expected token and setConfig requests, got %d", requestCount)
	}
}

func TestIOTReportingCycleUpdaterRejectsPerDeviceFailure(t *testing.T) {
	client := reportingCycleTestClient(t, `{"code":200,"msg":"成功","data":{"count":1,"data":[{"name":"50275156712","value":"设备离线","status":1}]}}`)
	updater, err := NewIOTReportingCycleUpdater(Config{
		BaseURL: "https://example.test/iot",
		Key:     "key-1",
		Secret:  "secret-1",
	}, client)
	if err != nil {
		t.Fatalf("NewIOTReportingCycleUpdater returned error: %v", err)
	}

	err = updater.SetReportingCycle(context.Background(), "50275156712", 10*time.Second)
	assertRuntimeCode(t, err, CodeIOTRequestFailed.RuntimeCode())
}

func TestIOTReportingCycleUpdaterRejectsMissingDeviceResult(t *testing.T) {
	client := reportingCycleTestClient(t, `{"code":200,"msg":"成功","data":{"count":1,"data":[{"name":"OTHER","value":null,"status":0}]}}`)
	updater, err := NewIOTReportingCycleUpdater(Config{
		BaseURL: "https://example.test/iot",
		Key:     "key-1",
		Secret:  "secret-1",
	}, client)
	if err != nil {
		t.Fatalf("NewIOTReportingCycleUpdater returned error: %v", err)
	}

	err = updater.SetReportingCycle(context.Background(), "50275156712", 10*time.Second)
	assertRuntimeCode(t, err, CodeIOTResponseInvalid.RuntimeCode())
}

func TestUnconfiguredReportingCycleUpdaterDefersFailure(t *testing.T) {
	updater := NewUnconfiguredReportingCycleUpdater()
	err := updater.SetReportingCycle(context.Background(), "50275156712", 10*time.Second)
	assertRuntimeCode(t, err, CodeIOTNotConfigured.RuntimeCode())
}

func TestNewIOTReportingCycleUpdaterRejectsNilHTTPClient(t *testing.T) {
	updater, err := NewIOTReportingCycleUpdater(Config{
		BaseURL: "https://example.test/iot",
		Key:     "key-1",
		Secret:  "secret-1",
	}, nil)
	if err == nil {
		t.Fatal("expected nil HTTP client to be rejected")
	}
	if updater != nil {
		t.Fatalf("expected nil updater, got %#v", updater)
	}
}

func TestIOTReportingCycleUpdaterRejectsInvalidCycleBeforeHTTP(t *testing.T) {
	called := false
	client := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		called = true
		return nil, nil
	})
	updater, err := NewIOTReportingCycleUpdater(Config{
		BaseURL: "https://example.test/iot",
		Key:     "key-1",
		Secret:  "secret-1",
	}, client)
	if err != nil {
		t.Fatalf("NewIOTReportingCycleUpdater returned error: %v", err)
	}

	err = updater.SetReportingCycle(context.Background(), "50275156712", 9*time.Second)
	assertRuntimeCode(t, err, CodeIOTCommandInvalid.RuntimeCode())
	if called {
		t.Fatal("expected invalid cycle to avoid HTTP calls")
	}
}

func reportingCycleTestClient(t *testing.T, setConfigResponse string) HTTPClient {
	t.Helper()
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/iot/open/device/getToken":
			return jsonResponse(`{"code":200,"msg":"成功","data":{"data":"TOKEN-1"}}`), nil
		case "/iot/open/device/setConfig":
			if req.Header.Get("Authorization") != "TOKEN-1" {
				t.Fatalf("expected Authorization TOKEN-1, got %q", req.Header.Get("Authorization"))
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("read request body: %v", err)
			}
			if !strings.Contains(string(body), `"value":"10"`) {
				t.Fatalf("expected 10-second cycle body, got %s", body)
			}
			return jsonResponse(setConfigResponse), nil
		default:
			t.Fatalf("unexpected path %s", req.URL.Path)
			return nil, nil
		}
	})
}
