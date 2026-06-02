// This file covers legacy route response adapters that wrap host-owned
// system-management routes after host authentication and handlers have run.

package backend

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/pkg/plugin/pluginhost"
)

func TestLegacyRouteInterceptorRewritesHostListResponse(t *testing.T) {
	server := startLegacyInterceptorServer(t, "role-list", true, func(group *ghttp.RouterGroup) {
		group.GET("/role", func(r *ghttp.Request) {
			if got := r.GetRequest("pageNum").String(); got != "3" {
				t.Fatalf("expected legacy pageIndex to be aliased to pageNum, got %q", got)
			}
			if got := r.GetRequest("name").String(); got != "admin" {
				t.Fatalf("expected legacy roleName to be aliased to name, got %q", got)
			}
			r.Response.WriteJson(map[string]any{
				"code":    0,
				"message": "OK",
				"data": map[string]any{
					"list":  []map[string]any{{"id": 1, "name": "admin"}},
					"total": 1,
				},
			})
		})
	})

	body := g.Client().GetContent(
		context.Background(),
		server+"/api/v1/role?pageIndex=3&pageSize=7&roleName=admin",
	)
	for _, want := range []string{
		`"code":200`,
		`"msg":"查询成功"`,
		`"count":1`,
		`"pageIndex":3`,
		`"pageSize":7`,
		`"roleId":1`,
		`"roleName":"admin"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected legacy list response to contain %s, got %s", want, body)
		}
	}
}

func TestLegacyRouteInterceptorKeepsUnauthorizedHostResponse(t *testing.T) {
	server := startLegacyInterceptorServer(t, "role-unauthorized", true, func(group *ghttp.RouterGroup) {
		group.GET("/role", func(r *ghttp.Request) {
			r.Response.WriteStatus(http.StatusUnauthorized)
		})
	})

	resp, err := g.Client().Get(context.Background(), server+"/api/v1/role")
	if err != nil {
		t.Fatalf("request legacy host route: %v", err)
	}
	defer resp.Close()
	body := resp.ReadAllString()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected host unauthorized status to remain, got %d body %s", resp.StatusCode, body)
	}
	if strings.Contains(body, `"code":200`) {
		t.Fatalf("expected unauthorized response not to be rewritten as old success, got %s", body)
	}
}

func TestLegacyRouteInterceptorAliasesLegacyJSONBody(t *testing.T) {
	server := startLegacyInterceptorServer(t, "role-create", true, func(group *ghttp.RouterGroup) {
		group.POST("/role", func(r *ghttp.Request) {
			if got := r.GetRequest("name").String(); got != "legacy-admin" {
				t.Fatalf("expected legacy roleName to be aliased to name, got %q", got)
			}
			if got := r.GetRequest("key").String(); got != "legacy_admin" {
				t.Fatalf("expected legacy roleKey to be aliased to key, got %q", got)
			}
			r.Response.WriteJson(map[string]any{
				"code": 0,
				"data": map[string]any{"id": 7},
			})
		})
	})

	body := g.Client().ContentJson().PostContent(
		context.Background(),
		server+"/api/v1/role",
		`{"roleName":"legacy-admin","roleKey":"legacy_admin"}`,
	)
	for _, want := range []string{
		`"code":200`,
		`"msg":"创建成功"`,
		`"data":7`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected legacy create response to contain %s, got %s", want, body)
		}
	}
}

func TestLegacyRouteInterceptorMapsHostMenuTreeResponse(t *testing.T) {
	server := startLegacyInterceptorServer(t, "menu-tree", true, func(group *ghttp.RouterGroup) {
		group.GET("/menu", func(r *ghttp.Request) {
			r.Response.WriteJson(map[string]any{
				"code": 0,
				"data": map[string]any{
					"list": []map[string]any{
						{
							"id":       10,
							"name":     "系统管理",
							"type":     "M",
							"perms":    "system:menu:list",
							"isCache":  0,
							"parentId": 0,
							"children": []map[string]any{{"id": 11, "name": "菜单", "type": "C"}},
						},
					},
				},
			})
		})
	})

	body := g.Client().GetContent(context.Background(), server+"/api/v1/menu")
	for _, want := range []string{
		`"menuId":10`,
		`"menuName":"系统管理"`,
		`"menuType":"M"`,
		`"permission":"system:menu:list"`,
		`"noCache":true`,
		`"menuId":11`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected legacy menu response to contain %s, got %s", want, body)
		}
	}
}

func TestLegacyRouteInterceptorMapsHostDictAndConfigResponses(t *testing.T) {
	server := startLegacyInterceptorServer(t, "dict-config", true, func(group *ghttp.RouterGroup) {
		group.GET("/dict/data", func(r *ghttp.Request) {
			if got := r.GetRequest("type").String(); got != "sys_normal_disable" {
				t.Fatalf("expected dictType query to be aliased to type, got %q", got)
			}
			r.Response.WriteJson(map[string]any{
				"code": 0,
				"data": map[string]any{
					"items": []map[string]any{{"id": 21, "label": "正常", "value": "2", "type": "sys_normal_disable"}},
					"total": 1,
				},
			})
		})
		group.GET("/config/9", func(r *ghttp.Request) {
			r.Response.WriteJson(map[string]any{
				"code": 0,
				"data": map[string]any{
					"item": map[string]any{"id": 9, "name": "CAS", "key": "cas.url", "value": "https://cas.example.com"},
				},
			})
		})
	})

	dictBody := g.Client().GetContent(context.Background(), server+"/api/v1/dict/data?dictType=sys_normal_disable")
	for _, want := range []string{
		`"dictCode":21`,
		`"dictLabel":"正常"`,
		`"dictValue":"2"`,
		`"dictType":"sys_normal_disable"`,
	} {
		if !strings.Contains(dictBody, want) {
			t.Fatalf("expected legacy dict response to contain %s, got %s", want, dictBody)
		}
	}

	configBody := g.Client().GetContent(context.Background(), server+"/api/v1/config/9")
	for _, want := range []string{
		`"id":9`,
		`"configId":9`,
		`"configName":"CAS"`,
		`"configKey":"cas.url"`,
		`"configValue":"https://cas.example.com"`,
	} {
		if !strings.Contains(configBody, want) {
			t.Fatalf("expected legacy config response to contain %s, got %s", want, configBody)
		}
	}
}

func TestLegacyRouteInterceptorMapsHostProfileResponse(t *testing.T) {
	server := startLegacyInterceptorServer(t, "profile", true, func(group *ghttp.RouterGroup) {
		group.GET("/user/profile", func(r *ghttp.Request) {
			r.Response.WriteJson(map[string]any{
				"code": 0,
				"data": map[string]any{
					"user":    map[string]any{"id": 3, "name": "管理员", "username": "admin", "dept": map[string]any{"id": 2, "name": "研发部"}},
					"roles":   []map[string]any{{"id": 1, "name": "admin", "key": "admin"}},
					"posts":   []map[string]any{{"id": 5, "name": "工程师", "code": "engineer"}},
					"roleIds": []int{1},
					"postIds": []int{5},
				},
			})
		})
	})

	body := g.Client().GetContent(context.Background(), server+"/api/v1/user/profile")
	for _, want := range []string{
		`"userId":3`,
		`"nickName":"管理员"`,
		`"deptId":2`,
		`"deptName":"研发部"`,
		`"roleId":1`,
		`"roleName":"admin"`,
		`"postId":5`,
		`"postName":"工程师"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected legacy profile response to contain %s, got %s", want, body)
		}
	}
}

func TestLegacyRouteInterceptorBypassesNonLegacyRoutes(t *testing.T) {
	server := startLegacyInterceptorServer(t, "modern-route", true, func(group *ghttp.RouterGroup) {
		group.GET("/modern", func(r *ghttp.Request) {
			r.Response.WriteJson(map[string]any{"host": true})
		})
	})

	body := g.Client().GetContent(context.Background(), server+"/api/v1/modern")
	if !strings.Contains(body, `"host":true`) {
		t.Fatalf("expected non-legacy route to pass through, got %s", body)
	}
}

func TestLegacyRouteInterceptorDisabledPluginBypassesResponseRewrite(t *testing.T) {
	server := startLegacyInterceptorServer(t, "disabled", false, func(group *ghttp.RouterGroup) {
		group.GET("/role", func(r *ghttp.Request) {
			r.Response.WriteJson(map[string]any{
				"code": 0,
				"data": map[string]any{
					"list":  []map[string]any{},
					"total": 0,
				},
			})
		})
	})

	body := g.Client().GetContent(context.Background(), server+"/api/v1/role")
	if strings.Contains(body, `"msg":"查询成功"`) {
		t.Fatalf("expected disabled plugin middleware to bypass rewrite, got %s", body)
	}
	if !strings.Contains(body, `"code":0`) {
		t.Fatalf("expected original host response to remain, got %s", body)
	}
}

func startLegacyInterceptorServer(t *testing.T, name string, enabled bool, register func(group *ghttp.RouterGroup)) string {
	t.Helper()
	server := g.Server("legacy-interceptor-" + name + "-" + guid.S())
	server.SetDumpRouterMap(false)
	server.SetPort(0)
	server.Group("/api/v1", register)
	registrar := pluginhost.NewGlobalMiddlewareRegistrar(server, pluginID, func(_ context.Context, _ string) bool {
		return enabled
	})
	if err := registerLegacyRouteInterceptors(registrar); err != nil {
		t.Fatalf("register legacy interceptor: %v", err)
	}
	server.Start()
	t.Cleanup(func() {
		server.Shutdown()
	})
	time.Sleep(100 * time.Millisecond)
	return fmt.Sprintf("http://127.0.0.1:%d", server.GetListenedPort())
}
