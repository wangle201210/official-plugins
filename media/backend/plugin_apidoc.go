// This file binds the media plugin-owned OpenAPI JSON and Stoplight page
// endpoints without changing the host-wide LinaPro API document service.

package backend

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/plugin/pluginhost"
	mediaapidoc "lina-plugin-media/backend/internal/apidoc"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// mediaAPIDocsHTML renders the plugin-scoped HTTP API document and TCP
// collection protocol notes without changing the host-wide API document.
const mediaAPIDocsHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Media API Documentation</title>
  <link rel="stylesheet" href="/admin/stoplight/styles.min.css" />
  <style>
    html,
    body {
      height: 100%;
      margin: 0;
      background: #ffffff;
    }

    body {
      min-height: 100vh;
      color: #151b26;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }

    .media-doc-shell {
      display: grid;
      grid-template-rows: 52px minmax(0, 1fr);
      min-height: 100vh;
    }

    .media-doc-tabs {
      display: flex;
      align-items: center;
      gap: 8px;
      min-height: 52px;
      padding: 0 16px;
      border-bottom: 1px solid #e5e8ef;
      background: #ffffff;
    }

    .media-doc-tab {
      height: 34px;
      padding: 0 14px;
      border: 1px solid transparent;
      border-radius: 6px;
      background: transparent;
      color: #455266;
      font-size: 14px;
      cursor: pointer;
    }

    .media-doc-tab[aria-selected="true"] {
      border-color: #b8c4d9;
      background: #eef3fb;
      color: #13233a;
      font-weight: 600;
    }

    .media-doc-panel {
      min-height: 0;
    }

    .media-doc-panel[hidden] {
      display: none;
    }

    .media-tcp-doc {
      max-width: 1120px;
      padding: 28px 32px 40px;
    }

    .media-tcp-doc h1 {
      margin: 0 0 8px;
      font-size: 24px;
      line-height: 1.25;
      font-weight: 700;
    }

    .media-tcp-doc h2 {
      margin: 28px 0 12px;
      font-size: 17px;
      line-height: 1.35;
      font-weight: 700;
    }

    .media-tcp-doc p {
      margin: 0 0 12px;
      color: #4c586b;
      line-height: 1.7;
      font-size: 14px;
    }

    .media-tcp-doc code {
      padding: 1px 5px;
      border-radius: 4px;
      background: #f2f4f8;
      color: #22324a;
      font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size: 13px;
    }

    .media-tcp-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
      gap: 16px;
      margin-top: 16px;
    }

    .media-tcp-section {
      border: 1px solid #e5e8ef;
      border-radius: 8px;
      padding: 16px;
      background: #ffffff;
    }

    .media-tcp-section h3 {
      margin: 0 0 10px;
      font-size: 15px;
      line-height: 1.4;
    }

    .media-tcp-list {
      margin: 0;
      padding-left: 18px;
      color: #4c586b;
      line-height: 1.7;
      font-size: 14px;
    }

    .media-tcp-table {
      width: 100%;
      border-collapse: collapse;
      font-size: 14px;
      line-height: 1.5;
    }

    .media-tcp-table th,
    .media-tcp-table td {
      padding: 10px 12px;
      border: 1px solid #e5e8ef;
      text-align: left;
      vertical-align: top;
    }

    .media-tcp-table th {
      width: 210px;
      background: #f7f9fc;
      color: #26364d;
      font-weight: 600;
    }

    elements-api {
      display: block;
      width: 100%;
      height: calc(100vh - 52px);
    }
  </style>
</head>
<body>
  <div class="media-doc-shell">
    <nav class="media-doc-tabs" aria-label="Media API documentation views">
      <button class="media-doc-tab" type="button" role="tab" aria-controls="media-http-doc" data-view="http">HTTP API</button>
      <button class="media-doc-tab" type="button" role="tab" aria-controls="media-tcp-doc" data-view="tcp">TCP 采集协议</button>
    </nav>
    <section id="media-http-doc" class="media-doc-panel" role="tabpanel"></section>
    <section id="media-tcp-doc" class="media-doc-panel media-tcp-doc" role="tabpanel" hidden>
      <h1>TCP 采集协议</h1>
      <p>媒体采集 server 是 <code>net-flux</code> 兼容的 TCP 服务，不属于 HTTP OpenAPI 路径。HTTP API 视图只展示管理端和 <code>mediaopen</code> 路由；采集客户端应连接 <code>collectionServer.addr</code>，默认监听地址为 <code>:1911</code>。</p>
      <p>服务由 <code>collectionServer.enabled</code> 控制，默认关闭；启用后在宿主 <code>system.started</code> hook 中启动，并随宿主上下文停止。</p>

      <div class="media-tcp-grid">
        <section class="media-tcp-section">
          <h3>系统命令</h3>
          <ul class="media-tcp-list">
            <li><code>Ping</code>：返回包含相同时间戳的 <code>Pong</code>。</li>
          </ul>
        </section>
        <section class="media-tcp-section">
          <h3>数据上报</h3>
          <ul class="media-tcp-list">
            <li><code>MachineMetric</code>：写入实例/容器最新投影。</li>
            <li><code>NetworkMetric</code>：写入节点网络速率和延迟矩阵。</li>
            <li><code>StreamMetric</code>：写入流最新投影，支持 <code>STREAM_ADD</code> 与 <code>STREAM_DELETE</code> 生命周期事件。</li>
            <li><code>SessionMetric</code>：写入会话最新投影，支持 <code>SESSION_ADD</code> 与 <code>SESSION_DELETE</code> 生命周期事件。</li>
          </ul>
        </section>
        <section class="media-tcp-section">
          <h3>服务发现</h3>
          <ul class="media-tcp-list">
            <li><code>Instance</code>：注册实例到 Nacos。</li>
            <li><code>Deregister</code>：按实例名、节点、IP 和端口注销实例。</li>
            <li><code>Lookup</code>：按服务名和节点查询实例并返回 <code>LookupAck</code>。</li>
          </ul>
        </section>
      </div>

      <h2>配置项</h2>
      <table class="media-tcp-table">
        <tbody>
          <tr><th><code>collectionServer.enabled</code></th><td>是否启动 TCP 采集 server；默认 <code>false</code>。</td></tr>
          <tr><th><code>collectionServer.addr</code></th><td>TCP 监听地址；默认 <code>:1911</code>。</td></tr>
          <tr><th><code>collectionServer.discovery.enabled</code></th><td>是否启用 Nacos discovery 命令；默认 <code>false</code>。</td></tr>
          <tr><th><code>collectionServer.discovery.node</code></th><td>discovery 命令缺省节点；服务注册、查询和注销时映射为 Nacos group。</td></tr>
          <tr><th><code>collectionServer.discovery.notLoadCacheAtStart</code></th><td>是否跳过加载 Nacos SDK 本地磁盘缓存；默认 <code>true</code>。</td></tr>
        </tbody>
      </table>

      <h2>处理边界</h2>
      <p><code>config</code>、<code>event</code> 和 <code>control</code> 命令当前只接受并记录日志，不触发 LinaPro 业务状态变更。实时流数和会话数通过宿主发布给插件的共享 cache 维护，数据库中的 <code>media_report_instance</code> 是供看板读取的最新投影。</p>
      <p>联调可使用插件内工具：进入<code>apps/lina-plugins/media</code>，将环境变量设为<code>GOWORK=off</code>后执行<code>go run ./hack/tools/collection-client -action smoke -addr 127.0.0.1:1911</code>。</p>
    </section>
  </div>
  <script>
    (function() {
      var params = new URLSearchParams(window.location.search);
      var token = params.get('token');
      var innerApiKey = params.get('innerApiKey') || params.get('apiKey');
      var lang = params.get('lang') || 'zh-CN';
      var securityValues = {};
      var tabs = Array.prototype.slice.call(document.querySelectorAll('.media-doc-tab'));
      var panels = {
        http: document.getElementById('media-http-doc'),
        tcp: document.getElementById('media-tcp-doc')
      };
      document.documentElement.setAttribute('lang', lang);
      if (token) {
        securityValues.BearerAuth = token;
      }
      if (innerApiKey) {
        securityValues.InnerApiKeyAuth = innerApiKey;
      }
      if (Object.keys(securityValues).length > 0) {
        localStorage.setItem('TryIt_securitySchemeValues', JSON.stringify(securityValues));
      }
      var apiDescriptionUrl = '/api/v1/media/openapi.json';
      var apiElement = document.createElement('elements-api');
      apiElement.setAttribute('apiDescriptionUrl', apiDescriptionUrl);
      apiElement.setAttribute('router', 'memory');
      apiElement.setAttribute('layout', 'sidebar');
      apiElement.setAttribute('tryItCredentialsPolicy', 'same-origin');
      apiElement.setAttribute('hideInternal', 'true');
      apiElement.apiDescriptionUrl = apiDescriptionUrl;
      apiElement.router = 'memory';
      apiElement.layout = 'sidebar';
      apiElement.tryItCredentialsPolicy = 'same-origin';
      apiElement.hideInternal = true;
      panels.http.appendChild(apiElement);

      function showView(view) {
        var nextView = view === 'tcp' ? 'tcp' : 'http';
        tabs.forEach(function(tab) {
          tab.setAttribute('aria-selected', tab.getAttribute('data-view') === nextView ? 'true' : 'false');
        });
        panels.http.hidden = nextView !== 'http';
        panels.tcp.hidden = nextView !== 'tcp';
        if (window.history && window.history.replaceState) {
          window.history.replaceState(null, '', nextView === 'tcp' ? '#tcp' : window.location.pathname + window.location.search);
        }
      }

      tabs.forEach(function(tab) {
        tab.addEventListener('click', function() {
          showView(tab.getAttribute('data-view'));
        });
      });
      showView(params.get('view') === 'tcp' || window.location.hash === '#tcp' ? 'tcp' : 'http');
    })();
  </script>
  <script src="/admin/stoplight/web-components.min.js"></script>
</body>
</html>`

// mediaRegisterAPIDocRoutes binds the media-only OpenAPI JSON endpoint and page.
func mediaRegisterAPIDocRoutes(group pluginhost.RouteGroup, mediaSvc mediasvc.Service) {
	group.GET("/media/apidocs.html", func(r *ghttp.Request) {
		r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		r.Response.Write(mediaAPIDocsHTML)
		r.ExitAll()
	})
	group.GET("/media/openapi.json", func(r *ghttp.Request) {
		document, err := mediaapidoc.Build(r.Context(), mediaSvc)
		if err != nil {
			r.SetError(err)
			r.Response.WriteStatus(http.StatusInternalServerError)
			return
		}
		r.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
		r.Response.WriteJson(document)
		r.ExitAll()
	})
}
