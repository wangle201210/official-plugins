-- ------------------------------------------------------------
-- 003 add media data collection server SQL file
-- Purpose: Stores media collection report projections for dashboard read models.
-- 用途：存储媒体采集上报数据看板读模型。
-- ------------------------------------------------------------

-- Purpose: Stores the latest node overview projection reported by collection clients.
-- 用途：存储采集客户端上报的节点总览最新投影。
CREATE TABLE IF NOT EXISTS media_report_node (
    "node_id" VARCHAR(64) PRIMARY KEY,
    "node_name" VARCHAR(128) NOT NULL DEFAULT '',
    "region" VARCHAR(128) NOT NULL DEFAULT '',
    "parent_node_id" VARCHAR(64) NOT NULL DEFAULT '0',
    "status" VARCHAR(32) NOT NULL DEFAULT '',
    "total_nodes" INTEGER NOT NULL DEFAULT 0,
    "alive_nodes" INTEGER NOT NULL DEFAULT 0,
    "cpu_allocated" NUMERIC(12,2) NOT NULL DEFAULT 0,
    "cpu_load" NUMERIC(12,2) NOT NULL DEFAULT 0,
    "memory_allocated" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "memory_used" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "disk_io_read" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "disk_io_write" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "network_in" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "network_out" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "live_streams" INTEGER NOT NULL DEFAULT 0,
    "sessions" INTEGER NOT NULL DEFAULT 0,
    "avg_delay" INTEGER NOT NULL DEFAULT 0,
    "last_heartbeat" TIMESTAMP,
    "node_latency_map" JSONB NOT NULL DEFAULT '{}'::JSONB,
    "report_time" BIGINT NOT NULL,
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_media_report_node_parent
    ON media_report_node ("parent_node_id");

CREATE INDEX IF NOT EXISTS idx_media_report_node_status
    ON media_report_node ("status");

CREATE INDEX IF NOT EXISTS idx_media_report_node_report_time
    ON media_report_node ("report_time" DESC);

COMMENT ON TABLE media_report_node IS '媒体看板节点总览表，保存节点概览最新上报投影';
COMMENT ON COLUMN media_report_node."node_id" IS '上报节点业务标识，不依赖配置表外键';
COMMENT ON COLUMN media_report_node."node_name" IS '节点展示名称，按上报时间点冗余';
COMMENT ON COLUMN media_report_node."region" IS '节点所属区域或机房名称';
COMMENT ON COLUMN media_report_node."parent_node_id" IS '父节点业务标识，根节点使用0';
COMMENT ON COLUMN media_report_node."status" IS '节点运行状态，保存上报原始枚举值';
COMMENT ON COLUMN media_report_node."total_nodes" IS '当前节点视角下的节点总数';
COMMENT ON COLUMN media_report_node."alive_nodes" IS '当前节点视角下的在线节点数';
COMMENT ON COLUMN media_report_node."cpu_allocated" IS '已分配CPU容量，单位由上报端统一';
COMMENT ON COLUMN media_report_node."cpu_load" IS 'CPU负载或利用率，单位由上报端统一';
COMMENT ON COLUMN media_report_node."memory_allocated" IS '已分配内存容量，单位MB';
COMMENT ON COLUMN media_report_node."memory_used" IS '已使用内存容量，单位MB';
COMMENT ON COLUMN media_report_node."disk_io_read" IS '磁盘读取速率，单位KB/S';
COMMENT ON COLUMN media_report_node."disk_io_write" IS '磁盘写入速率，单位KB/S';
COMMENT ON COLUMN media_report_node."network_in" IS '网络入站速率，单位KB/S';
COMMENT ON COLUMN media_report_node."network_out" IS '网络出站速率，单位KB/S';
COMMENT ON COLUMN media_report_node."live_streams" IS '当前直播流数量';
COMMENT ON COLUMN media_report_node."sessions" IS '当前会话数量';
COMMENT ON COLUMN media_report_node."avg_delay" IS '节点平均延迟，单位毫秒';
COMMENT ON COLUMN media_report_node."last_heartbeat" IS '节点最后心跳时间';
COMMENT ON COLUMN media_report_node."node_latency_map" IS '节点到其他节点的延迟矩阵JSON，结构来自接口node_latency_map';
COMMENT ON COLUMN media_report_node."report_time" IS '上报端采样时间';
COMMENT ON COLUMN media_report_node."updated_at" IS '记录更新时间';

-- Purpose: Stores idempotent node overview snapshots for trend and replay.
-- 用途：存储节点总览历史采样，用于趋势和回放。
CREATE TABLE IF NOT EXISTS media_report_node_snapshot (
    "node_id" VARCHAR(64) NOT NULL,
    "report_time" BIGINT NOT NULL,
    "node_name" VARCHAR(128) NOT NULL DEFAULT '',
    "region" VARCHAR(128) NOT NULL DEFAULT '',
    "parent_node_id" VARCHAR(64) NOT NULL DEFAULT '0',
    "status" VARCHAR(32) NOT NULL DEFAULT '',
    "total_nodes" INTEGER NOT NULL DEFAULT 0,
    "alive_nodes" INTEGER NOT NULL DEFAULT 0,
    "cpu_allocated" NUMERIC(12,2) NOT NULL DEFAULT 0,
    "cpu_load" NUMERIC(12,2) NOT NULL DEFAULT 0,
    "memory_allocated" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "memory_used" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "disk_io_read" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "disk_io_write" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "network_in" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "network_out" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "live_streams" INTEGER NOT NULL DEFAULT 0,
    "sessions" INTEGER NOT NULL DEFAULT 0,
    "avg_delay" INTEGER NOT NULL DEFAULT 0,
    "last_heartbeat" TIMESTAMP,
    "node_latency_map" JSONB NOT NULL DEFAULT '{}'::JSONB,
    PRIMARY KEY ("node_id", "report_time")
);

CREATE INDEX IF NOT EXISTS idx_media_report_node_snapshot_time
    ON media_report_node_snapshot ("report_time" DESC);

CREATE INDEX IF NOT EXISTS idx_media_report_node_snapshot_parent_time
    ON media_report_node_snapshot ("parent_node_id", "report_time" DESC);

COMMENT ON TABLE media_report_node_snapshot IS '媒体看板节点总览快照表，仅保存总节点概览需要的历史采样';
COMMENT ON COLUMN media_report_node_snapshot."node_id" IS '上报节点业务标识，不依赖配置表外键';
COMMENT ON COLUMN media_report_node_snapshot."report_time" IS '上报端采样时间，也是快照时间';
COMMENT ON COLUMN media_report_node_snapshot."node_name" IS '节点展示名称，按快照时间点冗余';
COMMENT ON COLUMN media_report_node_snapshot."region" IS '节点所属区域或机房名称';
COMMENT ON COLUMN media_report_node_snapshot."parent_node_id" IS '父节点业务标识，根节点使用0';
COMMENT ON COLUMN media_report_node_snapshot."status" IS '节点运行状态，保存上报原始枚举值';
COMMENT ON COLUMN media_report_node_snapshot."total_nodes" IS '快照时节点视角下的节点总数';
COMMENT ON COLUMN media_report_node_snapshot."alive_nodes" IS '快照时节点视角下的在线节点数';
COMMENT ON COLUMN media_report_node_snapshot."cpu_allocated" IS '快照时已分配CPU容量，单位由上报端统一';
COMMENT ON COLUMN media_report_node_snapshot."cpu_load" IS '快照时CPU负载或利用率，单位由上报端统一';
COMMENT ON COLUMN media_report_node_snapshot."memory_allocated" IS '快照时已分配内存容量，单位MB';
COMMENT ON COLUMN media_report_node_snapshot."memory_used" IS '快照时已使用内存容量，单位MB';
COMMENT ON COLUMN media_report_node_snapshot."disk_io_read" IS '快照时磁盘读取速率，单位KB/S';
COMMENT ON COLUMN media_report_node_snapshot."disk_io_write" IS '快照时磁盘写入速率，单位KB/S';
COMMENT ON COLUMN media_report_node_snapshot."network_in" IS '快照时网络入站速率，单位KB/S';
COMMENT ON COLUMN media_report_node_snapshot."network_out" IS '快照时网络出站速率，单位KB/S';
COMMENT ON COLUMN media_report_node_snapshot."live_streams" IS '快照时直播流数量';
COMMENT ON COLUMN media_report_node_snapshot."sessions" IS '快照时会话数量';
COMMENT ON COLUMN media_report_node_snapshot."avg_delay" IS '快照时节点平均延迟，单位毫秒';
COMMENT ON COLUMN media_report_node_snapshot."last_heartbeat" IS '快照时节点最后心跳时间';
COMMENT ON COLUMN media_report_node_snapshot."node_latency_map" IS '快照时节点延迟矩阵JSON，结构来自接口node_latency_map';

-- Purpose: Stores the latest instance projection when the report protocol provides instance keys.
-- 用途：存储实例列表最新投影，供后续实例上报协议写入。
CREATE TABLE IF NOT EXISTS media_report_instance (
    "instance_id" VARCHAR(128) PRIMARY KEY,
    "instance_name" VARCHAR(128) NOT NULL DEFAULT '',
    "node_id" VARCHAR(64) NOT NULL DEFAULT '',
    "node_name" VARCHAR(128) NOT NULL DEFAULT '',
    "region" VARCHAR(128) NOT NULL DEFAULT '',
    "node_status" VARCHAR(32) NOT NULL DEFAULT '',
    "status" VARCHAR(32) NOT NULL DEFAULT '',
    "cpu_allocated" NUMERIC(12,2) NOT NULL DEFAULT 0,
    "cpu_load" NUMERIC(12,2) NOT NULL DEFAULT 0,
    "memory_allocated" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "memory_used" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "disk_io_read" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "disk_io_write" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "network_in" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "network_out" NUMERIC(14,2) NOT NULL DEFAULT 0,
    "live_streams" INTEGER NOT NULL DEFAULT 0,
    "sessions" INTEGER NOT NULL DEFAULT 0,
    "start_time" TIMESTAMP,
    "version" VARCHAR(64) NOT NULL DEFAULT '',
    "report_time" BIGINT NOT NULL,
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_media_report_instance_node
    ON media_report_instance ("node_id", "status");

CREATE INDEX IF NOT EXISTS idx_media_report_instance_status
    ON media_report_instance ("status");

COMMENT ON TABLE media_report_instance IS '媒体看板实例表，保存实例列表最新上报投影';
COMMENT ON COLUMN media_report_instance."instance_id" IS '实例业务标识，不依赖实例配置外键';
COMMENT ON COLUMN media_report_instance."instance_name" IS '实例展示名称，按上报时间点冗余';
COMMENT ON COLUMN media_report_instance."node_id" IS '实例所属节点业务标识';
COMMENT ON COLUMN media_report_instance."node_name" IS '实例所属节点展示名称，按上报时间点冗余';
COMMENT ON COLUMN media_report_instance."region" IS '实例所属区域或机房名称';
COMMENT ON COLUMN media_report_instance."node_status" IS '实例所属节点状态，用于返回node_info.status';
COMMENT ON COLUMN media_report_instance."status" IS '实例运行状态，保存上报原始枚举值';
COMMENT ON COLUMN media_report_instance."cpu_allocated" IS '实例已分配CPU容量，单位由上报端统一';
COMMENT ON COLUMN media_report_instance."cpu_load" IS '实例CPU负载或利用率，单位由上报端统一';
COMMENT ON COLUMN media_report_instance."memory_allocated" IS '实例已分配内存容量，单位MB';
COMMENT ON COLUMN media_report_instance."memory_used" IS '实例已使用内存容量，单位MB';
COMMENT ON COLUMN media_report_instance."disk_io_read" IS '实例磁盘读取速率，单位KB/S';
COMMENT ON COLUMN media_report_instance."disk_io_write" IS '实例磁盘写入速率，单位KB/S';
COMMENT ON COLUMN media_report_instance."network_in" IS '实例网络入站速率，单位KB/S';
COMMENT ON COLUMN media_report_instance."network_out" IS '实例网络出站速率，单位KB/S';
COMMENT ON COLUMN media_report_instance."live_streams" IS '实例当前承载直播流数量';
COMMENT ON COLUMN media_report_instance."sessions" IS '实例当前会话数量';
COMMENT ON COLUMN media_report_instance."start_time" IS '实例启动时间';
COMMENT ON COLUMN media_report_instance."version" IS '实例版本号';
COMMENT ON COLUMN media_report_instance."report_time" IS '上报端采样时间';
COMMENT ON COLUMN media_report_instance."updated_at" IS '记录更新时间';

-- Purpose: Stores the latest stream projection reported by collection clients.
-- 用途：存储采集客户端上报的流信息最新投影。
CREATE TABLE IF NOT EXISTS media_report_stream (
    "stream_id" VARCHAR(128) PRIMARY KEY,
    "source_type" VARCHAR(32) NOT NULL DEFAULT '',
    "source_id" VARCHAR(128) NOT NULL DEFAULT '',
    "tenant_id" VARCHAR(64) NOT NULL DEFAULT '',
    "node_id" VARCHAR(64) NOT NULL DEFAULT '',
    "node_name" VARCHAR(128) NOT NULL DEFAULT '',
    "instance_id" VARCHAR(128) NOT NULL DEFAULT '',
    "instance_name" VARCHAR(128) NOT NULL DEFAULT '',
    "source_url" VARCHAR(1024) NOT NULL DEFAULT '',
    "stream_name" VARCHAR(255) NOT NULL DEFAULT '',
    "resolution" VARCHAR(32) NOT NULL DEFAULT '',
    "fps" NUMERIC(8,2) NOT NULL DEFAULT 0,
    "bitrate" INTEGER NOT NULL DEFAULT 0,
    "packet_loss" NUMERIC(8,4) NOT NULL DEFAULT 0,
    "status" VARCHAR(32) NOT NULL DEFAULT '',
    "start_time" TIMESTAMP,
    "duration" INTEGER NOT NULL DEFAULT 0,
    "avg_delay" INTEGER NOT NULL DEFAULT 0,
    "protocol_count" INTEGER NOT NULL DEFAULT 0,
    "total_sessions_lifetime" BIGINT NOT NULL DEFAULT 0,
    "current_active_sessions" INTEGER NOT NULL DEFAULT 0,
    "watermark_enabled" BOOLEAN NOT NULL DEFAULT FALSE,
    "protocol_summary" JSONB NOT NULL DEFAULT '[]'::JSONB,
    "report_time" BIGINT NOT NULL,
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_media_report_stream_source
    ON media_report_stream ("source_type", "source_id", "status");

CREATE INDEX IF NOT EXISTS idx_media_report_stream_tenant
    ON media_report_stream ("tenant_id", "status");

CREATE INDEX IF NOT EXISTS idx_media_report_stream_node_instance
    ON media_report_stream ("node_id", "instance_id");

COMMENT ON TABLE media_report_stream IS '媒体看板流表，保存流列表和协议摘要的最新上报投影';
COMMENT ON COLUMN media_report_stream."stream_id" IS '流业务标识，不依赖流配置外键';
COMMENT ON COLUMN media_report_stream."source_type" IS '来源类型，例如节点或实例';
COMMENT ON COLUMN media_report_stream."source_id" IS '来源业务标识，用于节点或实例下钻';
COMMENT ON COLUMN media_report_stream."tenant_id" IS '租户标识，用于数据权限过滤和租户统计';
COMMENT ON COLUMN media_report_stream."node_id" IS '流所属节点业务标识';
COMMENT ON COLUMN media_report_stream."node_name" IS '流所属节点展示名称，按上报时间点冗余';
COMMENT ON COLUMN media_report_stream."instance_id" IS '流所属实例业务标识';
COMMENT ON COLUMN media_report_stream."instance_name" IS '流所属实例展示名称，按上报时间点冗余';
COMMENT ON COLUMN media_report_stream."source_url" IS '流源地址';
COMMENT ON COLUMN media_report_stream."stream_name" IS '流展示名称';
COMMENT ON COLUMN media_report_stream."resolution" IS '当前分辨率';
COMMENT ON COLUMN media_report_stream."fps" IS '当前帧率';
COMMENT ON COLUMN media_report_stream."bitrate" IS '当前码率，单位Kbps';
COMMENT ON COLUMN media_report_stream."packet_loss" IS '当前丢包率';
COMMENT ON COLUMN media_report_stream."status" IS '流运行状态，保存上报原始枚举值';
COMMENT ON COLUMN media_report_stream."start_time" IS '流开始时间';
COMMENT ON COLUMN media_report_stream."duration" IS '流持续时间，单位秒';
COMMENT ON COLUMN media_report_stream."avg_delay" IS '流平均延迟，单位毫秒';
COMMENT ON COLUMN media_report_stream."protocol_count" IS '支持的协议数量';
COMMENT ON COLUMN media_report_stream."total_sessions_lifetime" IS '流历史累计会话数量';
COMMENT ON COLUMN media_report_stream."current_active_sessions" IS '流当前活跃会话数量';
COMMENT ON COLUMN media_report_stream."watermark_enabled" IS '当前是否启用水印';
COMMENT ON COLUMN media_report_stream."protocol_summary" IS '协议摘要JSON数组，结构来自接口protocol_summary';
COMMENT ON COLUMN media_report_stream."report_time" IS '上报端采样时间';
COMMENT ON COLUMN media_report_stream."updated_at" IS '记录更新时间';

-- Purpose: Stores the latest session projection when the report protocol provides session keys.
-- 用途：存储会话列表最新投影，供后续会话上报协议写入。
CREATE TABLE IF NOT EXISTS media_report_session (
    "session_id" VARCHAR(128) PRIMARY KEY,
    "stream_id" VARCHAR(128) NOT NULL DEFAULT '',
    "stream_name" VARCHAR(255) NOT NULL DEFAULT '',
    "tenant_id" VARCHAR(64) NOT NULL DEFAULT '',
    "client_id" VARCHAR(128) NOT NULL DEFAULT '',
    "client_ip" INET,
    "client_type" VARCHAR(32) NOT NULL DEFAULT '',
    "user_name" VARCHAR(128) NOT NULL DEFAULT '',
    "protocol_type" VARCHAR(32) NOT NULL DEFAULT '',
    "start_time" TIMESTAMP,
    "play_duration" INTEGER NOT NULL DEFAULT 0,
    "current_fps" NUMERIC(8,2) NOT NULL DEFAULT 0,
    "current_bitrate" INTEGER NOT NULL DEFAULT 0,
    "current_resolution" VARCHAR(32) NOT NULL DEFAULT '',
    "node_id" VARCHAR(64) NOT NULL DEFAULT '',
    "node_name" VARCHAR(128) NOT NULL DEFAULT '',
    "instance_id" VARCHAR(128) NOT NULL DEFAULT '',
    "instance_name" VARCHAR(128) NOT NULL DEFAULT '',
    "link_hops" JSONB NOT NULL DEFAULT '[]'::JSONB,
    "total_link_latency" INTEGER NOT NULL DEFAULT 0,
    "report_time" BIGINT NOT NULL,
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_media_report_session_stream
    ON media_report_session ("stream_id", "protocol_type");

CREATE INDEX IF NOT EXISTS idx_media_report_session_tenant
    ON media_report_session ("tenant_id", "protocol_type");

CREATE INDEX IF NOT EXISTS idx_media_report_session_node_instance
    ON media_report_session ("node_id", "instance_id");

COMMENT ON TABLE media_report_session IS '媒体看板会话表，保存会话列表和链路跳点的最新上报投影';
COMMENT ON COLUMN media_report_session."session_id" IS '会话业务标识';
COMMENT ON COLUMN media_report_session."stream_id" IS '会话所属流业务标识';
COMMENT ON COLUMN media_report_session."stream_name" IS '会话所属流展示名称，按上报时间点冗余';
COMMENT ON COLUMN media_report_session."tenant_id" IS '租户标识，用于数据权限过滤和租户统计';
COMMENT ON COLUMN media_report_session."client_id" IS '客户端标识';
COMMENT ON COLUMN media_report_session."client_ip" IS '客户端IP地址';
COMMENT ON COLUMN media_report_session."client_type" IS '客户端类型';
COMMENT ON COLUMN media_report_session."user_name" IS '播放用户展示名称';
COMMENT ON COLUMN media_report_session."protocol_type" IS '播放协议类型';
COMMENT ON COLUMN media_report_session."start_time" IS '会话开始时间';
COMMENT ON COLUMN media_report_session."play_duration" IS '播放持续时间，单位秒';
COMMENT ON COLUMN media_report_session."current_fps" IS '当前播放帧率';
COMMENT ON COLUMN media_report_session."current_bitrate" IS '当前播放码率，单位Kbps';
COMMENT ON COLUMN media_report_session."current_resolution" IS '当前播放分辨率';
COMMENT ON COLUMN media_report_session."node_id" IS '会话承载节点业务标识';
COMMENT ON COLUMN media_report_session."node_name" IS '会话承载节点展示名称，按上报时间点冗余';
COMMENT ON COLUMN media_report_session."instance_id" IS '会话承载实例业务标识';
COMMENT ON COLUMN media_report_session."instance_name" IS '会话承载实例展示名称，按上报时间点冗余';
COMMENT ON COLUMN media_report_session."link_hops" IS '会话链路跳点JSON数组，结构来自接口link_hops';
COMMENT ON COLUMN media_report_session."total_link_latency" IS '会话链路总延迟，单位毫秒';
COMMENT ON COLUMN media_report_session."report_time" IS '上报端采样时间';
COMMENT ON COLUMN media_report_session."updated_at" IS '记录更新时间';
