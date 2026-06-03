// This file verifies old host-colliding routes are dispatched to plugin-owned
// legacy handlers after bearer authentication instead of being served by host
// system-management handlers.

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
	uidentitycontroller "lina-plugin-linapro-uidentity-cas/backend/internal/controller/uidentity"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

func TestLegacyRouteInterceptorDispatchesToPluginListHandler(t *testing.T) {
	service := &legacyInterceptorFakeService{}
	hostMiddlewares := &legacyInterceptorFakeHostMiddlewares{}
	server := startLegacyInterceptorServer(t, "role-list", true, hostMiddlewares.routeMiddlewares(), service, func(group *ghttp.RouterGroup) {
		group.GET("/role", func(r *ghttp.Request) {
			t.Fatal("host /role handler must not run for old compatibility route")
		})
	})

	body := legacyInterceptorGet(t, server+"/api/v1/role?pageIndex=3&pageSize=7&roleName=admin")
	if hostMiddlewares.authHeader != "Bearer test-token" {
		t.Fatalf("expected host auth middleware to receive authorization header, got %q", hostMiddlewares.authHeader)
	}
	if hostMiddlewares.ctxCalls != 1 || hostMiddlewares.authCalls != 1 || hostMiddlewares.tenancyCalls != 1 {
		t.Fatalf("expected ctx/auth/tenancy to run once, got ctx=%d auth=%d tenancy=%d",
			hostMiddlewares.ctxCalls,
			hostMiddlewares.authCalls,
			hostMiddlewares.tenancyCalls,
		)
	}
	if service.lastList.Resource != "roles" || service.lastList.PageNum != 3 || service.lastList.PageSize != 7 {
		t.Fatalf("unexpected list input: %#v", service.lastList)
	}
	if got := strings.TrimSpace(fmt.Sprint(service.lastList.Filters["name"])); got != "admin" {
		t.Fatalf("expected roleName alias to populate name filter, got %q", got)
	}
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

func TestLegacyRouteInterceptorRejectsMissingBearer(t *testing.T) {
	service := &legacyInterceptorFakeService{}
	hostMiddlewares := &legacyInterceptorFakeHostMiddlewares{}
	server := startLegacyInterceptorServer(t, "missing-bearer", true, hostMiddlewares.routeMiddlewares(), service, func(group *ghttp.RouterGroup) {
		group.GET("/role", func(r *ghttp.Request) {
			t.Fatal("host /role handler must not run without bearer")
		})
	})

	resp, err := g.Client().Get(context.Background(), server+"/api/v1/role")
	if err != nil {
		t.Fatalf("request legacy route without bearer: %v", err)
	}
	defer resp.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d body %s", resp.StatusCode, resp.ReadAllString())
	}
	if service.listCalls != 0 {
		t.Fatalf("expected service not to be called, got %d calls", service.listCalls)
	}
}

func TestLegacyRouteInterceptorAliasesJSONAndPathParams(t *testing.T) {
	service := &legacyInterceptorFakeService{}
	hostMiddlewares := &legacyInterceptorFakeHostMiddlewares{}
	server := startLegacyInterceptorServer(t, "role-update", true, hostMiddlewares.routeMiddlewares(), service, func(group *ghttp.RouterGroup) {
		group.PUT("/role/{id}", func(r *ghttp.Request) {
			t.Fatal("host /role/{id} handler must not run for old compatibility route")
		})
	})

	body := g.Client().
		Header(map[string]string{"Authorization": "Bearer test-token"}).
		ContentJson().
		PutContent(
			context.Background(),
			server+"/api/v1/role/7",
			`{"roleName":"legacy-admin","roleKey":"legacy_admin","roleSort":9}`,
		)
	if service.lastUpdateResource != "roles" || service.lastUpdateID != 7 {
		t.Fatalf("unexpected update target resource=%q id=%d", service.lastUpdateResource, service.lastUpdateID)
	}
	for key, want := range map[string]string{"name": "legacy-admin", "key": "legacy_admin", "sort": "9"} {
		if got := fmt.Sprint(service.lastUpdateBody[key]); got != want {
			t.Fatalf("expected %s=%q after legacy aliasing, got %q in %#v", key, want, got, service.lastUpdateBody)
		}
	}
	for _, want := range []string{
		`"code":200`,
		`"msg":"修改成功"`,
		`"data":7`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected legacy update response to contain %s, got %s", want, body)
		}
	}
}

func TestLegacyRouteInterceptorInjectsDictCodePathParam(t *testing.T) {
	service := &legacyInterceptorFakeService{}
	hostMiddlewares := &legacyInterceptorFakeHostMiddlewares{}
	server := startLegacyInterceptorServer(t, "dict-data-get", true, hostMiddlewares.routeMiddlewares(), service, func(group *ghttp.RouterGroup) {
		group.GET("/dict/data/{dictCode}", func(r *ghttp.Request) {
			t.Fatal("host /dict/data/{dictCode} handler must not run for old compatibility route")
		})
	})

	body := legacyInterceptorGet(t, server+"/api/v1/dict/data/21")
	if service.lastGetResource != "dict-data" || service.lastGetID != 21 {
		t.Fatalf("unexpected get target resource=%q id=%d", service.lastGetResource, service.lastGetID)
	}
	for _, want := range []string{
		`"code":200`,
		`"msg":"查询成功"`,
		`"dictCode":21`,
		`"dictLabel":"正常"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected legacy dict response to contain %s, got %s", want, body)
		}
	}
}

func TestLegacyRouteInterceptorMapsMenuListToOldTree(t *testing.T) {
	service := &legacyInterceptorFakeService{}
	hostMiddlewares := &legacyInterceptorFakeHostMiddlewares{}
	server := startLegacyInterceptorServer(t, "menu-tree", true, hostMiddlewares.routeMiddlewares(), service, func(group *ghttp.RouterGroup) {
		group.GET("/menu", func(r *ghttp.Request) {
			t.Fatal("host /menu handler must not run for old compatibility route")
		})
	})

	body := legacyInterceptorGet(t, server+"/api/v1/menu?title=系统")
	if service.lastMenuFilters["title"] != "系统" {
		t.Fatalf("expected title filter to reach menu tree handler, got %#v", service.lastMenuFilters)
	}
	for _, want := range []string{
		`"menuId":10`,
		`"menuName":"系统管理"`,
		`"menuType":"M"`,
		`"permission":"system:menu:list"`,
		`"children"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected legacy menu response to contain %s, got %s", want, body)
		}
	}
}

func TestLegacyRouteInterceptorMapsProfileFromPluginService(t *testing.T) {
	service := &legacyInterceptorFakeService{}
	hostMiddlewares := &legacyInterceptorFakeHostMiddlewares{}
	server := startLegacyInterceptorServer(t, "profile", true, hostMiddlewares.routeMiddlewares(), service, func(group *ghttp.RouterGroup) {
		group.GET("/user/profile", func(r *ghttp.Request) {
			t.Fatal("host /user/profile handler must not run for old compatibility route")
		})
	})

	body := legacyInterceptorGet(t, server+"/api/v1/user/profile")
	if !service.profileCalled {
		t.Fatal("expected plugin profile service to be called")
	}
	for _, want := range []string{
		`"userId":3`,
		`"nickName":"管理员"`,
		`"deptName":"研发部"`,
		`"roleName":"admin"`,
		`"postName":"工程师"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected legacy profile response to contain %s, got %s", want, body)
		}
	}
}

func TestLegacyRouteInterceptorBypassesNonLegacyRoutes(t *testing.T) {
	service := &legacyInterceptorFakeService{}
	hostMiddlewares := &legacyInterceptorFakeHostMiddlewares{}
	server := startLegacyInterceptorServer(t, "modern-route", true, hostMiddlewares.routeMiddlewares(), service, func(group *ghttp.RouterGroup) {
		group.GET("/modern", func(r *ghttp.Request) {
			r.Response.WriteJson(map[string]any{"host": true})
		})
	})

	body := g.Client().GetContent(context.Background(), server+"/api/v1/modern")
	if !strings.Contains(body, `"host":true`) {
		t.Fatalf("expected non-legacy route to pass through, got %s", body)
	}
	if hostMiddlewares.ctxCalls != 0 || hostMiddlewares.authCalls != 0 || hostMiddlewares.tenancyCalls != 0 {
		t.Fatalf("expected non-legacy route not to run host middleware wrappers, got ctx=%d auth=%d tenancy=%d",
			hostMiddlewares.ctxCalls,
			hostMiddlewares.authCalls,
			hostMiddlewares.tenancyCalls,
		)
	}
}

func TestLegacyRouteInterceptorDisabledPluginBypassesMiddleware(t *testing.T) {
	service := &legacyInterceptorFakeService{}
	hostMiddlewares := &legacyInterceptorFakeHostMiddlewares{}
	server := startLegacyInterceptorServer(t, "disabled", false, hostMiddlewares.routeMiddlewares(), service, func(group *ghttp.RouterGroup) {
		group.GET("/role", func(r *ghttp.Request) {
			r.Response.WriteJson(map[string]any{"host": true})
		})
	})

	body := g.Client().GetContent(context.Background(), server+"/api/v1/role")
	if !strings.Contains(body, `"host":true`) {
		t.Fatalf("expected disabled plugin middleware to bypass dispatch, got %s", body)
	}
	if service.listCalls != 0 {
		t.Fatalf("expected disabled plugin not to call service, got %d calls", service.listCalls)
	}
	if hostMiddlewares.ctxCalls != 0 || hostMiddlewares.authCalls != 0 || hostMiddlewares.tenancyCalls != 0 {
		t.Fatalf("expected disabled plugin not to run host middleware wrappers, got ctx=%d auth=%d tenancy=%d",
			hostMiddlewares.ctxCalls,
			hostMiddlewares.authCalls,
			hostMiddlewares.tenancyCalls,
		)
	}
}

func legacyInterceptorGet(t *testing.T, url string) string {
	t.Helper()
	return g.Client().
		Header(map[string]string{"Authorization": "Bearer test-token"}).
		GetContent(context.Background(), url)
}

func startLegacyInterceptorServer(
	t *testing.T,
	name string,
	enabled bool,
	middlewares pluginhost.RouteMiddlewares,
	service uidentitysvc.Service,
	register func(group *ghttp.RouterGroup),
) string {
	t.Helper()
	server := g.Server("legacy-interceptor-" + name + "-" + guid.S())
	server.SetDumpRouterMap(false)
	server.SetPort(0)
	server.Group("/api/v1", register)
	registrar := pluginhost.NewGlobalMiddlewareRegistrar(server, pluginID, func(_ context.Context, _ string) bool {
		return enabled
	})
	controller := uidentitycontroller.NewLegacy(service)
	if err := registerLegacyRouteInterceptors(registrar, middlewares, controller); err != nil {
		t.Fatalf("register legacy interceptor: %v", err)
	}
	server.Start()
	t.Cleanup(func() {
		server.Shutdown()
	})
	time.Sleep(100 * time.Millisecond)
	return fmt.Sprintf("http://127.0.0.1:%d", server.GetListenedPort())
}

type legacyInterceptorFakeHostMiddlewares struct {
	ctxCalls     int
	authCalls    int
	authHeader   string
	tenancyCalls int
}

func (f *legacyInterceptorFakeHostMiddlewares) routeMiddlewares() pluginhost.RouteMiddlewares {
	noop := func(r *ghttp.Request) {
		r.Middleware.Next()
	}
	return pluginhost.NewRouteMiddlewares(
		noop,
		noop,
		noop,
		noop,
		func(r *ghttp.Request) {
			f.ctxCalls++
			r.Middleware.Next()
		},
		func(r *ghttp.Request) {
			f.authCalls++
			f.authHeader = r.GetHeader("Authorization")
			if strings.TrimSpace(f.authHeader) == "" || !strings.HasPrefix(f.authHeader, "Bearer ") {
				r.Response.WriteStatus(http.StatusUnauthorized)
				return
			}
			r.Middleware.Next()
		},
		func(r *ghttp.Request) {
			f.tenancyCalls++
			r.Middleware.Next()
		},
		noop,
	)
}

type legacyInterceptorFakeService struct {
	uidentitysvc.Service

	listCalls          int
	lastList           uidentitysvc.LegacySystemResourceListInput
	lastGetResource    string
	lastGetID          int64
	lastUpdateResource string
	lastUpdateID       int64
	lastUpdateBody     map[string]any
	lastMenuFilters    map[string]any
	profileCalled      bool
}

func (s *legacyInterceptorFakeService) ListLegacySystemResource(_ context.Context, in uidentitysvc.LegacySystemResourceListInput) (*uidentitysvc.ResourceListOutput, error) {
	s.listCalls++
	s.lastList = in
	return &uidentitysvc.ResourceListOutput{
		Total: 1,
		List:  []uidentitysvc.Record{{"roleId": 1, "roleName": "admin"}},
	}, nil
}

func (s *legacyInterceptorFakeService) GetLegacySystemResource(_ context.Context, resource string, id int64) (uidentitysvc.Record, error) {
	s.lastGetResource = resource
	s.lastGetID = id
	if resource == "dict-data" {
		return uidentitysvc.Record{"dictCode": id, "dictLabel": "正常", "dictValue": "2", "dictType": "sys_normal_disable"}, nil
	}
	return uidentitysvc.Record{"id": id}, nil
}

func (s *legacyInterceptorFakeService) UpdateLegacySystemResource(_ context.Context, resource string, id int64, body map[string]any) error {
	s.lastUpdateResource = resource
	s.lastUpdateID = id
	s.lastUpdateBody = body
	return nil
}

func (s *legacyInterceptorFakeService) LegacyMenuTree(_ context.Context, filters map[string]any) ([]uidentitysvc.Record, error) {
	s.lastMenuFilters = filters
	return []uidentitysvc.Record{
		{
			"menuId":     10,
			"menuName":   "系统管理",
			"title":      "系统管理",
			"menuType":   "M",
			"permission": "system:menu:list",
			"children":   []uidentitysvc.Record{{"menuId": 11, "menuName": "菜单", "title": "菜单", "menuType": "C"}},
		},
	}, nil
}

func (s *legacyInterceptorFakeService) LegacySystemProfile(context.Context) (uidentitysvc.Record, error) {
	s.profileCalled = true
	return uidentitysvc.Record{
		"user":  uidentitysvc.Record{"userId": 3, "nickName": "管理员", "dept": uidentitysvc.Record{"deptId": 2, "deptName": "研发部"}},
		"roles": []uidentitysvc.Record{{"roleId": 1, "roleName": "admin", "roleKey": "admin"}},
		"posts": []uidentitysvc.Record{{"postId": 5, "postName": "工程师", "postCode": "engineer"}},
	}, nil
}
