// This file preserves complete known protobuf report fields alongside dashboard
// projections. Delete payloads are retained without blanking existing display data.

package collection

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
)

// encodeReportPayload keeps original names, default values, raw enum numbers,
// int64 precision, extension maps and nested messages. Unknown wire fields are
// outside this known-field JSON contract; no normalized projection is used here.
func encodeReportPayload(metric proto.Message) (string, error) {
	content, err := (protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
		UseEnumNumbers:  true,
	}).Marshal(metric)
	if err != nil {
		return "", gerror.Wrap(err, "encode complete media report payload failed")
	}
	return string(content), nil
}

// storeDeletedStreamPayload stores the full delete report before lifecycle work.
// First-seen deletes create closed rows; existing rows retain their display fields.
func storeDeletedStreamPayload(ctx context.Context, report streamReport) error {
	columns := dao.MediaReportStream.Columns()
	_, err := dao.MediaReportStream.Ctx(ctx).
		Data(do.MediaReportStream{
			StreamId:         report.streamID,
			Status:           string(reportStreamStatusClosed),
			CloseTime:        reportTimeToGTime(normalizeReportTime(report.reportTime)),
			ReportTime:       report.reportTime,
			ReportPayload:    report.payload,
			ReportSubcommand: report.subcommand,
		}).
		OnConflict(columns.StreamId).
		OnDuplicate(columns.ReportPayload, columns.ReportSubcommand).
		Save()
	return err
}

// storeDeletedSessionPayload preserves delete content even when no add was seen.
// Existing session data stays intact, and newly inserted rows are already closed.
func storeDeletedSessionPayload(ctx context.Context, report sessionReport) error {
	columns := dao.MediaReportSession.Columns()
	_, err := dao.MediaReportSession.Ctx(ctx).
		Data(do.MediaReportSession{
			SessionId:        report.sessionID,
			CloseTime:        reportTimeToGTime(normalizeReportTime(report.reportTime)),
			ReportTime:       report.reportTime,
			ReportPayload:    report.payload,
			ReportSubcommand: report.subcommand,
		}).
		OnConflict(columns.SessionId).
		OnDuplicate(columns.ReportPayload, columns.ReportSubcommand).
		Save()
	return err
}
