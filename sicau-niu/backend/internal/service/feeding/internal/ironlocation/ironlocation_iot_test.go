// ironlocation_iot_test.go covers the pure IOT locator integration logic:
// configuration validation, token/list request construction and documented
// response parsing. Database writes are covered by the refresher contract tests
// in the feeding package when PostgreSQL is enabled.

package ironlocation

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"lina-core/pkg/bizerr"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNewIOTRefresherRequiresCredentials(t *testing.T) {
	_, err := NewIOTRefresher(Config{}, nil)
	assertRuntimeCode(t, err, "")
}

func TestIOTTokenAndLocatorListRequests(t *testing.T) {
	ctx := context.Background()
	requests := make([]*http.Request, 0, 2)
	client := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		requests = append(requests, req.Clone(req.Context()))
		requests[len(requests)-1].Body = io.NopCloser(strings.NewReader(string(body)))
		switch req.URL.Path {
		case "/iot/open/device/getToken":
			if string(body) != `{"key":"key-1","secret":"secret-1"}` {
				t.Fatalf("unexpected token body: %s", body)
			}
			return jsonResponse(`{"code":200,"msg":"成功","data":{"data":"TOKEN-1"}}`), nil
		case "/iot/open/device/getInfoList":
			if req.Header.Get("Authorization") != "TOKEN-1" {
				t.Fatalf("expected Authorization TOKEN-1, got %q", req.Header.Get("Authorization"))
			}
			if string(body) != `{"page":1,"pageSize":2,"sign":"LOCATOR"}` {
				t.Fatalf("unexpected locator body: %s", body)
			}
			return jsonResponse(`{"code":200,"msg":"成功","data":{"count":1,"data":[{"locatorNo":"IRON-1","latitude":"30.12","longitude":"103.45","lastLocateTime":"2026-02-12 14:31:23"}]}}`), nil
		default:
			t.Fatalf("unexpected path %s", req.URL.Path)
			return nil, nil
		}
	})

	refresher, err := NewIOTRefresher(Config{
		BaseURL:  "https://example.test/iot",
		Key:      "key-1",
		Secret:   "secret-1",
		PageSize: 2,
		MaxPages: 1,
		TokenTTL: 72 * time.Hour,
	}, client)
	if err != nil {
		t.Fatalf("NewIOTRefresher returned error: %v", err)
	}
	iot := refresher.(*iotRefresher)

	locations, fetched, err := iot.fetchLocations(ctx, map[string]struct{}{"IRON-1": {}})
	if err != nil {
		t.Fatalf("fetchLocations returned error: %v", err)
	}
	if fetched != 1 || len(locations) != 1 {
		t.Fatalf("expected one fetched/matched location, got fetched=%d locations=%d", fetched, len(locations))
	}
	location := locations["IRON-1"]
	if location.Lat != 30.12 || location.Lng != 103.45 {
		t.Fatalf("expected Gaode coordinates 30.12/103.45, got %f/%f", location.Lat, location.Lng)
	}
	if len(requests) != 2 {
		t.Fatalf("expected token and locator requests, got %d", len(requests))
	}
}

func TestLocatorItemWithoutGaodeCoordinatesSkipped(t *testing.T) {
	// The raw payload carries only WGS-84 lat/lon; the Gaode (GCJ-02) fields are
	// empty. The item must be skipped so a WGS-84 value never enters the GCJ-02
	// iron table.
	payload := `{"locatorNo":"IRON-RAW","latitude":"","longitude":"","lat":"30.1","lon":"103.2"}`
	var item locatorItem
	if err := json.Unmarshal([]byte(payload), &item); err != nil {
		t.Fatalf("unmarshal locator payload failed: %v", err)
	}

	if location, ok := item.location(); ok {
		t.Fatalf("expected item without Gaode coordinates to be skipped, got %+v", location)
	}
}

func TestLocatorItemRejectsInvalidCoordinates(t *testing.T) {
	cases := []locatorItem{
		{LocatorNo: "IRON-NAN", Latitude: "NaN", Longitude: "103.2"},
		{LocatorNo: "IRON-INF", Latitude: "30.1", Longitude: "+Inf"},
		{LocatorNo: "IRON-LAT", Latitude: "91", Longitude: "103.2"},
		{LocatorNo: "IRON-LNG", Latitude: "30.1", Longitude: "181"},
	}

	for _, item := range cases {
		if location, ok := item.location(); ok {
			t.Fatalf("expected invalid item %+v to be rejected, got %+v", item, location)
		}
	}
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func assertRuntimeCode(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if want == "" {
		return
	}
	bizErr, ok := bizerr.As(err)
	if !ok || bizErr.RuntimeCode() != want {
		t.Fatalf("expected runtime code %s, got %T %v", want, err, err)
	}
}
