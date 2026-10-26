// This file checks the generated media OpenAPI document exposes raw reports as
// JSON objects and retains the existing authentication and permission contracts.

package apidoc

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/net/goai"

	v1 "lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// documentationOnlyService satisfies controller construction without implementing
// business methods: building documentation must never invoke those methods.
type documentationOnlyService struct {
	mediasvc.Service
}

// TestBuildDashboardInstanceTimestampSchemas checks actual serialized OpenAPI
// types, explicit nullability and the units third-party clients must consume.
func TestBuildDashboardInstanceTimestampSchemas(t *testing.T) {
	document, err := Build(t.Context(), &documentationOnlyService{})
	if err != nil {
		t.Fatalf("build instance timestamp documentation: %v", err)
	}
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("serialize instance timestamp documentation: %v", err)
	}
	var rendered struct {
		Components struct {
			Schemas map[string]struct {
				Properties map[string]struct {
					Type        string `json:"type"`
					Format      string `json:"format"`
					Nullable    bool   `json:"nullable"`
					Description string `json:"description"`
				} `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(raw, &rendered); err != nil {
		t.Fatalf("decode instance timestamp documentation: %v", err)
	}
	typeName := reflect.TypeFor[v1.DashboardInstanceItem]().Name()
	matches := 0
	for name, schema := range rendered.Components.Schemas {
		if name != typeName && !strings.HasSuffix(name, "."+typeName) {
			continue
		}
		matches++
		for _, field := range []string{"report_time", "updated_at"} {
			property, exists := schema.Properties[field]
			if !exists || property.Type != "integer" || property.Format != "int64" || !property.Nullable {
				t.Fatalf("instance %s must be a nullable integer/int64: exists=%v schema=%+v", field, exists, property)
			}
			if !strings.Contains(property.Description, "Unix timestamp in milliseconds") || !strings.Contains(property.Description, "null") {
				t.Fatalf("instance %s must explain milliseconds and unavailable timestamps: %q", field, property.Description)
			}
		}
	}
	if matches != 1 {
		t.Fatalf("expected one instance response schema, got %d", matches)
	}
}

// TestBuildDashboardNetworkRateUnits checks actual generated schema documentation
// for all network resource projections, including the nested direct resources.
func TestBuildDashboardNetworkRateUnits(t *testing.T) {
	document, err := Build(t.Context(), &documentationOnlyService{})
	if err != nil {
		t.Fatalf("build network throughput documentation: %v", err)
	}
	for _, typeName := range []string{"DashboardNodeOverviewItem", "DashboardNodeResources", "DashboardInstanceItem"} {
		matches := 0
		for name, schema := range document.Components.Schemas.Map() {
			if name != typeName && !strings.HasSuffix(name, "."+typeName) {
				continue
			}
			matches++
			if schema.Value == nil || schema.Value.Properties == nil {
				t.Fatalf("missing properties for network resource schema %s", name)
			}
			for _, field := range []string{"network_in", "network_out"} {
				property := schema.Value.Properties.Get(field)
				if property == nil || property.Value == nil {
					t.Fatalf("missing %s.%s network field", name, field)
				}
				description := property.Value.Description
				if !strings.Contains(description, "Kbps") || !strings.Contains(description, "1000") || strings.Contains(strings.ToUpper(description), "KB/S") {
					t.Fatalf("%s.%s must describe decimal Kbps without the old byte-rate unit: %q", name, field, description)
				}
				if property.Value.Type != "number" {
					t.Fatalf("%s.%s must retain fractional numeric values: %#v", name, field, property.Value)
				}
			}
		}
		if matches != 1 {
			t.Fatalf("expected one %s schema, got %d", typeName, matches)
		}
	}
}

// TestBuildDashboardSummarySchema verifies the new parameterless summary route
// publishes three int64 counters and inherits the management authentication scheme.
func TestBuildDashboardSummarySchema(t *testing.T) {
	document, err := Build(t.Context(), &documentationOnlyService{})
	if err != nil {
		t.Fatalf("build summary OpenAPI: %v", err)
	}
	operation := document.Paths["/api/v1/media/dashboard/summary"].Get
	if operation == nil || operation.Security != nil || len(operation.Parameters) != 0 {
		t.Fatalf("summary must be a parameterless GET inheriting bearer auth: %#v", operation)
	}
	if !strings.Contains(operation.Description, "running") || !strings.Contains(operation.Description, "close_time") ||
		!strings.Contains(operation.Description, "关联") || !strings.Contains(operation.Description, "10000") {
		t.Fatalf("summary must explain running-stream session association and uncapped totals: %q", operation.Description)
	}
	typeName := reflect.TypeFor[v1.GetDashboardSummaryRes]().Name()
	matches := 0
	for name, schema := range document.Components.Schemas.Map() {
		if name != typeName && !strings.HasSuffix(name, "."+typeName) {
			continue
		}
		matches++
		if schema.Value == nil || schema.Value.Properties == nil || len(schema.Value.Properties.Map()) != 3 {
			t.Fatalf("summary must publish exactly three counters: %#v", schema)
		}
		for _, field := range []string{"active_stream_count", "active_session_count", "alive_node_count"} {
			property := schema.Value.Properties.Get(field)
			if property == nil || property.Value == nil || property.Value.Type != "integer" || property.Value.Format != "int64" {
				t.Fatalf("%s must be documented as integer/int64: %#v", field, property)
			}
			if field == "alive_node_count" {
				description := property.Value.Description
				if !strings.Contains(description, "updated_at") || !strings.Contains(description, "3分钟") || !strings.Contains(description, "服务端") ||
					(!strings.Contains(description, "不按状态") && !strings.Contains(description, "不按status")) ||
					strings.Contains(description, "status=running") || strings.Contains(description, "close_time IS NULL") {
					t.Fatalf("alive node count must document its three-minute server-update window independently of status: %q", description)
				}
			}
		}
	}
	if matches != 1 {
		t.Fatalf("expected one summary response schema, got %d", matches)
	}
	meta, ok := reflect.TypeFor[v1.GetDashboardSummaryReq]().FieldByName("Meta")
	if !ok || meta.Tag.Get("permission") != "media:management:query" {
		t.Fatal("summary lost its management query permission")
	}
	if document.Security == nil || len(*document.Security) != 1 {
		t.Fatal("summary document lost the default authentication requirement")
	}
	if _, ok := (*document.Security)[0]["BearerAuth"]; !ok {
		t.Fatal("summary document no longer defaults to BearerAuth")
	}
}

// TestBuildDashboardFilterDocumentation verifies generated filter and response
// descriptions use collector values and distinguish limited details from counts.
func TestBuildDashboardFilterDocumentation(t *testing.T) {
	document, err := Build(t.Context(), &documentationOnlyService{})
	if err != nil {
		t.Fatalf("build dashboard filter documentation: %v", err)
	}
	parameter := func(path, name string) *goai.Parameter {
		t.Helper()
		operation := document.Paths[path].Get
		if operation == nil {
			t.Fatalf("missing documented GET %s", path)
		}
		for _, ref := range operation.Parameters {
			if ref.Value != nil && ref.Value.Name == name {
				return ref.Value
			}
		}
		t.Fatalf("missing %s parameter on %s", name, path)
		return nil
	}
	property := func(typeName, name string) *goai.Schema {
		t.Helper()
		for key, schema := range document.Components.Schemas.Map() {
			if key != typeName && !strings.HasSuffix(key, "."+typeName) {
				continue
			}
			if schema.Value != nil && schema.Value.Properties != nil {
				ref := schema.Value.Properties.Get(name)
				if ref != nil && ref.Value != nil {
					return ref.Value
				}
			}
		}
		t.Fatalf("missing documented %s.%s", typeName, name)
		return nil
	}
	assertValues := func(description string, expected []string) {
		t.Helper()
		words := strings.FieldsFunc(description, func(char rune) bool {
			return !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_')
		})
		for _, value := range expected {
			if !slices.Contains(words, value) {
				t.Fatalf("description omitted collector value %q: %q", value, description)
			}
		}
	}
	statuses := []string{"universal", "running", "inactive", "error", "closed", "timeout", "cancelled", "failed"}
	statusParameter := parameter("/api/v1/media/dashboard/streams", "status")
	assertValues(statusParameter.Description, statuses)
	if statusParameter.Example != "running" || statusParameter.Required {
		t.Fatalf("stream status must remain optional with the running example: %#v", statusParameter)
	}
	for _, typeName := range []string{"DashboardStreamItem", "DashboardTopologyStream"} {
		status := property(typeName, "status")
		assertValues(status.Description, statuses)
		if status.Example != "running" {
			t.Fatalf("%s.status must use the actual running value, got %#v", typeName, status.Example)
		}
	}
	protocols := []string{"UNIVERSAL", "RTMP", "RTSP", "HLS", "HTTP_FLV", "WS_FLV", "HTTPS_FLV", "WSS_FLV", "GB28181"}
	protocol := parameter("/api/v1/media/dashboard/sessions", "protocolType")
	assertValues(protocol.Description, protocols)
	if protocol.Required {
		t.Fatal("session protocol filter must remain optional")
	}
	streamID := parameter("/api/v1/media/dashboard/sessions", "streamId")
	if !streamID.Required || streamID.In != "query" {
		t.Fatalf("sessions must document a required streamId query parameter: %#v", streamID)
	}
	sessionDescription := document.Paths["/api/v1/media/dashboard/sessions"].Get.Description
	if !strings.Contains(sessionDescription, "10000") || !strings.Contains(sessionDescription, "session_count") ||
		!strings.Contains(sessionDescription, "完整计数") || !strings.Contains(sessionDescription, "明细") {
		t.Fatalf("sessions must distinguish detail limits from complete counts: %q", sessionDescription)
	}
	countDescription := property("DashboardSessionProtocol", "session_count").Description
	if !strings.Contains(countDescription, "10000") || !strings.Contains(countDescription, "不受") ||
		!strings.Contains(countDescription, "active_session_list") {
		t.Fatalf("session count description must explain why it can exceed returned details: %q", countDescription)
	}
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
