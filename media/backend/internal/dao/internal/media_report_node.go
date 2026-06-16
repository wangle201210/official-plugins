// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MediaReportNodeDao is the data access object for the table media_report_node.
type MediaReportNodeDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  MediaReportNodeColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// MediaReportNodeColumns defines and stores column names for the table media_report_node.
type MediaReportNodeColumns struct {
	NodeId          string // 上报节点业务标识，不依赖配置表外键
	NodeName        string // 节点展示名称，按上报时间点冗余
	Region          string // 节点所属区域或机房名称
	ParentNodeId    string // 父节点业务标识，根节点使用0
	Status          string // 节点运行状态，保存上报原始枚举值
	TotalNodes      string // 当前节点视角下的节点总数
	AliveNodes      string // 当前节点视角下的在线节点数
	CpuAllocated    string // 已分配CPU容量，单位由上报端统一
	CpuLoad         string // CPU负载或利用率，单位由上报端统一
	MemoryAllocated string // 已分配内存容量，单位MB
	MemoryUsed      string // 已使用内存容量，单位MB
	DiskIoRead      string // 磁盘读取速率，单位KB/S
	DiskIoWrite     string // 磁盘写入速率，单位KB/S
	NetworkIn       string // 网络入站速率，单位KB/S
	NetworkOut      string // 网络出站速率，单位KB/S
	LiveStreams     string // 当前直播流数量
	Sessions        string // 当前会话数量
	AvgDelay        string // 节点平均延迟，单位毫秒
	LastHeartbeat   string // 节点最后心跳时间
	NodeLatencyMap  string // 节点到其他节点的延迟矩阵JSON，结构来自接口node_latency_map
	ReportTime      string // 上报端采样时间
	UpdatedAt       string // 记录更新时间
}

// mediaReportNodeColumns holds the columns for the table media_report_node.
var mediaReportNodeColumns = MediaReportNodeColumns{
	NodeId:          "node_id",
	NodeName:        "node_name",
	Region:          "region",
	ParentNodeId:    "parent_node_id",
	Status:          "status",
	TotalNodes:      "total_nodes",
	AliveNodes:      "alive_nodes",
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
	AvgDelay:        "avg_delay",
	LastHeartbeat:   "last_heartbeat",
	NodeLatencyMap:  "node_latency_map",
	ReportTime:      "report_time",
	UpdatedAt:       "updated_at",
}

// NewMediaReportNodeDao creates and returns a new DAO object for table data access.
func NewMediaReportNodeDao(handlers ...gdb.ModelHandler) *MediaReportNodeDao {
	return &MediaReportNodeDao{
		group:    "default",
		table:    "media_report_node",
		columns:  mediaReportNodeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MediaReportNodeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MediaReportNodeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MediaReportNodeDao) Columns() MediaReportNodeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MediaReportNodeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MediaReportNodeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MediaReportNodeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
