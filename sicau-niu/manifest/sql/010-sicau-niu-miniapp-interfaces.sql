-- ------------------------------------------------------------
-- 010 sicau-niu mini-program interface SQL file
-- 010 sicau-niu 小程序完整接口 SQL 文件
-- Purpose: Mini-program runtime config, private activation photos, request idempotency and persistent iron transport.
-- Dialect: PostgreSQL. Idempotent: safe to re-run.
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS plugin_sicau_niu_activation_photo (
    "id"            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "token"         VARCHAR(64) NOT NULL DEFAULT '',
    "user_id"       BIGINT NOT NULL DEFAULT 0,
    "request_id"    VARCHAR(64) NOT NULL DEFAULT '',
    "activity_date" VARCHAR(10) NOT NULL DEFAULT '',
    "daily_slot"    INT NOT NULL DEFAULT 0,
    "object_path"   VARCHAR(512) NOT NULL DEFAULT '',
    "original_name" VARCHAR(255) NOT NULL DEFAULT '',
    "content_type"  VARCHAR(64) NOT NULL DEFAULT '',
    "size_bytes"    BIGINT NOT NULL DEFAULT 0,
    "used_at"       TIMESTAMPTZ,
    "created_at"    TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE plugin_sicau_niu_activation_photo IS 'Private mini-program activation photo metadata';
COMMENT ON COLUMN plugin_sicau_niu_activation_photo."token" IS 'Opaque unguessable player-facing photo identifier';
COMMENT ON COLUMN plugin_sicau_niu_activation_photo."user_id" IS 'Owning player ID';
COMMENT ON COLUMN plugin_sicau_niu_activation_photo."request_id" IS 'Player-scoped idempotency key';
COMMENT ON COLUMN plugin_sicau_niu_activation_photo."activity_date" IS 'Beijing natural-day key';
COMMENT ON COLUMN plugin_sicau_niu_activation_photo."daily_slot" IS 'Bounded successful upload slot 1..10 within the activity date';
COMMENT ON COLUMN plugin_sicau_niu_activation_photo."object_path" IS 'Plugin-private logical object storage path';
COMMENT ON COLUMN plugin_sicau_niu_activation_photo."used_at" IS 'Successful activation consumption time; NULL means unused';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_activation_photo_token
    ON plugin_sicau_niu_activation_photo ("token");
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_activation_photo_object_path
    ON plugin_sicau_niu_activation_photo ("object_path");
CREATE INDEX IF NOT EXISTS idx_sicau_niu_activation_photo_user_created
    ON plugin_sicau_niu_activation_photo ("user_id", "created_at" DESC);
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_activation_photo_user_request
    ON plugin_sicau_niu_activation_photo ("user_id", "request_id") WHERE "request_id" <> '';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_activation_photo_user_day_slot
    ON plugin_sicau_niu_activation_photo ("user_id", "activity_date", "daily_slot") WHERE "daily_slot" > 0;

CREATE TABLE IF NOT EXISTS plugin_sicau_niu_miniapp_config (
    "id"                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "config_key"            VARCHAR(32) NOT NULL DEFAULT 'default',
    "assets_version"        VARCHAR(32) NOT NULL DEFAULT '1',
    "static_asset_base_url" VARCHAR(500) NOT NULL DEFAULT '',
    "activity_phase"        VARCHAR(16) NOT NULL DEFAULT 'active',
    "default_campus"        VARCHAR(16) NOT NULL DEFAULT 'cd',
    "anniversary"           VARCHAR(100) NOT NULL DEFAULT '',
    "anniversary_at"        VARCHAR(10) NOT NULL DEFAULT '',
    "debug"                 INT NOT NULL DEFAULT 0,
    "campuses_json"         TEXT NOT NULL DEFAULT '[]',
    "created_at"            TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at"            TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE plugin_sicau_niu_miniapp_config IS 'Operator-maintained public mini-program runtime configuration';
COMMENT ON COLUMN plugin_sicau_niu_miniapp_config."campuses_json" IS 'Validated campus map asset and affine calibration JSON';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_miniapp_config_key
    ON plugin_sicau_niu_miniapp_config ("config_key");

CREATE TABLE IF NOT EXISTS plugin_sicau_niu_transport_team (
    "id"             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "code"           VARCHAR(12) NOT NULL DEFAULT '',
    "name"           VARCHAR(64) NOT NULL DEFAULT '',
    "campus_id"      VARCHAR(16) NOT NULL DEFAULT '',
    "leader_user_id" BIGINT NOT NULL DEFAULT 0,
    "create_request_id" VARCHAR(64) NOT NULL DEFAULT '',
    "iron_id"        BIGINT NOT NULL DEFAULT 0,
    "status"         VARCHAR(16) NOT NULL DEFAULT 'forming',
    "min_members"    INT NOT NULL DEFAULT 3,
    "max_members"    INT NOT NULL DEFAULT 6,
    "visible"        INT NOT NULL DEFAULT 1,
    "created_at"     TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at"     TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"     TIMESTAMPTZ
);

COMMENT ON TABLE plugin_sicau_niu_transport_team IS 'Iron-cow transport team';
COMMENT ON COLUMN plugin_sicau_niu_transport_team."status" IS 'Team status: forming, active, ended';
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_attribute
        WHERE attrelid = 'plugin_sicau_niu_transport_team'::regclass
          AND attname = 'code'
          AND NOT attisdropped
    ) THEN
        CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_transport_team_code
            ON plugin_sicau_niu_transport_team ("code") WHERE "deleted_at" IS NULL;
    END IF;
END $$;
CREATE INDEX IF NOT EXISTS idx_sicau_niu_transport_team_state
    ON plugin_sicau_niu_transport_team ("status", "visible", "created_at" DESC)
    WHERE "deleted_at" IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_transport_team_create_request
    ON plugin_sicau_niu_transport_team ("leader_user_id", "create_request_id")
    WHERE "create_request_id" <> '';

CREATE TABLE IF NOT EXISTS plugin_sicau_niu_transport_member (
    "id"                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "team_id"           BIGINT NOT NULL DEFAULT 0,
    "user_id"           BIGINT NOT NULL DEFAULT 0,
    "join_request_id"   VARCHAR(64) NOT NULL DEFAULT '',
    "leave_request_id"  VARCHAR(64) NOT NULL DEFAULT '',
    "role"              VARCHAR(16) NOT NULL DEFAULT 'member',
    "joined_at"         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "left_at"           TIMESTAMPTZ,
    "last_heartbeat_at" TIMESTAMPTZ,
    "created_at"        TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at"        TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE plugin_sicau_niu_transport_member IS 'Iron-cow transport team membership history';
COMMENT ON COLUMN plugin_sicau_niu_transport_member."role" IS 'Member role: leader, member';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_transport_member_user_active
    ON plugin_sicau_niu_transport_member ("user_id") WHERE "left_at" IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_transport_member_team_user_active
    ON plugin_sicau_niu_transport_member ("team_id", "user_id") WHERE "left_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_sicau_niu_transport_member_team_active
    ON plugin_sicau_niu_transport_member ("team_id", "joined_at") WHERE "left_at" IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_transport_member_join_request
    ON plugin_sicau_niu_transport_member ("user_id", "join_request_id") WHERE "join_request_id" <> '';
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_attribute
        WHERE attrelid = 'plugin_sicau_niu_transport_member'::regclass
          AND attname = 'leave_request_id'
          AND NOT attisdropped
    ) THEN
        CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_transport_member_leave_request
            ON plugin_sicau_niu_transport_member ("user_id", "leave_request_id") WHERE "leave_request_id" <> '';
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS plugin_sicau_niu_transport_session (
    "id"             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "team_id"        BIGINT NOT NULL DEFAULT 0,
    "iron_id"        BIGINT NOT NULL DEFAULT 0,
    "started_by_user_id" BIGINT NOT NULL DEFAULT 0,
    "start_request_id"   VARCHAR(64) NOT NULL DEFAULT '',
    "ended_by_user_id"   BIGINT NOT NULL DEFAULT 0,
    "end_request_id"     VARCHAR(64) NOT NULL DEFAULT '',
    "status"         VARCHAR(20) NOT NULL DEFAULT 'active',
    "started_at"     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "last_active_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "ended_at"       TIMESTAMPTZ,
    "moved_meters"   DOUBLE PRECISION NOT NULL DEFAULT 0,
    "last_lat"       DOUBLE PRECISION NOT NULL DEFAULT 0,
    "last_lng"       DOUBLE PRECISION NOT NULL DEFAULT 0,
    "created_at"     TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at"     TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE plugin_sicau_niu_transport_session IS 'Iron-cow transport session';
COMMENT ON COLUMN plugin_sicau_niu_transport_session."status" IS 'Session status: active, idle_timeout, ended';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_transport_session_active
    ON plugin_sicau_niu_transport_session ((1)) WHERE "status" = 'active';
CREATE INDEX IF NOT EXISTS idx_sicau_niu_transport_session_team
    ON plugin_sicau_niu_transport_session ("team_id", "started_at" DESC);
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_transport_session_start_request
    ON plugin_sicau_niu_transport_session ("started_by_user_id", "start_request_id") WHERE "start_request_id" <> '';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_transport_session_end_request
    ON plugin_sicau_niu_transport_session ("ended_by_user_id", "end_request_id") WHERE "end_request_id" <> '';

CREATE TABLE IF NOT EXISTS plugin_sicau_niu_transport_track (
    "id"              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "session_id"      BIGINT NOT NULL DEFAULT 0,
    "user_id"         BIGINT NOT NULL DEFAULT 0,
    "request_id"      VARCHAR(64) NOT NULL DEFAULT '',
    "lat"             DOUBLE PRECISION NOT NULL DEFAULT 0,
    "lng"             DOUBLE PRECISION NOT NULL DEFAULT 0,
    "distance_meters" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "recorded_at"     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_at"      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE plugin_sicau_niu_transport_track IS 'Bounded iron-cow transport heartbeat trajectory';
CREATE INDEX IF NOT EXISTS idx_sicau_niu_transport_track_session
    ON plugin_sicau_niu_transport_track ("session_id", "id" DESC);
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_transport_track_user_request
    ON plugin_sicau_niu_transport_track ("user_id", "request_id") WHERE "request_id" <> '';

ALTER TABLE plugin_sicau_niu_activation
    ADD COLUMN IF NOT EXISTS "request_id" VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "response_json" TEXT NOT NULL DEFAULT '';
COMMENT ON COLUMN plugin_sicau_niu_activation."request_id" IS 'Player-scoped idempotency key';
COMMENT ON COLUMN plugin_sicau_niu_activation."response_json" IS 'Stable successful activation response for idempotent replay';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_activation_user_request
    ON plugin_sicau_niu_activation ("user_id", "request_id")
    WHERE "deleted_at" IS NULL AND "request_id" <> '';

ALTER TABLE plugin_sicau_niu_checkin
    ADD COLUMN IF NOT EXISTS "request_id" VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "result_balance" BIGINT NOT NULL DEFAULT 0;
COMMENT ON COLUMN plugin_sicau_niu_checkin."request_id" IS 'Player-scoped idempotency key';
COMMENT ON COLUMN plugin_sicau_niu_checkin."result_balance" IS 'Stable account balance returned by the first successful request';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_checkin_user_request
    ON plugin_sicau_niu_checkin ("user_id", "request_id")
    WHERE "deleted_at" IS NULL AND "request_id" <> '';
