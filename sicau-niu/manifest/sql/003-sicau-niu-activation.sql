-- ------------------------------------------------------------
-- 003 sicau-niu activation SQL file
-- 003 sicau-niu 激活记录 SQL 文件
-- Purpose: Player activation records (LBS activation, shared-pool first-activator, daily limit) for C3 niu-activation.
-- 用途:玩家激活记录(LBS 激活、共享池首发、每日限次),C3 niu-activation。
-- Dialect: PostgreSQL. Idempotent: safe to re-run.
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS plugin_sicau_niu_activation (
    "id"            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id"       BIGINT NOT NULL DEFAULT 0,
    "niu_id"        BIGINT NOT NULL DEFAULT 0,
    "activity_date" VARCHAR(10) NOT NULL DEFAULT '',
    "activated_at"  TIMESTAMP,
    "is_first"      INT NOT NULL DEFAULT 0,
    "order_no"      INT NOT NULL DEFAULT 0,
    "photo_path"    VARCHAR(500) NOT NULL DEFAULT '',
    "created_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"    TIMESTAMP
);

COMMENT ON TABLE plugin_sicau_niu_activation IS 'sicau-niu player activation record table';
COMMENT ON COLUMN plugin_sicau_niu_activation."user_id" IS 'Activating player ID';
COMMENT ON COLUMN plugin_sicau_niu_activation."niu_id" IS 'Activated cattle ID';
COMMENT ON COLUMN plugin_sicau_niu_activation."activity_date" IS 'Activation date YYYY-MM-DD for the per-day limit';
COMMENT ON COLUMN plugin_sicau_niu_activation."activated_at" IS 'Activation time';
COMMENT ON COLUMN plugin_sicau_niu_activation."is_first" IS 'Whether this is the cattle first-activation: 1 first, 0 later';
COMMENT ON COLUMN plugin_sicau_niu_activation."order_no" IS 'Arrival order for this cattle, starting at 1';
COMMENT ON COLUMN plugin_sicau_niu_activation."photo_path" IS 'Optional activation photo storage path (evidence only, no recognition)';
COMMENT ON COLUMN plugin_sicau_niu_activation."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_sicau_niu_activation."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_sicau_niu_activation."deleted_at" IS 'Soft-delete time, NULL means active';

-- 每名玩家每个自然日最多一条激活记录。
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_activation_user_date
    ON plugin_sicau_niu_activation ("user_id", "activity_date") WHERE "deleted_at" IS NULL;
-- 同一头牛对同一玩家不可重复激活。
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_activation_user_niu
    ON plugin_sicau_niu_activation ("user_id", "niu_id") WHERE "deleted_at" IS NULL;
-- 按牛统计已激活数 / 首发 / 序号。
CREATE INDEX IF NOT EXISTS idx_sicau_niu_activation_niu
    ON plugin_sicau_niu_activation ("niu_id");
