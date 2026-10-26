// This file exercises the summary endpoint through actual plugin registration,
// management authentication and SQLite-backed service queries with isolated data.

package backend

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/plugin/pluginhost"
	v1 "lina-plugin-media/backend/api/media/v1"
	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
)

// TestMediaDashboardSummaryRoute returns exact global active counts and rejects
// requests when both host authentication and the existing fallback fail.
func TestMediaDashboardSummaryRoute(t *testing.T) {
	// g.DB caches the default driver separately from gdb.SetConfig. Isolate this
	// fixture so its temporary SQLite file cannot become later tests' cached DB.
	const childProcessEnv = "LINAPRO_MEDIA_SUMMARY_ROUTE_CHILD"
	if os.Getenv(childProcessEnv) != "1" {
		executable, err := os.Executable()
		if err != nil {
			t.Fatalf("locate route test executable: %v", err)
		}
		cmd := exec.CommandContext(t.Context(), executable,
			"-test.run=^TestMediaDashboardSummaryRoute$", "-test.count=1", "-test.timeout=45s")
		cmd.Env = append(os.Environ(), childProcessEnv+"=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated summary route regression: %v\n%s", err, output)
		}
		t.Logf("isolated summary route regression: %s", output)
		return
	}
	ctx := t.Context()
	setMediaRouteTietaMock(t, true)
	setupMediaRouteSQLite(t)
	for _, statement := range []string{
		`CREATE TABLE media_report_stream (
			stream_id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL DEFAULT '',
			node_id TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT '', close_time TIMESTAMP
		)`,
		`CREATE TABLE media_report_session (
			session_id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL DEFAULT '',
			node_id TEXT NOT NULL DEFAULT '', close_time TIMESTAMP
		)`,
	} {
		if _, err := g.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("create isolated summary schema: %v", err)
		}
	}
	var (
		authCalls       atomic.Int32
		tenancyCalls    atomic.Int32
		permissionCalls atomic.Int32
	)
	middlewares := pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware, mediaRouteTestResponse,
		mediaRouteNoOpMiddleware, mediaRouteNoOpMiddleware, mediaRouteNoOpMiddleware,
		func(r *ghttp.Request) {
			authCalls.Add(1)
			if r.Header.Get("Authorization") != "Bearer summary-host" {
				r.Response.Status = http.StatusUnauthorized
				r.Response.Write("unauthorized")
				return
			}
			r.Middleware.Next()
		},
		func(r *ghttp.Request) {
			tenancyCalls.Add(1)
			r.Middleware.Next()
		},
		func(r *ghttp.Request) {
			permissionCalls.Add(1)
			r.Middleware.Next()
		},
	)
	baseURL, shutdown := startMediaRouteTestServer(t, middlewares)
	defer shutdown()
	assertSummary := func(streams, sessions int64) {
		t.Helper()
		response := doMediaRouteRequest(t, http.MethodGet, baseURL+"/api/v1/media/dashboard/summary", "",
			map[string]string{"Authorization": "Bearer summary-host"})
		if response.status != http.StatusOK {
			t.Fatalf("summary request failed: status=%d body=%s", response.status, response.body)
		}
		var counts map[string]int64
		if err := json.Unmarshal([]byte(response.body), &counts); err != nil {
			t.Fatalf("decode summary count response: %v; body=%s", err, response.body)
		}
		want := map[string]int64{"active_stream_count": streams, "active_session_count": sessions}
		if !reflect.DeepEqual(counts, want) {
			t.Fatalf("summary must contain only exact active counts: got=%v want=%v", counts, want)
		}
	}
	assertSummary(0, 0)
	if _, err := dao.MediaReportStream.Ctx(ctx).Data([]do.MediaReportStream{
		{StreamId: "active-a", TenantId: "tenant-a", NodeId: "4", Status: "running"},
		{StreamId: "active-b", TenantId: "tenant-b", NodeId: "5", Status: "inactive"},
		{StreamId: "closed", TenantId: "tenant-c", NodeId: "6", Status: "running", CloseTime: gtime.New("2026-01-01 00:00:00")},
	}).Insert(); err != nil {
		t.Fatalf("insert stream fixtures: %v", err)
	}
	if _, err := dao.MediaReportSession.Ctx(ctx).Data([]do.MediaReportSession{
		{SessionId: "active-a", TenantId: "tenant-a", NodeId: "4"},
		{SessionId: "active-b", TenantId: "tenant-b", NodeId: "5"},
		{SessionId: "active-c", TenantId: "tenant-c", NodeId: "6"},
		{SessionId: "closed", TenantId: "tenant-d", NodeId: "7", CloseTime: gtime.New("2026-01-01 00:00:00")},
	}).Insert(); err != nil {
		t.Fatalf("insert session fixtures: %v", err)
	}
	assertSummary(2, 3)
	if _, err := dao.MediaReportStream.Ctx(ctx).Where(do.MediaReportStream{StreamId: "active-a"}).
		Data(do.MediaReportStream{CloseTime: gtime.New("2026-01-02 00:00:00")}).Update(); err != nil {
		t.Fatalf("close stream fixture: %v", err)
	}
	assertSummary(1, 3)
	if authCalls.Load() != 3 || tenancyCalls.Load() != 3 || permissionCalls.Load() != 3 {
		t.Fatalf("summary skipped management middleware: auth=%d tenancy=%d permission=%d", authCalls.Load(), tenancyCalls.Load(), permissionCalls.Load())
	}
	denied := doMediaRouteRequest(t, http.MethodGet, baseURL+"/api/v1/media/dashboard/summary", "")
	if denied.status != http.StatusUnauthorized || strings.Contains(denied.body, "active_stream_count") || strings.Contains(denied.body, "active_session_count") {
		t.Fatalf("both auth paths failing must not expose counts: status=%d body=%s", denied.status, denied.body)
	}
	if authCalls.Load() != 4 || tenancyCalls.Load() != 3 || permissionCalls.Load() != 3 {
		t.Fatal("unauthorized summary request continued through the protected host chain")
	}
	request := reflect.TypeFor[v1.GetDashboardSummaryReq]()
	meta, ok := request.FieldByName("Meta")
	if !ok || request.NumField() != 1 || meta.Tag.Get("permission") != "media:management:query" || meta.Tag.Get("method") != "get" {
		t.Fatalf("summary must remain a protected GET without query fields: %s", request)
	}
}
