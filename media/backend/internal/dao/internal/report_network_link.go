// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ReportNetworkLinkDao is the data access object for the table plugin_media_report_network_link.
type ReportNetworkLinkDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  ReportNetworkLinkColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// ReportNetworkLinkColumns defines and stores column names for the table plugin_media_report_network_link.
type ReportNetworkLinkColumns struct {
	NodeId        string // 上报所属节点业务标识
	MachineId     string // 上报来源机器或实例标识
	SourceIp      string // 原始探测源地址
	DestinationIp string // 原始探测目的地址
	Rtt           string // 原始往返延迟，单位毫秒，保留探测失败固定值
	Jitter        string // 原始抖动，单位毫秒，保留探测失败固定值
	PacketLoss    string // 原始丢包比例0至1，当前上游口径为HTTP探测失败比例
	StatusCode    string // 原始探测HTTP状态码
	Reachable     string // 原始extra.reachable布尔状态，未提供或无法解析时为空
	ReportTime    string // 上报端采样时间，Unix毫秒时间戳
	ReportPayload string // 未经业务归一化的最新完整NetworkMetric protobuf JSON，保存全部已定义字段；旧行未接收新报文时为NULL
}

// reportNetworkLinkColumns holds the columns for the table plugin_media_report_network_link.
var reportNetworkLinkColumns = ReportNetworkLinkColumns{
	NodeId:        "node_id",
	MachineId:     "machine_id",
	SourceIp:      "source_ip",
	DestinationIp: "destination_ip",
	Rtt:           "rtt",
	Jitter:        "jitter",
	PacketLoss:    "packet_loss",
	StatusCode:    "status_code",
	Reachable:     "reachable",
	ReportTime:    "report_time",
	ReportPayload: "report_payload",
}

// NewReportNetworkLinkDao creates and returns a new DAO object for table data access.
func NewReportNetworkLinkDao(handlers ...gdb.ModelHandler) *ReportNetworkLinkDao {
	return &ReportNetworkLinkDao{
		group:    "default",
		table:    "plugin_media_report_network_link",
		columns:  reportNetworkLinkColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ReportNetworkLinkDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ReportNetworkLinkDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ReportNetworkLinkDao) Columns() ReportNetworkLinkColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ReportNetworkLinkDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ReportNetworkLinkDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ReportNetworkLinkDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
