// This file verifies complete configured-node lists through the internal route,
// its API key boundary, and removal of the redundant dashboard full-list route.

package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/plugin/pluginhost"
	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
)

// TestMediaInternalAllNodesRoute checks full results and credential isolation
// in a subprocess so the default database cache cannot retain the temporary DB.
func TestMediaInternalAllNodesRoute(t *testing.T) {
	const childProcessEnv = "LINAPRO_MEDIA_INTERNAL_NODES_CHILD"
	if os.Getenv(childProcessEnv) != "1" {
		executable, err := os.Executable()
		if err != nil {
			t.Fatalf("locate node route test executable: %v", err)
		}
		cmd := exec.CommandContext(t.Context(), executable,
			"-test.run=^TestMediaInternalAllNodesRoute$", "-test.count=1", "-test.timeout=45s")
		cmd.Env = append(os.Environ(), childProcessEnv+"=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated all-node route regression: %v\n%s", err, output)
		}
		t.Logf("isolated all-node route regression: %s", output)
		return
	}
	ctx := t.Context()
	setMediaRouteConfig(t, mediaRouteTestConfig{tietaMock: true, innerAPIKey: "node-test-key", includeInnerAPIKey: true})
	setupMediaRouteSQLite(t)
	baseURL, shutdown := startMediaRouteTestServer(t, pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware, mediaRouteTestResponse,
		mediaRouteNoOpMiddleware, mediaRouteNoOpMiddleware, mediaRouteNoOpMiddleware,
		func(r *ghttp.Request) {
			if r.Header.Get("Authorization") != "Bearer management-node-test" {
				r.Response.Status = http.StatusUnauthorized
				r.Response.Write("unauthorized")
				return
			}
			r.Middleware.Next()
		},
		mediaRouteNoOpMiddleware, mediaRouteNoOpMiddleware,
	))
	defer shutdown()
	var (
		managementHeaders = map[string]string{"Authorization": "Bearer management-node-test"}
		internalHeaders   = map[string]string{mediaInnerAPIKeyHeader: "node-test-key"}
		created           = gtime.New("2026-10-01 12:34:56")
		updated           = gtime.New("2026-10-02 13:35:57")
	)
	readAll := func(path string, headers map[string]string) []map[string]any {
		t.Helper()
		response := doMediaRouteRequest(t, http.MethodGet, baseURL+path, "", headers)
		if response.status != http.StatusOK {
			t.Fatalf("node list %s failed: status=%d body=%s", path, response.status, response.body)
		}
		var envelope map[string]json.RawMessage
		if err := json.Unmarshal([]byte(response.body), &envelope); err != nil {
			t.Fatalf("decode node list %s: %v", path, err)
		}
		if len(envelope) != 1 || envelope["list"] == nil {
			t.Fatalf("full node endpoint must expose only list, without pagination: %s", response.body)
		}
		var items []map[string]any
		if err := json.Unmarshal(envelope["list"], &items); err != nil || items == nil {
			t.Fatalf("node list must be an array, including empty results: body=%s err=%v", response.body, err)
		}
		return items
	}
	if len(readAll("/api/v1/nodes/all", internalHeaders)) != 0 {
		t.Fatal("empty configuration must return an empty array from the internal node route")
	}
	rows := make([]do.MediaNode, 0, 256)
	for offset := range 256 {
		nodeNum := 255 - offset
		row := do.MediaNode{
			NodeNum: nodeNum, Name: fmt.Sprintf("节点-%03d", nodeNum),
			QnUrl: fmt.Sprintf("https://qn-%d.example.test", nodeNum), BasicUrl: fmt.Sprintf("https://basic-%d.example.test", nodeNum),
			DnUrl: fmt.Sprintf("https://dn-%d.example.test", nodeNum), CreatorId: nodeNum + 10, UpdaterId: nodeNum + 20,
			CreateTime: created, UpdateTime: updated,
		}
		if nodeNum == 0 {
			row.UpdateTime = nil
		}
		rows = append(rows, row)
	}
	if _, err := dao.MediaNode.Ctx(ctx).Data(rows).Batch(64).Insert(); err != nil {
		t.Fatalf("insert all 256 nodes: %v", err)
	}
	internalNodes := readAll("/api/v1/nodes/all", internalHeaders)
	if len(internalNodes) != 256 {
		t.Fatalf("internal route must return all 256 configured nodes, got %d", len(internalNodes))
	}
	for nodeNum, item := range internalNodes {
		id, ok := item["id"].(float64)
		if !ok || id <= 0 {
			t.Fatalf("node %d lost generated ID: %#v", nodeNum, item)
		}
		want := map[string]any{
			"id": id, "nodeNum": float64(nodeNum), "name": fmt.Sprintf("节点-%03d", nodeNum),
			"qnUrl": fmt.Sprintf("https://qn-%d.example.test", nodeNum), "basicUrl": fmt.Sprintf("https://basic-%d.example.test", nodeNum),
			"dnUrl": fmt.Sprintf("https://dn-%d.example.test", nodeNum), "creatorId": float64(nodeNum + 10), "updaterId": float64(nodeNum + 20),
			"createTime": float64(created.TimestampMilli()), "updateTime": float64(updated.TimestampMilli()),
		}
		if nodeNum == 0 {
			want["updateTime"] = nil
		}
		if !reflect.DeepEqual(item, want) {
			t.Fatalf("node %d must retain all ten fields, ascending order and millisecond times: got=%#v want=%#v", nodeNum, item, want)
		}
	}
	paged := doMediaRouteRequest(t, http.MethodGet, baseURL+"/api/v1/media/nodes?pageNum=1&pageSize=10", "", managementHeaders)
	var page struct {
		List  []map[string]any `json:"list"`
		Total int              `json:"total"`
	}
	if err := json.Unmarshal([]byte(paged.body), &page); err != nil || paged.status != http.StatusOK || len(page.List) != 10 || page.Total != 256 {
		t.Fatalf("existing paged node route changed: status=%d body=%s err=%v", paged.status, paged.body, err)
	}
	if !reflect.DeepEqual(page.List, internalNodes[:10]) {
		t.Fatalf("internal node fields and values must match the management list: internal=%#v management=%#v", internalNodes[:10], page.List)
	}
	for _, test := range []struct {
		name    string
		headers map[string]string
	}{
		{name: "internal missing key"},
		{name: "internal wrong key", headers: map[string]string{mediaInnerAPIKeyHeader: "wrong-key"}},
		{name: "bearer cannot replace internal key", headers: managementHeaders},
	} {
		response := doMediaRouteRequest(t, http.MethodGet, baseURL+"/api/v1/nodes/all", "", test.headers)
		if response.status != http.StatusUnauthorized || strings.Contains(response.body, "\"list\"") {
			t.Fatalf("%s must reject without revealing nodes: status=%d body=%s", test.name, response.status, response.body)
		}
	}
	for _, headers := range []map[string]string{nil, internalHeaders, managementHeaders} {
		response := doMediaRouteRequest(t, http.MethodGet, baseURL+"/api/v1/media/dashboard/nodes/all", "", headers)
		if response.status != http.StatusNotFound || strings.Contains(response.body, "\"list\"") {
			t.Fatalf("removed dashboard node route must return 404: status=%d body=%s", response.status, response.body)
		}
	}
}
