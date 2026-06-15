// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MediaReportSessionDao is the data access object for the table media_report_session.
type MediaReportSessionDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  MediaReportSessionColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// MediaReportSessionColumns defines and stores column names for the table media_report_session.
type MediaReportSessionColumns struct {
	SessionId         string // 会话业务标识
	StreamId          string // 会话所属流业务标识
	StreamName        string // 会话所属流展示名称，按上报时间点冗余
	TenantId          string // 租户标识，用于数据权限过滤和租户统计
	ClientId          string // 客户端标识
	ClientIp          string // 客户端IP地址
	ClientType        string // 客户端类型
	UserName          string // 播放用户展示名称
	ProtocolType      string // 播放协议类型
	StartTime         string // 会话开始时间
	PlayDuration      string // 播放持续时间，单位秒
	CurrentFps        string // 当前播放帧率
	CurrentBitrate    string // 当前播放码率，单位Kbps
	CurrentResolution string // 当前播放分辨率
	NodeId            string // 会话承载节点业务标识
	NodeName          string // 会话承载节点展示名称，按上报时间点冗余
	InstanceId        string // 会话承载实例业务标识
	InstanceName      string // 会话承载实例展示名称，按上报时间点冗余
	LinkHops          string // 会话链路跳点JSON数组，结构来自接口link_hops
	TotalLinkLatency  string // 会话链路总延迟，单位毫秒
	ReportTime        string // 上报端采样时间
	UpdatedAt         string // 记录更新时间
}

// mediaReportSessionColumns holds the columns for the table media_report_session.
var mediaReportSessionColumns = MediaReportSessionColumns{
	SessionId:         "session_id",
	StreamId:          "stream_id",
	StreamName:        "stream_name",
	TenantId:          "tenant_id",
	ClientId:          "client_id",
	ClientIp:          "client_ip",
	ClientType:        "client_type",
	UserName:          "user_name",
	ProtocolType:      "protocol_type",
	StartTime:         "start_time",
	PlayDuration:      "play_duration",
	CurrentFps:        "current_fps",
	CurrentBitrate:    "current_bitrate",
	CurrentResolution: "current_resolution",
	NodeId:            "node_id",
	NodeName:          "node_name",
	InstanceId:        "instance_id",
	InstanceName:      "instance_name",
	LinkHops:          "link_hops",
	TotalLinkLatency:  "total_link_latency",
	ReportTime:        "report_time",
	UpdatedAt:         "updated_at",
}

// NewMediaReportSessionDao creates and returns a new DAO object for table data access.
func NewMediaReportSessionDao(handlers ...gdb.ModelHandler) *MediaReportSessionDao {
	return &MediaReportSessionDao{
		group:    "default",
		table:    "media_report_session",
		columns:  mediaReportSessionColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MediaReportSessionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MediaReportSessionDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MediaReportSessionDao) Columns() MediaReportSessionColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MediaReportSessionDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MediaReportSessionDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MediaReportSessionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
