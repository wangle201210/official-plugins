// This file builds the media plugin-owned OpenAPI document without using the
// host-wide lina-core apidoc service.

package apidoc

import (
	"context"
	"net/http"
	"reflect"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/util/gmeta"

	mediacontroller "lina-plugin-media/backend/internal/controller/media"
	mediaopencontroller "lina-plugin-media/backend/internal/controller/mediaopen"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

const (
	bearerAuthSecurityName  = "BearerAuth"
	innerAPIKeySecurityName = "InnerApiKeyAuth"
	openAPIAccessMetaKey    = "access"
	openAPIAccessPublic     = "public"
)

// routeSecurityOverrides records operations that should not inherit document-level BearerAuth.
type routeSecurityOverrides map[string]map[string]struct{}

// Build returns an OpenAPI document containing only media plugin routes.
func Build(_ context.Context, mediaSvc mediasvc.Service) (*goai.OpenApiV3, error) {
	if mediaSvc == nil {
		return nil, gerror.New("media apidoc requires media service")
	}
	document := newDocument()
	innerAPIKeyRoutes := routeSecurityOverrides{}

	publicController, err := mediaopencontroller.NewV1(mediaSvc)
	if err != nil {
		return nil, err
	}
	protectedController, err := mediacontroller.NewV1(mediaSvc)
	if err != nil {
		return nil, err
	}
	if err = addController(document, "/api/v1", publicController, innerAPIKeyRoutes); err != nil {
		return nil, err
	}
	if err = addController(document, "/api/v1", protectedController, nil); err != nil {
		return nil, err
	}
	applySecurity(document, innerAPIKeyRoutes)
	return document, nil
}

// newDocument creates the media plugin document metadata and security schemes.
func newDocument() *goai.OpenApiV3 {
	document := goai.New()
	document.Info.Title = "Media Plugin API"
	document.Info.Description = "Media management and mediaopen API reference."
	document.Info.Version = "v1"
	document.Servers = &goai.Servers{
		{
			URL:         "/",
			Description: "Current LinaPro host",
		},
	}
	document.Components.SecuritySchemes = goai.SecuritySchemes{
		bearerAuthSecurityName: goai.SecuritySchemeRef{
			Value: &goai.SecurityScheme{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
				Description:  "LinaPro Bearer token for media management APIs.",
			},
		},
		innerAPIKeySecurityName: goai.SecuritySchemeRef{
			Value: &goai.SecurityScheme{
				Type:        "apiKey",
				Name:        "X-Inner-Api-Key",
				In:          "header",
				Description: "Mediaopen inner API key. The default value is media unless innerapi.apiKey is explicitly configured.",
			},
		},
	}
	document.Security = &goai.SecurityRequirements{{bearerAuthSecurityName: {}}}
	return document
}

// addController registers one controller object's documentable methods into
// the target OpenAPI document under the public API prefix.
func addController(
	document *goai.OpenApiV3,
	prefix string,
	controller interface{},
	innerAPIKeyRoutes routeSecurityOverrides,
) error {
	if document == nil || controller == nil {
		return nil
	}
	reflectValue := reflect.ValueOf(controller)
	if !reflectValue.IsValid() {
		return nil
	}
	if reflectValue.Kind() == reflect.Struct {
		newValue := reflect.New(reflectValue.Type())
		newValue.Elem().Set(reflectValue)
		reflectValue = newValue
	}
	if reflectValue.Kind() != reflect.Pointer || reflectValue.Elem().Kind() != reflect.Struct {
		return nil
	}
	reflectType := reflectValue.Type()
	for i := 0; i < reflectValue.NumMethod(); i++ {
		method := reflectType.Method(i)
		if method.Name == "Init" || method.Name == "Shut" {
			continue
		}
		handler := reflectValue.Method(i).Interface()
		if !isDocumentableHandler(handler) {
			continue
		}
		reqObject, ok := newHandlerReqObject(handler)
		if !ok {
			continue
		}
		path := joinOpenAPIPath(prefix, gmeta.Get(reqObject, "path").String())
		methods := expandOpenAPIMethods(gmeta.Get(reqObject, "method").String())
		if handlerDeclaresPublicAccess(reqObject) {
			innerAPIKeyRoutes.add(path, methods)
		}
		for _, item := range methods {
			if err := document.Add(goai.AddInput{
				Path:   path,
				Method: item,
				Object: handler,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

// applySecurity marks mediaopen routes with InnerApiKeyAuth while media
// management routes inherit the document-level BearerAuth requirement.
func applySecurity(document *goai.OpenApiV3, innerAPIKeyRoutes routeSecurityOverrides) {
	if document == nil {
		return
	}
	for path, item := range document.Paths {
		visitPathOperations(&item, func(method string, operation *goai.Operation) {
			if operation != nil && innerAPIKeyRoutes.contains(path, method) {
				security := goai.SecurityRequirements{{innerAPIKeySecurityName: {}}}
				operation.Security = &security
			}
		})
		document.Paths[path] = item
	}
}

// handlerDeclaresPublicAccess reports whether a request DTO marks the route as mediaopen public.
func handlerDeclaresPublicAccess(reqObject interface{}) bool {
	accessMode := strings.TrimSpace(gmeta.Get(reqObject, openAPIAccessMetaKey).String())
	return strings.EqualFold(accessMode, openAPIAccessPublic)
}

// visitPathOperations applies a callback to every method operation on a path.
func visitPathOperations(path *goai.Path, visit func(method string, operation *goai.Operation)) {
	if path == nil || visit == nil {
		return
	}
	visit(http.MethodConnect, path.Connect)
	visit(http.MethodDelete, path.Delete)
	visit(http.MethodGet, path.Get)
	visit(http.MethodHead, path.Head)
	visit(http.MethodOptions, path.Options)
	visit(http.MethodPatch, path.Patch)
	visit(http.MethodPost, path.Post)
	visit(http.MethodPut, path.Put)
	visit(http.MethodTrace, path.Trace)
}

// newHandlerReqObject creates the request DTO instance used for route metadata.
func newHandlerReqObject(handler interface{}) (interface{}, bool) {
	reflectType := reflect.TypeOf(handler)
	if !isDocumentableHandler(handler) || reflectType == nil {
		return nil, false
	}
	return reflect.New(reflectType.In(1).Elem()).Interface(), true
}

// isDocumentableHandler verifies the GoFrame route handler signature used by OpenAPI generation.
func isDocumentableHandler(handler interface{}) bool {
	reflectType := reflect.TypeOf(handler)
	if reflectType == nil || reflectType.Kind() != reflect.Func {
		return false
	}
	if reflectType.NumIn() != 2 || reflectType.NumOut() != 2 {
		return false
	}
	if !reflectType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return false
	}
	if !reflectType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return false
	}
	return reflectType.In(1).Kind() == reflect.Pointer && reflectType.In(1).Elem().Kind() == reflect.Struct
}

// expandOpenAPIMethods expands empty or ALL route methods to GoFrame's HTTP method set.
func expandOpenAPIMethods(method string) []string {
	normalized := strings.ToUpper(strings.TrimSpace(method))
	if normalized == "" || normalized == "ALL" {
		methods := make([]string, 0, len(ghttpSupportedMethods()))
		for _, item := range ghttpSupportedMethods() {
			methods = append(methods, strings.ToUpper(strings.TrimSpace(item)))
		}
		return methods
	}
	return []string{normalized}
}

// ghttpSupportedMethods isolates the external method list for tests and readability.
func ghttpSupportedMethods() []string {
	return []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodHead,
		http.MethodOptions,
		http.MethodConnect,
		http.MethodTrace,
	}
}

// joinOpenAPIPath joins the host API prefix and DTO path into one OpenAPI path.
func joinOpenAPIPath(prefix string, path string) string {
	normalizedPrefix := normalizeOpenAPIPath(prefix)
	normalizedPath := normalizeOpenAPIPath(path)
	if normalizedPrefix == "/" {
		return normalizedPath
	}
	if normalizedPath == "/" {
		return normalizedPrefix
	}
	return normalizedPrefix + "/" + strings.TrimLeft(normalizedPath, "/")
}

// normalizeOpenAPIPath canonicalizes an OpenAPI path with one leading slash.
func normalizeOpenAPIPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" || trimmed == "/" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return strings.TrimRight(trimmed, "/")
}

// add records security overrides for the supplied path and methods.
func (r routeSecurityOverrides) add(path string, methods []string) {
	if r == nil || len(methods) == 0 {
		return
	}
	normalizedPath := normalizeOpenAPIPath(path)
	if r[normalizedPath] == nil {
		r[normalizedPath] = make(map[string]struct{}, len(methods))
	}
	for _, method := range methods {
		r[normalizedPath][strings.ToUpper(strings.TrimSpace(method))] = struct{}{}
	}
}

// contains reports whether the given operation should use InnerApiKeyAuth.
func (r routeSecurityOverrides) contains(path string, method string) bool {
	if r == nil {
		return false
	}
	methods := r[normalizeOpenAPIPath(path)]
	if methods == nil {
		return false
	}
	_, ok := methods[strings.ToUpper(strings.TrimSpace(method))]
	return ok
}
