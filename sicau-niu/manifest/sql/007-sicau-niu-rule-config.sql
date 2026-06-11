-- ------------------------------------------------------------
-- 007 sicau-niu rule config SQL file
-- 007 sicau-niu 运行规则配置 SQL 文件
-- Purpose: Operator-maintained runtime rule parameters for C4 interaction,
--          C5 ranking, C10 anomaly alerts, C3 poster badge and C6 miniapp link.
-- Dialect: PostgreSQL. Idempotent.
-- ------------------------------------------------------------

DROP INDEX IF EXISTS idx_sicau_niu_niu_stage_online;
DROP INDEX IF EXISTS idx_sicau_niu_niu_online;

ALTER TABLE IF EXISTS plugin_sicau_niu_niu
    DROP COLUMN IF EXISTS "release_stage";

CREATE TABLE IF NOT EXISTS plugin_sicau_niu_rule_config (
    "id"           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "config_key"   VARCHAR(64) NOT NULL DEFAULT '',
    "config_value" VARCHAR(255) NOT NULL DEFAULT '',
    "remark"       VARCHAR(255) NOT NULL DEFAULT '',
    "created_at"   TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at"   TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"   TIMESTAMPTZ
);

COMMENT ON TABLE plugin_sicau_niu_rule_config IS 'sicau-niu operator-maintained runtime rule configuration';
COMMENT ON COLUMN plugin_sicau_niu_rule_config."config_key" IS 'Stable runtime rule key';
COMMENT ON COLUMN plugin_sicau_niu_rule_config."config_value" IS 'Runtime rule value stored as text and validated by service';
COMMENT ON COLUMN plugin_sicau_niu_rule_config."remark" IS 'Operator-facing description of the rule';
COMMENT ON COLUMN plugin_sicau_niu_rule_config."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_sicau_niu_rule_config."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_sicau_niu_rule_config."deleted_at" IS 'Soft-delete time, NULL means active';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_rule_config_key
    ON plugin_sicau_niu_rule_config ("config_key") WHERE "deleted_at" IS NULL;

DO $$
DECLARE
    time_column RECORD;
BEGIN
    FOR time_column IN
        SELECT column_meta.table_name, column_meta.column_name
        FROM information_schema.columns AS column_meta
        JOIN information_schema.tables AS table_meta
          ON table_meta.table_schema = column_meta.table_schema
         AND table_meta.table_name = column_meta.table_name
        WHERE column_meta.table_schema = current_schema()
          AND column_meta.table_name LIKE 'plugin_sicau_niu_%'
          AND column_meta.data_type = 'timestamp without time zone'
          AND table_meta.table_type = 'BASE TABLE'
        ORDER BY column_meta.table_name, column_meta.ordinal_position
    LOOP
        EXECUTE format(
            'ALTER TABLE %I.%I ALTER COLUMN %I TYPE TIMESTAMPTZ',
            current_schema(),
            time_column.table_name,
            time_column.column_name
        );
    END LOOP;
END $$;

DO $$
BEGIN
    IF to_regclass('plugin_sicau_niu_niu') IS NOT NULL THEN
        CREATE INDEX IF NOT EXISTS idx_sicau_niu_niu_online
            ON plugin_sicau_niu_niu ("online_at") WHERE "deleted_at" IS NULL;
    END IF;
END $$;

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'activation.lbsThresholdMeters', '50', 'LBS 激活判距阈值（米）'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'activation.lbsThresholdMeters' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'poster.campusBadge', '', '激活海报和电子证书上的校庆标识'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'poster.campusBadge' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'checkin.minAmount', '20', '每日签到草量下限'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'checkin.minAmount' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'checkin.maxAmount', '50', '每日签到草量上限'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'checkin.maxAmount' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'steal.dailyTargets', '12', '每日随机可偷名单人数'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'steal.dailyTargets' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'steal.dailyLimit', '5', '每日偷草次数上限'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'steal.dailyLimit' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'steal.minAmount', '5', '单次偷草草量下限'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'steal.minAmount' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'steal.maxAmount', '20', '单次偷草草量上限'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'steal.maxAmount' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'gift.dailyLimit', '12', '每日送草次数上限'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'gift.dailyLimit' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'gift.minAmount', '12', '单次送草草量下限'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'gift.minAmount' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'ironBonus.thresholdMeters', '12', '铁牛邻近喂草加成阈值（米）'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'ironBonus.thresholdMeters' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'ranking.topN', '100', '排行榜返回 Top-N 上限'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'ranking.topN' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'anomaly.feedDailyThreshold', '100', '单日喂草次数异常阈值'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'anomaly.feedDailyThreshold' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'anomaly.stealDailyThreshold', '5', '单日偷草次数异常阈值'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'anomaly.stealDailyThreshold' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'anomaly.listLimit', '200', '异常告警返回上限'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'anomaly.listLimit' AND "deleted_at" IS NULL
);

INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
SELECT 'miniapp.url', '', 'H5 数字纪念墙回跳小程序链接'
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_sicau_niu_rule_config WHERE "config_key" = 'miniapp.url' AND "deleted_at" IS NULL
);
