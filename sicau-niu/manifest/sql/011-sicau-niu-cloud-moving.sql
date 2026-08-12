-- ------------------------------------------------------------
-- 011 sicau-niu cloud-moving migration
-- 011 sicau-niu 组队云搬牛迁移
-- Purpose: Replace shared iron-transport sessions with persistent teams and immutable player contribution reports.
-- Dialect: PostgreSQL. Idempotent: safe to re-run.
-- ------------------------------------------------------------

ALTER TABLE plugin_sicau_niu_transport_team
    ADD COLUMN IF NOT EXISTS "member_count" INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS "total_contribution_meters" BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS "last_active_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN IF NOT EXISTS "invalidated_at" TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS "invalid_reason" VARCHAR(64) NOT NULL DEFAULT '';

UPDATE plugin_sicau_niu_transport_team
SET "status" = 'invalid',
    "visible" = 0,
    "invalidated_at" = COALESCE("invalidated_at", CURRENT_TIMESTAMP),
    "invalid_reason" = CASE WHEN "invalid_reason" = '' THEN 'legacy_model' ELSE "invalid_reason" END
WHERE "status" NOT IN ('effective', 'invalid');

ALTER TABLE plugin_sicau_niu_transport_team
    DROP COLUMN IF EXISTS "code",
    DROP COLUMN IF EXISTS "campus_id",
    DROP COLUMN IF EXISTS "iron_id",
    DROP COLUMN IF EXISTS "min_members",
    DROP COLUMN IF EXISTS "max_members";
ALTER TABLE plugin_sicau_niu_transport_team
    ALTER COLUMN "status" SET DEFAULT 'effective';

COMMENT ON TABLE plugin_sicau_niu_transport_team IS 'Cloud-moving team history';
COMMENT ON COLUMN plugin_sicau_niu_transport_team."leader_user_id" IS 'Team creator player ID; creator has no lifecycle authority';
COMMENT ON COLUMN plugin_sicau_niu_transport_team."status" IS 'Team status: effective, invalid';
COMMENT ON COLUMN plugin_sicau_niu_transport_team."last_active_at" IS 'Latest successful create, join or contribution server commit time';
COMMENT ON COLUMN plugin_sicau_niu_transport_team."name" IS 'Trimmed display name; unique while the team is effective';
DROP INDEX IF EXISTS uk_sicau_niu_transport_team_code;
DROP INDEX IF EXISTS idx_sicau_niu_transport_team_state;
CREATE INDEX idx_sicau_niu_transport_team_state
    ON plugin_sicau_niu_transport_team ("status", "last_active_at", "created_at" DESC)
    WHERE "deleted_at" IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_transport_team_effective_name
    ON plugin_sicau_niu_transport_team ("name")
    WHERE "status" = 'effective' AND "deleted_at" IS NULL;

INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, '云搬牛团状态', 'sicau_niu_transport_team_status', 1, 1, 'Cloud-moving team status options', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'sicau_niu_transport_team_status', '有效', 'effective', 1, 'success', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'sicau_niu_transport_team_status', '已失效', 'invalid', 2, 'default', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;

ALTER TABLE plugin_sicau_niu_transport_member
    ADD COLUMN IF NOT EXISTS "total_contribution_meters" BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS "last_report_lat" DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS "last_report_lng" DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS "last_report_at" TIMESTAMPTZ;

UPDATE plugin_sicau_niu_transport_member AS member
SET "left_at" = COALESCE(member."left_at", team."invalidated_at", CURRENT_TIMESTAMP)
FROM plugin_sicau_niu_transport_team AS team
WHERE member."team_id" = team."id"
  AND member."left_at" IS NULL
  AND team."status" = 'invalid';

ALTER TABLE plugin_sicau_niu_transport_member
    DROP COLUMN IF EXISTS "leave_request_id",
    DROP COLUMN IF EXISTS "last_heartbeat_at";

COMMENT ON TABLE plugin_sicau_niu_transport_member IS 'Cloud-moving team membership history';
COMMENT ON COLUMN plugin_sicau_niu_transport_member."role" IS 'Member display role: creator, member';
DROP INDEX IF EXISTS uk_sicau_niu_transport_member_leave_request;

INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, '云搬牛成员角色', 'sicau_niu_transport_member_role', 1, 1, 'Cloud-moving member role options', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'sicau_niu_transport_member_role', '创建人', 'creator', 1, 'primary', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'sicau_niu_transport_member_role', '成员', 'member', 2, 'default', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, '云搬牛失效原因', 'sicau_niu_transport_invalid_reason', 1, 1, 'Cloud-moving invalidation reason options', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'sicau_niu_transport_invalid_reason', '连续三天不活跃', 'inactive_72h', 1, 'warning', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'sicau_niu_transport_invalid_reason', '旧搬运模型迁移', 'legacy_model', 2, 'default', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS plugin_sicau_niu_transport_report (
    "id"              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "activity_key"    VARCHAR(64) NOT NULL DEFAULT 'default',
    "team_id"         BIGINT NOT NULL DEFAULT 0,
    "member_id"       BIGINT NOT NULL DEFAULT 0,
    "user_id"         BIGINT NOT NULL DEFAULT 0,
    "request_id"      VARCHAR(64) NOT NULL DEFAULT '',
    "activity_date"   VARCHAR(10) NOT NULL DEFAULT '',
    "start_lat"       DOUBLE PRECISION,
    "start_lng"       DOUBLE PRECISION,
    "end_lat"         DOUBLE PRECISION NOT NULL DEFAULT 0,
    "end_lng"         DOUBLE PRECISION NOT NULL DEFAULT 0,
    "sampled_at"      TIMESTAMPTZ NOT NULL,
    "accepted_at"     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "contribution_meters" BIGINT NOT NULL DEFAULT 0,
    "user_total_meters" BIGINT NOT NULL DEFAULT 0,
    "team_total_meters" BIGINT NOT NULL DEFAULT 0,
    "daily_report_count" INT NOT NULL DEFAULT 0,
    "created_at"      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE plugin_sicau_niu_transport_report IS 'Immutable successful cloud-moving position report facts';
COMMENT ON COLUMN plugin_sicau_niu_transport_report."activity_date" IS 'Beijing natural-day key derived from accepted_at';
COMMENT ON COLUMN plugin_sicau_niu_transport_report."start_lat" IS 'Previous successful report latitude; NULL for the first report in a membership';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_transport_report_user_request
    ON plugin_sicau_niu_transport_report ("user_id", "request_id") WHERE "request_id" <> '';
CREATE INDEX IF NOT EXISTS idx_sicau_niu_transport_report_user_day
    ON plugin_sicau_niu_transport_report ("user_id", "activity_date", "accepted_at" DESC);
CREATE INDEX IF NOT EXISTS idx_sicau_niu_transport_report_team_time
    ON plugin_sicau_niu_transport_report ("team_id", "accepted_at" DESC);
CREATE INDEX IF NOT EXISTS idx_sicau_niu_transport_report_activity_day
    ON plugin_sicau_niu_transport_report ("activity_key", "activity_date", "accepted_at" DESC);

DROP TABLE IF EXISTS plugin_sicau_niu_transport_track;
DROP TABLE IF EXISTS plugin_sicau_niu_transport_session;
