// This file verifies full known-field protobuf payload preservation, including
// lifecycle deletes, without relying on normalized dashboard fields or test order.

package collection

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/dellinger2023/net-flux/gen"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
	"lina-plugin-media/backend/internal/model/entity"
)

// reportPayloadFixture contains independent complete messages with shared resource keys.
type reportPayloadFixture struct {
	machine *gen.MachineMetric
	network *gen.NetworkMetric
	stream  *gen.StreamMetric
	session *gen.SessionMetric
}

// newReportPayloadFixture explicitly populates every field in the four known reports.
func newReportPayloadFixture(prefix string) reportPayloadFixture {
	const sampledAt int64 = 1_790_000_000_123
	return reportPayloadFixture{
		machine: &gen.MachineMetric{
			MachineId: prefix + "-machine", InstanceId: prefix + "-instance", InstanceName: " Raw instance ",
			NodeId: prefix + "-node", NodeName: " Raw node ", Region: " Region ", NodeStatus: " warning ", Status: " running ",
			CpuUsage: 12.3456789, CpuCount: 8, MemUsed: 1_048_577, MemTotal: 8_388_609,
			DiskUsed: 1234567, DiskTotal: 9876543, DiskReadBytes: 1025, DiskWriteBytes: 2049,
			DiskReadTime: 17, DiskWriteTime: 19, DiskReadCount: 9_007_199_254_740_993, DiskWriteCount: 29,
			Timestamp: sampledAt, NetworkIn: 8193, NetworkOut: 16385, StartTime: sampledAt - 60000, Version: " v1.2.3 ",
			Extra: map[string]string{"vendor": "原始内容", "nul": "before\x00after", "node_id": "unused-node", "node_name": "unused-name", "empty": "", "quoted": "\"value\"\nline"},
		},
		network: &gen.NetworkMetric{
			MachineId: prefix + "-instance", NodeId: prefix + "-node", SourceIp: "192.0.2.1", DestinationIp: "192.0.2.2",
			Rtt: 18, Jitter: 3, PacketLoss: 0.0123456789, StatusCode: 503, Timestamp: sampledAt,
			Extra: map[string]string{"nul": "before\x00after", "url": "https://example.test:8443/probe", "status_code": "503", "success": "2", "total": "3", "reachable": "false", "error": "上游错误", "throughput_bps": "8193", "empty": ""},
		},
		stream: &gen.StreamMetric{
			MachineId: prefix + "-machine", MachineType: gen.MachineType(777), StreamId: prefix + "-stream",
			Status: gen.StreamStatus(778), VideoCodec: gen.StreamCodec(779), AudioCodec: gen.StreamCodec(780), Protocol: gen.StreamProtocol(781),
			Bitrate: 1537, Width: 1920, Height: 1080, StreamPath: "rtsp://example.test/raw/path", StreamAlias: " Raw alias ",
			Timestamp: sampledAt, Extra: map[string]string{"vendor": "原始流", "nul": "before\x00after", "empty": "", "source_url": "rtsp://fallback.test/ignored"},
			TenantId: "tenant-raw", NodeId: prefix + "-node", NodeName: " Raw node ", InstanceId: prefix + "-instance", InstanceName: " Raw instance ",
			Fps: 29.97654321, PacketLoss: 0.0123456789, StartTime: sampledAt - 60000, Duration: 60, AvgDelay: 25,
			TotalSessionsLifetime: 9_007_199_254_740_993, CurrentActiveSessions: 3, WatermarkEnabled: true, DeviceId: "device-raw",
		},
		session: &gen.SessionMetric{
			SessionId: prefix + "-session", StreamId: prefix + "-stream", StreamName: " Raw stream ", TenantId: "tenant-raw",
			ClientId: "client-raw", ClientIp: "192.0.2.15", ClientType: " custom-client ", UserName: " Raw user ", Protocol: gen.StreamProtocol(782),
			StartTime: sampledAt - 60000, PlayDuration: 60, CurrentFps: 29.97654321, CurrentBitrate: 1537, CurrentWidth: 1920, CurrentHeight: 1080,
			MachineId: prefix + "-machine", NodeId: prefix + "-node", NodeName: " Raw node ", InstanceId: prefix + "-instance", InstanceName: " Raw instance ",
			LinkHops: []*gen.LinkHop{
				{HopId: "hop-a", HopName: "完整跳点", HopType: "gateway", NodeId: "4", NodeName: "节点四", Latency: 12},
				{HopId: "hop-b", HopName: "第二跳", HopType: "media", NodeId: "5", NodeName: "节点五", Latency: 0},
			},
			TotalLinkLatency: 12, Timestamp: sampledAt, Extra: map[string]string{"vendor": "原始会话", "nul": "before\x00after", "empty": "", "unknown": "preserved"}, DeviceId: "device-raw",
		},
	}
}

// assertReportPayloadRoundTrip requires exact protobuf values and every declared JSON field.
func assertReportPayloadRoundTrip(t *testing.T, payload string, want proto.Message) {
	t.Helper()
	got := want.ProtoReflect().New().Interface()
	if err := protojson.Unmarshal([]byte(payload), got); err != nil {
		t.Fatalf("decode stored %T payload: %v", want, err)
	}
	if !proto.Equal(got, want) {
		t.Fatalf("%T payload changed original fields:\ngot: %v\nwant: %v", want, got, want)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(payload), &fields); err != nil {
		t.Fatalf("decode payload field names: %v", err)
	}
	descriptors := want.ProtoReflect().Descriptor().Fields()
	if len(fields) != descriptors.Len() {
		t.Fatalf("%T payload has %d keys, want all %d declared fields", want, len(fields), descriptors.Len())
	}
	for index := range descriptors.Len() {
		field := descriptors.Get(index)
		name := string(field.Name())
		if _, ok := fields[name]; !ok {
			t.Fatalf("%T payload omitted original field name %q", want, name)
		}
		if field.Kind() == protoreflect.EnumKind {
			var number int32
			if err := json.Unmarshal(fields[name], &number); err != nil {
				t.Fatalf("enum %s must remain a raw numeric value: %s", name, fields[name])
			}
		}
	}
	if session, ok := want.(*gen.SessionMetric); ok {
		var hops []json.RawMessage
		if err := json.Unmarshal(fields["link_hops"], &hops); err != nil || len(hops) != len(session.LinkHops) {
			t.Fatalf("nested hops changed: count=%d err=%v", len(hops), err)
		}
		for index, hop := range session.LinkHops {
			assertReportPayloadRoundTrip(t, string(hops[index]), hop)
		}
	}
}

// TestEncodeReportPayloadRoundTrip verifies all fields, exact int64s, extra and unknown enum values.
func TestEncodeReportPayloadRoundTrip(t *testing.T) {
	fixture := newReportPayloadFixture("payload-unit")
	for _, metric := range []proto.Message{fixture.machine, fixture.network, fixture.stream, fixture.session} {
		t.Run(string(metric.ProtoReflect().Descriptor().Name()), func(t *testing.T) {
			payload, err := encodeReportPayload(metric)
			if err != nil {
				t.Fatalf("encode full report: %v", err)
			}
			assertReportPayloadRoundTrip(t, payload, metric)
			zero := metric.ProtoReflect().New().Interface()
			payload, err = encodeReportPayload(zero)
			if err != nil {
				t.Fatalf("encode zero report: %v", err)
			}
			assertReportPayloadRoundTrip(t, payload, zero)
		})
	}
}

// TestEncodeReportPayloadRejectsInvalidText ensures serialization errors never become empty snapshots.
func TestEncodeReportPayloadRejectsInvalidText(t *testing.T) {
	for _, metric := range []proto.Message{
		&gen.MachineMetric{Extra: map[string]string{"invalid": "\xff"}},
		&gen.NetworkMetric{Extra: map[string]string{"invalid": "\xff"}},
		&gen.StreamMetric{Extra: map[string]string{"invalid": "\xff"}},
		&gen.SessionMetric{Extra: map[string]string{"invalid": "\xff"}},
	} {
		t.Run(string(metric.ProtoReflect().Descriptor().Name()), func(t *testing.T) {
			if payload, err := encodeReportPayload(metric); err == nil || payload != "" {
				t.Fatalf("expected empty result and serialization error, got payload=%q err=%v", payload, err)
			}
		})
	}
}

// TestReportPayloadPersistencePostgres verifies complete latest reports through the actual handlers.
func TestReportPayloadPersistencePostgres(t *testing.T) {
	if os.Getenv("LINAPRO_TEST_POSTGRES") != "1" {
		t.Skip("set LINAPRO_TEST_POSTGRES=1 to run against a local PostgreSQL database")
	}
	var (
		ctx             = t.Context()
		prefix          = "payload-" + rand.Text()[:10]
		fixture         = newReportPayloadFixture(prefix)
		writer          = newReportRuntime(newMemoryCollectionCache())
		firstStreamID   = prefix + "-first-stream"
		firstSessionID  = prefix + "-first-session"
		failedStreamID  = prefix + "-failed-stream"
		failedSessionID = prefix + "-failed-session"
		invalidID       = prefix + "-invalid"
	)
	setupCollectionReportPostgres(t, ctx)
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		cleanupCollectionReportRows(t, cleanupCtx, fixture.network.NodeId, fixture.machine.InstanceId, fixture.stream.StreamId, fixture.session.SessionId)
		cleanupCollectionReportRows(t, cleanupCtx, fixture.network.NodeId, invalidID, firstStreamID, firstSessionID)
		cleanupCollectionReportRows(t, cleanupCtx, fixture.network.NodeId, "", failedStreamID, failedSessionID)
	})
	readMachine := func() *entity.MediaReportInstance {
		t.Helper()
		var row *entity.MediaReportInstance
		if err := dao.MediaReportInstance.Ctx(ctx).Where(do.MediaReportInstance{InstanceId: fixture.machine.InstanceId}).Scan(&row); err != nil || row == nil {
			t.Fatalf("read machine payload row: row=%v err=%v", row, err)
		}
		return row
	}
	readNetwork := func() *entity.ReportNetworkLink {
		t.Helper()
		var row *entity.ReportNetworkLink
		if err := dao.ReportNetworkLink.Ctx(ctx).Where(do.ReportNetworkLink{NodeId: fixture.network.NodeId}).Scan(&row); err != nil || row == nil {
			t.Fatalf("read network payload row: row=%v err=%v", row, err)
		}
		return row
	}
	readStream := func(streamID string) *entity.MediaReportStream {
		t.Helper()
		var row *entity.MediaReportStream
		if err := dao.MediaReportStream.Ctx(ctx).Where(do.MediaReportStream{StreamId: streamID}).Scan(&row); err != nil || row == nil {
			t.Fatalf("read stream payload row: row=%v err=%v", row, err)
		}
		return row
	}
	readSession := func(sessionID string) *entity.MediaReportSession {
		t.Helper()
		var row *entity.MediaReportSession
		if err := dao.MediaReportSession.Ctx(ctx).Where(do.MediaReportSession{SessionId: sessionID}).Scan(&row); err != nil || row == nil {
			t.Fatalf("read session payload row: row=%v err=%v", row, err)
		}
		return row
	}
	if err := writer.HandleMachineMetric(ctx, fixture.machine); err != nil {
		t.Fatalf("persist full machine report: %v", err)
	}
	assertReportPayloadRoundTrip(t, readMachine().ReportPayload, fixture.machine)
	fixture.machine.Timestamp++
	fixture.machine.Extra["vendor"] = "updated machine"
	if err := writer.HandleMachineMetric(ctx, fixture.machine); err != nil {
		t.Fatalf("update machine payload: %v", err)
	}
	assertReportPayloadRoundTrip(t, readMachine().ReportPayload, fixture.machine)
	if err := writer.HandleNetworkMetric(ctx, fixture.network); err != nil {
		t.Fatalf("persist full network report: %v", err)
	}
	assertReportPayloadRoundTrip(t, readNetwork().ReportPayload, fixture.network)
	stale := proto.Clone(fixture.network).(*gen.NetworkMetric)
	fixture.network.Timestamp++
	fixture.network.Rtt, fixture.network.Jitter, fixture.network.PacketLoss = 0, 0, 0
	fixture.network.Extra["vendor"] = "updated network"
	if err := writer.HandleNetworkMetric(ctx, fixture.network); err != nil {
		t.Fatalf("update network payload: %v", err)
	}
	assertReportPayloadRoundTrip(t, readNetwork().ReportPayload, fixture.network)
	if err := writer.HandleNetworkMetric(ctx, stale); err != nil {
		t.Fatalf("handle stale network report: %v", err)
	}
	assertReportPayloadRoundTrip(t, readNetwork().ReportPayload, fixture.network)
	if err := writer.HandleStreamMetric(ctx, uint8(gen.SCMDDataReport_STREAM_ADD), fixture.stream); err != nil {
		t.Fatalf("persist full stream add: %v", err)
	}
	assertReportPayloadRoundTrip(t, readStream(fixture.stream.StreamId).ReportPayload, fixture.stream)
	fixture.stream.Timestamp++
	fixture.stream.StreamAlias = "Updated stream alias"
	if err := writer.HandleStreamMetric(ctx, uint8(gen.SCMDDataReport_STREAM_STATUS), fixture.stream); err != nil {
		t.Fatalf("update stream payload: %v", err)
	}
	streamBeforeDelete := readStream(fixture.stream.StreamId)
	assertReportPayloadRoundTrip(t, streamBeforeDelete.ReportPayload, fixture.stream)
	if streamBeforeDelete.ReportSubcommand != int(gen.SCMDDataReport_STREAM_STATUS) {
		t.Fatalf("stream subcommand was not updated: %d", streamBeforeDelete.ReportSubcommand)
	}
	if err := writer.HandleSessionMetric(ctx, uint8(gen.SCMDDataReport_SESSION_ADD), fixture.session); err != nil {
		t.Fatalf("persist full session add: %v", err)
	}
	assertReportPayloadRoundTrip(t, readSession(fixture.session.SessionId).ReportPayload, fixture.session)
	fixture.session.Timestamp++
	fixture.session.UserName = "Updated user"
	if err := writer.HandleSessionMetric(ctx, uint8(gen.SCMDDataReport_SESSION_STATUS), fixture.session); err != nil {
		t.Fatalf("update session payload: %v", err)
	}
	sessionBeforeDelete := readSession(fixture.session.SessionId)
	assertReportPayloadRoundTrip(t, sessionBeforeDelete.ReportPayload, fixture.session)
	if sessionBeforeDelete.ReportSubcommand != int(gen.SCMDDataReport_SESSION_STATUS) {
		t.Fatalf("session subcommand was not updated: %d", sessionBeforeDelete.ReportSubcommand)
	}
	assertReportInstanceCounters(t, ctx, fixture.machine.InstanceId, 1, 1)
	assertReportPayloadRoundTrip(t, readMachine().ReportPayload, fixture.machine)
	streamDelete := &gen.StreamMetric{StreamId: fixture.stream.StreamId, Timestamp: fixture.stream.Timestamp + 1, Extra: map[string]string{"reason": "delete-only"}}
	if err := writer.HandleStreamMetric(ctx, uint8(gen.SCMDDataReport_STREAM_DELETE), streamDelete); err != nil {
		t.Fatalf("persist existing stream delete: %v", err)
	}
	streamAfterDelete := readStream(streamDelete.StreamId)
	assertReportPayloadRoundTrip(t, streamAfterDelete.ReportPayload, streamDelete)
	if streamAfterDelete.ReportSubcommand != int(gen.SCMDDataReport_STREAM_DELETE) || streamAfterDelete.CloseTime == nil || streamAfterDelete.Status != string(reportStreamStatusClosed) ||
		streamAfterDelete.StreamName != streamBeforeDelete.StreamName || streamAfterDelete.TenantId != streamBeforeDelete.TenantId ||
		streamAfterDelete.InstanceId != streamBeforeDelete.InstanceId || streamAfterDelete.Resolution != streamBeforeDelete.Resolution || streamAfterDelete.Bitrate != streamBeforeDelete.Bitrate {
		t.Fatalf("stream delete erased display data or failed to close: %#v", streamAfterDelete)
	}
	sessionDelete := &gen.SessionMetric{SessionId: fixture.session.SessionId, Timestamp: fixture.session.Timestamp + 1, Extra: map[string]string{"reason": "delete-only"}}
	if err := writer.HandleSessionMetric(ctx, uint8(gen.SCMDDataReport_SESSION_DELETE), sessionDelete); err != nil {
		t.Fatalf("persist existing session delete: %v", err)
	}
	sessionAfterDelete := readSession(sessionDelete.SessionId)
	assertReportPayloadRoundTrip(t, sessionAfterDelete.ReportPayload, sessionDelete)
	if sessionAfterDelete.ReportSubcommand != int(gen.SCMDDataReport_SESSION_DELETE) || sessionAfterDelete.CloseTime == nil ||
		sessionAfterDelete.StreamId != sessionBeforeDelete.StreamId || sessionAfterDelete.ClientIp != sessionBeforeDelete.ClientIp ||
		sessionAfterDelete.UserName != sessionBeforeDelete.UserName || sessionAfterDelete.InstanceId != sessionBeforeDelete.InstanceId || sessionAfterDelete.LinkHops != sessionBeforeDelete.LinkHops {
		t.Fatalf("session delete erased display data or failed to close: %#v", sessionAfterDelete)
	}
	assertReportInstanceCounters(t, ctx, fixture.machine.InstanceId, 0, 0)
	assertReportPayloadRoundTrip(t, readMachine().ReportPayload, fixture.machine)
	streamDelete.StreamId = firstStreamID
	if err := writer.HandleStreamMetric(ctx, uint8(gen.SCMDDataReport_STREAM_DELETE), streamDelete); err != nil {
		t.Fatalf("persist first-seen stream delete: %v", err)
	}
	firstStream := readStream(firstStreamID)
	assertReportPayloadRoundTrip(t, firstStream.ReportPayload, streamDelete)
	if firstStream.CloseTime == nil || firstStream.Status != string(reportStreamStatusClosed) || firstStream.ReportSubcommand != int(gen.SCMDDataReport_STREAM_DELETE) {
		t.Fatalf("first-seen stream delete must create a closed payload row: %#v", firstStream)
	}
	sessionDelete.SessionId = firstSessionID
	if err := writer.HandleSessionMetric(ctx, uint8(gen.SCMDDataReport_SESSION_DELETE), sessionDelete); err != nil {
		t.Fatalf("persist first-seen session delete: %v", err)
	}
	firstSession := readSession(firstSessionID)
	assertReportPayloadRoundTrip(t, firstSession.ReportPayload, sessionDelete)
	if firstSession.CloseTime == nil || firstSession.ReportSubcommand != int(gen.SCMDDataReport_SESSION_DELETE) {
		t.Fatalf("first-seen session delete must create a closed payload row: %#v", firstSession)
	}
	// Payload storage precedes lifecycle counters. Even if counters are unavailable,
	// first-seen delete reports must remain complete and must not create active shells.
	withoutCounters := newReportRuntime(nil)
	streamDelete.StreamId = failedStreamID
	if err := withoutCounters.HandleStreamMetric(ctx, uint8(gen.SCMDDataReport_STREAM_DELETE), streamDelete); err == nil {
		t.Fatal("expected unavailable stream counter backend to fail")
	}
	failedStream := readStream(failedStreamID)
	assertReportPayloadRoundTrip(t, failedStream.ReportPayload, streamDelete)
	if failedStream.CloseTime == nil || failedStream.Status != string(reportStreamStatusClosed) {
		t.Fatalf("counter failure left an active stream shell: %#v", failedStream)
	}
	sessionDelete.SessionId = failedSessionID
	if err := withoutCounters.HandleSessionMetric(ctx, uint8(gen.SCMDDataReport_SESSION_DELETE), sessionDelete); err == nil {
		t.Fatal("expected unavailable session counter backend to fail")
	}
	failedSession := readSession(failedSessionID)
	assertReportPayloadRoundTrip(t, failedSession.ReportPayload, sessionDelete)
	if failedSession.CloseTime == nil {
		t.Fatalf("counter failure left an active session shell: %#v", failedSession)
	}
	if err := writer.HandleStreamMetric(ctx, uint8(gen.SCMDDataReport_STREAM_DELETE), streamDelete); err != nil {
		t.Fatalf("retry stream delete after counter recovery: %v", err)
	}
	assertReportPayloadRoundTrip(t, readStream(failedStreamID).ReportPayload, streamDelete)
	if err := writer.HandleSessionMetric(ctx, uint8(gen.SCMDDataReport_SESSION_DELETE), sessionDelete); err != nil {
		t.Fatalf("retry session delete after counter recovery: %v", err)
	}
	assertReportPayloadRoundTrip(t, readSession(failedSessionID).ReportPayload, sessionDelete)
	if err := writer.HandleMachineMetric(ctx, &gen.MachineMetric{InstanceId: invalidID, Extra: map[string]string{"invalid": "\xff"}}); err == nil {
		t.Fatal("invalid protobuf text must fail before database writes")
	}
	count, err := dao.MediaReportInstance.Ctx(ctx).Where(do.MediaReportInstance{InstanceId: invalidID}).Count()
	if err != nil || count != 0 {
		t.Fatalf("serialization failure wrote an instance row: count=%d err=%v", count, err)
	}
	// Rows created without a raw report keep SQL NULL, not a fabricated empty JSON payload.
	if _, err := dao.MediaReportInstance.Ctx(ctx).Data(do.MediaReportInstance{
		InstanceId: invalidID, ReportTime: int64(1_790_000_000_123),
	}).Insert(); err != nil {
		t.Fatalf("insert legacy-style instance without payload: %v", err)
	}
	count, err = dao.MediaReportInstance.Ctx(ctx).Where(do.MediaReportInstance{InstanceId: invalidID}).
		WhereNull(dao.MediaReportInstance.Columns().ReportPayload).Count()
	if err != nil || count != 1 {
		t.Fatalf("unreported payload must remain SQL NULL: count=%d err=%v", count, err)
	}
}
