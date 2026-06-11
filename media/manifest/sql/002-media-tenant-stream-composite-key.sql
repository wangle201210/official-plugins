-- Purpose: Changes tenant stream config identity from tenant-only to tenant and node.
-- 用途：将租户流配置唯一标识从单租户调整为租户和节点组合。

ALTER TABLE media_tenant_stream_config DROP CONSTRAINT IF EXISTS media_tenant_stream_config_pkey;
ALTER TABLE media_tenant_stream_config DROP CONSTRAINT IF EXISTS uk_media_tenant_stream_config_tenant;
ALTER TABLE media_tenant_stream_config DROP CONSTRAINT IF EXISTS uk_media_tenant_stream_config_tenant_node;

ALTER TABLE media_tenant_stream_config
    ADD CONSTRAINT media_tenant_stream_config_pkey PRIMARY KEY ("tenant_id", "node_num");

CREATE INDEX IF NOT EXISTS idx_media_tenant_stream_config_tenant_id
    ON media_tenant_stream_config ("tenant_id");
