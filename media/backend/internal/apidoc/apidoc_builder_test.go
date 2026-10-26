// This file checks the generated media OpenAPI document exposes raw reports as
// JSON objects and retains the existing authentication and permission contracts.

package apidoc

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	v1 "lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// documentationOnlyService satisfies controller construction without implementing
// business methods: building documentation must never invoke those methods.
type documentationOnlyService struct {
	mediasvc.Service
}

// TestBuildDashboardReportPayloadSchemas validates actual generated and serialized schema types.
func TestBuildDashboardReportPayloadSchemas(t *testing.T) {
	document, err := Build(t.Context(), &documentationOnlyService{})
	if err != nil {
		t.Fatalf("build media OpenAPI: %v", err)
	}
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("serialize media OpenAPI: %v", err)
	}
	var rendered struct {
		Components struct {
			Schemas map[string]struct {
				Type       string `json:"type"`
				Properties map[string]struct {
					Type                 string          `json:"type"`
					Format               string          `json:"format"`
					Nullable             bool            `json:"nullable"`
					AdditionalProperties json.RawMessage `json:"additionalProperties"`
					Example              json.RawMessage `json:"example"`
				} `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(raw, &rendered); err != nil {
		t.Fatalf("decode generated media OpenAPI: %v", err)
	}
	for _, itemType := range []reflect.Type{
		reflect.TypeFor[v1.DashboardInstanceItem](),
		reflect.TypeFor[v1.DashboardNetworkLink](),
		reflect.TypeFor[v1.DashboardStreamItem](),
		reflect.TypeFor[v1.DashboardSessionItem](),
	} {
		typeName := itemType.Name()
		t.Run(typeName, func(t *testing.T) {
			matches := 0
			for name, schema := range rendered.Components.Schemas {
				if name != typeName && !strings.HasSuffix(name, "."+typeName) {
					continue
				}
				matches++
				payload, exists := schema.Properties["report_payload"]
				if !exists || payload.Type != "object" || !payload.Nullable || payload.Format == "byte" || payload.Format == "binary" {
					t.Fatalf("%s report_payload must document a JSON object: exists=%v schema=%+v", name, exists, payload)
				}
				var example, expected map[string]json.RawMessage
				if err := json.Unmarshal(payload.Example, &example); err != nil || example == nil {
					t.Fatalf("%s report_payload example must be an object: example=%s err=%v", name, payload.Example, err)
				}
				field, exists := itemType.FieldByName("ReportPayload")
				if !exists {
					t.Fatalf("%s missing source ReportPayload field", typeName)
				}
				if err := json.Unmarshal([]byte(field.Tag.Get("eg")), &expected); err != nil {
					t.Fatalf("decode %s source example: %v", typeName, err)
				}
				if !reflect.DeepEqual(example, expected) {
					t.Fatalf("%s example changed source values: got=%v want=%v", name, example, expected)
				}
				// Original protobuf fields include scalar values, arrays and objects.
				// An object-only map value schema would reject the actual API response.
				switch strings.TrimSpace(string(payload.AdditionalProperties)) {
				case "", "true", "{}":
				default:
					t.Fatalf("%s report_payload values must accept any JSON type, got additionalProperties=%s; interface schema=%+v", name, payload.AdditionalProperties, rendered.Components.Schemas["interface"])
				}
				if typeName == "DashboardStreamItem" || typeName == "DashboardSessionItem" {
					if subcommand := schema.Properties["report_subcommand"]; subcommand.Type != "integer" {
						t.Fatalf("%s report_subcommand must document an integer: %+v", name, subcommand)
					}
				}
			}
			if matches != 1 {
				t.Fatalf("expected exactly one generated %s schema, got %d", typeName, matches)
			}
		})
	}

	bearer := document.Components.SecuritySchemes["BearerAuth"].Value
	if bearer == nil || bearer.Type != "http" || bearer.Scheme != "bearer" || bearer.BearerFormat != "JWT" {
		t.Fatalf("management BearerAuth scheme changed: %#v", bearer)
	}
	inner := document.Components.SecuritySchemes["InnerApiKeyAuth"].Value
	if inner == nil || inner.Type != "apiKey" || inner.Name != "X-Inner-Api-Key" || inner.In != "header" {
		t.Fatalf("inner API key scheme changed: %#v", inner)
	}
	if document.Security == nil || len(*document.Security) != 1 {
		t.Fatalf("expected one document-level bearer requirement: %#v", document.Security)
	}
	if _, exists := (*document.Security)[0]["BearerAuth"]; !exists {
		t.Fatal("media OpenAPI lost its default bearer requirement")
	}
	for _, request := range []reflect.Type{
		reflect.TypeFor[v1.GetDashboardNodeOverviewReq](),
		reflect.TypeFor[v1.ListDashboardInstancesReq](),
		reflect.TypeFor[v1.ListDashboardStreamsReq](),
		reflect.TypeFor[v1.ListDashboardSessionsReq](),
	} {
		meta, exists := request.FieldByName("Meta")
		if !exists || meta.Tag.Get("permission") != "media:management:query" || meta.Tag.Get("method") != "get" {
			t.Fatalf("dashboard request lost its protected GET contract: %s", request.Name())
		}
		path := "/api/v1" + meta.Tag.Get("path")
		operation := document.Paths[path].Get
		if operation == nil || operation.Security != nil {
			t.Fatalf("dashboard GET %s must inherit document bearer authentication: %#v", path, operation)
		}
	}
	mediaopen := document.Paths["/api/v1/strategies/user-device"].Get
	if mediaopen == nil || mediaopen.Security == nil || len(*mediaopen.Security) != 1 {
		t.Fatalf("mediaopen API lost its explicit API-key requirement: %#v", mediaopen)
	}
	if _, exists := (*mediaopen.Security)[0]["InnerApiKeyAuth"]; !exists {
		t.Fatal("mediaopen API unexpectedly inherits bearer authentication")
	}
}
