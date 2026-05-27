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

const mediaHostAPIDocsPagePath = "/stoplight/apidocs.html"

// mediaAPIDocsHTML renders the plugin-scoped Stoplight page and keeps it
// pointed at the media-only OpenAPI document instead of the host-wide document.
const mediaAPIDocsHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Media API Documentation</title>
  <link rel="stylesheet" href="/stoplight/styles.min.css" />
  <style>
    html,
    body {
      height: 100%;
      margin: 0;
      background: #ffffff;
    }

    body {
      min-height: 100vh;
    }

    elements-api {
      display: block;
      width: 100%;
      height: 100vh;
    }
  </style>
</head>
<body>
  <script>
    (function() {
      var params = new URLSearchParams(window.location.search);
      var token = params.get('token');
      var innerApiKey = params.get('innerApiKey') || params.get('apiKey');
      var lang = params.get('lang') || 'zh-CN';
      var securityValues = {};
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
      document.body.appendChild(apiElement);
    })();
  </script>
  <script src="/stoplight/web-components.min.js"></script>
</body>
</html>`

// mediaRegisterHostAPIDocBlock blocks the host-wide Stoplight HTML page while
// leaving shared Stoplight static assets available for the media-owned page.
func mediaRegisterHostAPIDocBlock(registrar pluginhost.GlobalMiddlewareRegistrar) error {
	if registrar == nil {
		return nil
	}
	return registrar.Bind(pluginhost.MiddlewareScope(mediaHostAPIDocsPagePath), func(r *ghttp.Request) {
		r.Response.WriteStatus(http.StatusNotFound)
		r.ExitAll()
	})
}

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
