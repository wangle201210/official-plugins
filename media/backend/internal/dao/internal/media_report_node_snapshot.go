// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MediaReportNodeSnapshotDao is the data access object for the table media_report_node_snapshot.
type MediaReportNodeSnapshotDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  MediaReportNodeSnapshotColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// MediaReportNodeSnapshotColumns defines and stores column names for the table media_report_node_snapshot.
type MediaReportNodeSnapshotColumns struct {
	NodeId          string // 上报节点业务标识，不依赖配置表外键
	ReportTime      string // 上报端采样时间，也是快照时间
	NodeName        string // 节点展示名称，按快照时间点冗余
	Region          string // 节点所属区域或机房名称
	ParentNodeId    string // 父节点业务标识，根节点使用0
	Status          string // 节点运行状态，保存上报原始枚举值
	TotalNodes      string // 快照时节点视角下的节点总数
	AliveNodes      string // 快照时节点视角下的在线节点数
	CpuAllocated    string // 快照时已分配CPU容量，单位由上报端统一
	CpuLoad         string // 快照时CPU负载或利用率，单位由上报端统一
	MemoryAllocated string // 快照时已分配内存容量，单位GB
	MemoryUsed      string // 快照时已使用内存容量，单位GB
	DiskIoRead      string // 快照时磁盘读取速率，单位MB/s
	DiskIoWrite     string // 快照时磁盘写入速率，单位MB/s
	NetworkIn       string // 快照时网络入站速率，单位Mbps
	NetworkOut      string // 快照时网络出站速率，单位Mbps
	LiveStreams     string // 快照时直播流数量
	Sessions        string // 快照时会话数量
	AvgDelay        string // 快照时节点平均延迟，单位毫秒
	LastHeartbeat   string // 快照时节点最后心跳时间
	NodeLatencyMap  string // 快照时节点延迟矩阵JSON，结构来自接口node_latency_map
}

// mediaReportNodeSnapshotColumns holds the columns for the table media_report_node_snapshot.
var mediaReportNodeSnapshotColumns = MediaReportNodeSnapshotColumns{
	NodeId:          "node_id",
	ReportTime:      "report_time",
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
}

// NewMediaReportNodeSnapshotDao creates and returns a new DAO object for table data access.
func NewMediaReportNodeSnapshotDao(handlers ...gdb.ModelHandler) *MediaReportNodeSnapshotDao {
	return &MediaReportNodeSnapshotDao{
		group:    "default",
		table:    "media_report_node_snapshot",
		columns:  mediaReportNodeSnapshotColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MediaReportNodeSnapshotDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MediaReportNodeSnapshotDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MediaReportNodeSnapshotDao) Columns() MediaReportNodeSnapshotColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MediaReportNodeSnapshotDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MediaReportNodeSnapshotDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MediaReportNodeSnapshotDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
