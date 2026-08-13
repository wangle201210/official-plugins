// feeding_ironlocation_test.go covers the real IOT refresh handoff into feeding:
// the external locator response updates the local iron-cow table, and player
// feeding later reads only that stored coordinate through NewStoredIronLocation.

package feeding

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
)

type iotRoundTripFunc func(*http.Request) (*http.Response, error)

func (f iotRoundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestIOTRefreshUpdatesStoredIronLocationForFeeding(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLFeedingDB(t, ctx)

	if _, err := dao.Iron.Ctx(ctx).Data(do.Iron{Code: "IRON-IOT-1", Name: "IOT Iron"}).Insert(); err != nil {
		t.Fatalf("insert iron row failed: %v", err)
	}
	refresher, err := NewIOTIronLocationRefresher(IronLocationConfig{
		BaseURL:  "https://example.test/iot",
		Key:      "key-1",
		Secret:   "secret-1",
		PageSize: 10,
		MaxPages: 1,
		TokenTTL: 72 * time.Hour,
	}, iotRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		body, readErr := io.ReadAll(req.Body)
		if readErr != nil {
			t.Fatalf("read request body: %v", readErr)
		}
		switch req.URL.Path {
		case "/iot/open/device/getToken":
			if string(body) != `{"key":"key-1","secret":"secret-1"}` {
				t.Fatalf("unexpected token request body: %s", body)
			}
			return testJSONResponse(`{"code":200,"msg":"成功","data":{"data":"TOKEN-IOT"}}`), nil
		case "/iot/open/device/getInfoList":
			if got := req.Header.Get("Authorization"); got != "TOKEN-IOT" {
				t.Fatalf("expected Authorization TOKEN-IOT, got %q", got)
			}
			return testJSONResponse(`{"code":200,"msg":"成功","data":{"count":1,"data":[{"locatorNo":"IRON-IOT-1","latitude":"30.000000","longitude":"103.000000","lastLocateTime":"2026-02-12 14:31:23"}]}}`), nil
		default:
			t.Fatalf("unexpected IOT path %s", req.URL.Path)
			return nil, nil
		}
	}))
	if err != nil {
		t.Fatalf("NewIOTIronLocationRefresher returned error: %v", err)
	}

	result, err := refresher.Refresh(ctx)
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if result.Registered != 1 || result.Matched != 1 || result.Updated != 1 {
		t.Fatalf("unexpected refresh result: %+v", result)
	}

	niuID := stageActiveNiu(t, ctx, "NIU-IOT-FEED", 30.0, 103.0)
	userID := insertUserRow(t, ctx, "openid-iot-feed")
	creditGrass(t, ctx, userID, 100)
	svc := newFeedingServiceForTest(NewStoredIronLocation())

	out, err := svc.Feed(ctx, userID, &FeedInput{NiuId: niuID, BaseAmount: 10, RequestId: "req-feed-iot"})
	if err != nil {
		t.Fatalf("Feed returned error: %v", err)
	}
	if !out.IsIronBonus || out.CoefficientBasis != ironBonusCoefficientBasis || out.EffectAmount != 15 {
		t.Fatalf("expected stored IOT coordinate to trigger iron bonus, got %+v", out)
	}
}

func testJSONResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
