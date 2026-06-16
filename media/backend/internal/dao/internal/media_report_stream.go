// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MediaReportStreamDao is the data access object for the table media_report_stream.
type MediaReportStreamDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  MediaReportStreamColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// MediaReportStreamColumns defines and stores column names for the table media_report_stream.
type MediaReportStreamColumns struct {
	StreamId              string // 流业务标识，不依赖流配置外键
	SourceType            string // 来源类型，例如节点或实例
	TenantId              string // 租户标识，用于数据权限过滤和租户统计
	NodeId                string // 流所属节点业务标识
	NodeName              string // 流所属节点展示名称，按上报时间点冗余
	InstanceId            string // 流所属实例业务标识
	InstanceName          string // 流所属实例展示名称，按上报时间点冗余
	SourceUrl             string // 流源地址
	StreamName            string // 流展示名称
	Resolution            string // 当前分辨率
	Fps                   string // 当前帧率
	Bitrate               string // 当前码率，单位Kbps
	PacketLoss            string // 当前丢包率
	Status                string // 流运行状态，保存上报原始枚举值
	StartTime             string // 流开始时间
	Duration              string // 上报端流持续时间，接口返回时优先通过start_time和close_time动态计算
	AvgDelay              string // 流平均延迟，单位毫秒
	ProtocolCount         string // 支持的协议数量
	TotalSessionsLifetime string // 流历史累计会话数量
	CurrentActiveSessions string // 流当前活跃会话数量
	WatermarkEnabled      string // 当前是否启用水印
	ProtocolSummary       string // 协议摘要JSON数组，结构来自接口protocol_summary
	ReportTime            string // 上报端采样时间
	UpdatedAt             string // 记录更新时间
	ProtocolType          string // 源流协议类型
	CloseTime             string // 流关闭时间，未关闭时为空
}

// mediaReportStreamColumns holds the columns for the table media_report_stream.
var mediaReportStreamColumns = MediaReportStreamColumns{
	StreamId:              "stream_id",
	SourceType:            "source_type",
	TenantId:              "tenant_id",
	NodeId:                "node_id",
	NodeName:              "node_name",
	InstanceId:            "instance_id",
	InstanceName:          "instance_name",
	SourceUrl:             "source_url",
	StreamName:            "stream_name",
	Resolution:            "resolution",
	Fps:                   "fps",
	Bitrate:               "bitrate",
	PacketLoss:            "packet_loss",
	Status:                "status",
	StartTime:             "start_time",
	Duration:              "duration",
	AvgDelay:              "avg_delay",
	ProtocolCount:         "protocol_count",
	TotalSessionsLifetime: "total_sessions_lifetime",
	CurrentActiveSessions: "current_active_sessions",
	WatermarkEnabled:      "watermark_enabled",
	ProtocolSummary:       "protocol_summary",
	ReportTime:            "report_time",
	UpdatedAt:             "updated_at",
	ProtocolType:          "protocol_type",
	CloseTime:             "close_time",
}

// NewMediaReportStreamDao creates and returns a new DAO object for table data access.
func NewMediaReportStreamDao(handlers ...gdb.ModelHandler) *MediaReportStreamDao {
	return &MediaReportStreamDao{
		group:    "default",
		table:    "media_report_stream",
		columns:  mediaReportStreamColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MediaReportStreamDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MediaReportStreamDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MediaReportStreamDao) Columns() MediaReportStreamColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MediaReportStreamDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MediaReportStreamDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *MediaReportStreamDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
