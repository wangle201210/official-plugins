-- ------------------------------------------------------------
-- 005 sicau-niu honor SQL file
-- 005 sicau-niu 荣誉 SQL 文件
-- Purpose: Honor definitions and player honor grants for C5 niu-ranking-honor. Rankings are aggregated from feeding (no table).
-- Dialect: PostgreSQL. Idempotent.
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS plugin_sicau_niu_honor_def (
    "id"          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "honor_type"  VARCHAR(20) NOT NULL DEFAULT '',
    "code"        VARCHAR(64) NOT NULL DEFAULT '',
    "name"        VARCHAR(64) NOT NULL DEFAULT '',
    "unlock_type" VARCHAR(20) NOT NULL DEFAULT '',
    "threshold"   INT NOT NULL DEFAULT 0,
    "category"    VARCHAR(16) NOT NULL DEFAULT '',
    "image_path"  VARCHAR(500) NOT NULL DEFAULT '',
    "sort"        INT NOT NULL DEFAULT 0,
    "created_at"  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"  TIMESTAMP
);
COMMENT ON TABLE plugin_sicau_niu_honor_def IS 'sicau-niu honor definition table';
COMMENT ON COLUMN plugin_sicau_niu_honor_def."honor_type" IS 'Honor type: badge, avatar_frame, certificate';
COMMENT ON COLUMN plugin_sicau_niu_honor_def."code" IS 'Honor unique code among active rows';
COMMENT ON COLUMN plugin_sicau_niu_honor_def."unlock_type" IS 'Unlock rule: participation, feed_count, activation_count, category_complete, full_complete';
COMMENT ON COLUMN plugin_sicau_niu_honor_def."threshold" IS 'Threshold for count-based unlock rules';
COMMENT ON COLUMN plugin_sicau_niu_honor_def."category" IS 'Card category for category_complete unlock rule';
COMMENT ON COLUMN plugin_sicau_niu_honor_def."image_path" IS 'Honor image/template storage path';
COMMENT ON COLUMN plugin_sicau_niu_honor_def."deleted_at" IS 'Soft-delete time, NULL means active';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_honor_def_code
    ON plugin_sicau_niu_honor_def ("code") WHERE "deleted_at" IS NULL;

CREATE TABLE IF NOT EXISTS plugin_sicau_niu_user_honor (
    "id"          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id"     BIGINT NOT NULL DEFAULT 0,
    "honor_id"    BIGINT NOT NULL DEFAULT 0,
    "unlocked_at" TIMESTAMP,
    "created_at"  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"  TIMESTAMP
);
COMMENT ON TABLE plugin_sicau_niu_user_honor IS 'sicau-niu player honor grant table (persistent grant by C7 settlement)';
COMMENT ON COLUMN plugin_sicau_niu_user_honor."user_id" IS 'Player ID';
COMMENT ON COLUMN plugin_sicau_niu_user_honor."honor_id" IS 'Granted honor definition ID';
COMMENT ON COLUMN plugin_sicau_niu_user_honor."unlocked_at" IS 'Grant time';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_user_honor
    ON plugin_sicau_niu_user_honor ("user_id", "honor_id") WHERE "deleted_at" IS NULL;
