// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MediaReportInstanceDao is the data access object for the table media_report_instance.
type MediaReportInstanceDao struct {
	table    string                     // table is the underlying table name of the DAO.
	group    string                     // group is the database configuration group name of the current DAO.
	columns  MediaReportInstanceColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler         // handlers for customized model modification.
}

// MediaReportInstanceColumns defines and stores column names for the table media_report_instance.
type MediaReportInstanceColumns struct {
	InstanceId      string // 实例业务标识，不依赖实例配置外键
	InstanceName    string // 实例展示名称，按上报时间点冗余
	NodeId          string // 实例所属节点业务标识
	NodeName        string // 实例所属节点展示名称，按上报时间点冗余
	Region          string // 实例所属区域或机房名称
	NodeStatus      string // 实例所属节点状态，用于返回node_info.status
	Status          string // 实例运行状态，保存上报原始枚举值
	CpuAllocated    string // 实例已分配CPU容量，单位由上报端统一
	CpuLoad         string // 实例CPU负载或利用率，单位由上报端统一
	MemoryAllocated string // 实例已分配内存容量，单位GB
	MemoryUsed      string // 实例已使用内存容量，单位GB
	DiskIoRead      string // 实例磁盘读取速率，单位MB/s
	DiskIoWrite     string // 实例磁盘写入速率，单位MB/s
	NetworkIn       string // 实例网络入站速率，单位Mbps
	NetworkOut      string // 实例网络出站速率，单位Mbps
	LiveStreams     string // 实例当前承载直播流数量
	Sessions        string // 实例当前会话数量
	StartTime       string // 实例启动时间
	Version         string // 实例版本号
	ReportTime      string // 上报端采样时间
	UpdatedAt       string // 记录更新时间
}

// mediaReportInstanceColumns holds the columns for the table media_report_instance.
var mediaReportInstanceColumns = MediaReportInstanceColumns{
	InstanceId:      "instance_id",
	InstanceName:    "instance_name",
	NodeId:          "node_id",
	NodeName:        "node_name",
	Region:          "region",
	NodeStatus:      "node_status",
	Status:          "status",
	CpuAllocated:    "cpu_allocated",
	CpuLoad:         "cpu_load",
	MemoryAllocated: "memory_allocated",
	MemoryUsed:      "memory_used",
	DiskIoRead:      "disk_io_read",
	DiskIoWrite:     "disk_io_write",
	NetworkIn:       "network_in",
	NetworkOut:      "network_out",
	LiveStreams:     "live_streams",
	Sessions:        "sessions",
	StartTime:       "start_time",
	Version:         "version",
	ReportTime:      "report_time",
	UpdatedAt:       "updated_at",
}

// NewMediaReportInstanceDao creates and returns a new DAO object for table data access.
func NewMediaReportInstanceDao(handlers ...gdb.ModelHandler) *MediaReportInstanceDao {
	return &MediaReportInstanceDao{
		group:    "default",
		table:    "media_report_instance",
		columns:  mediaReportInstanceColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MediaReportInstanceDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MediaReportInstanceDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MediaReportInstanceDao) Columns() MediaReportInstanceColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MediaReportInstanceDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MediaReportInstanceDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MediaReportInstanceDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
