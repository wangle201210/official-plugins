-- 001: linapro-uidentity-cas backend schema
-- Purpose: Creates tenant-scoped identity, CAS, OAuth, password policy, blacklist, and audit tables owned by the plugin.

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_account (
    "id"                  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"           INT          NOT NULL DEFAULT 0,
    "number"              VARCHAR(128) NOT NULL DEFAULT '',
    "name"                VARCHAR(128) NOT NULL DEFAULT '',
    "phone"               VARCHAR(64)  NOT NULL DEFAULT '',
    "password_hash"       VARCHAR(255) NOT NULL DEFAULT '',
    "effect_at"           TIMESTAMP    NULL DEFAULT NULL,
    "expire_at"           TIMESTAMP    NULL DEFAULT NULL,
    "password_updated_at" TIMESTAMP    NULL DEFAULT NULL,
    "pass_level"          SMALLINT     NOT NULL DEFAULT 0,
    "container_id"        BIGINT       NOT NULL DEFAULT 0,
    "unit_id"             BIGINT       NOT NULL DEFAULT 0,
    "status"              SMALLINT     NOT NULL DEFAULT 0,
    "created_by"          BIGINT       NOT NULL DEFAULT 0,
    "updated_by"          BIGINT       NOT NULL DEFAULT 0,
    "created_at"          TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"          TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"          TIMESTAMP    NULL DEFAULT NULL
);

COMMENT ON TABLE plugin_linapro_uidentity_cas_account IS 'UIdentity account table';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_account."tenant_id" IS 'Owning tenant ID, 0 means platform';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_account."number" IS 'Stable account number';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_account."name" IS 'Account display name';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_account."phone" IS 'Mobile phone number';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_account."password_hash" IS 'Password hash managed by the plugin';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_account."pass_level" IS 'Password strength level: 0=invalid, higher is stronger';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_account."container_id" IS 'Container ID';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_account."unit_id" IS 'Primary unit ID';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_account."status" IS 'Account status: 0=not active, 1=normal, 2=locked';

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_account_detail (
    "account_id"    BIGINT       PRIMARY KEY,
    "tenant_id"     INT          NOT NULL DEFAULT 0,
    "birthday"      VARCHAR(10)  NOT NULL DEFAULT '',
    "email"         VARCHAR(128) NOT NULL DEFAULT '',
    "gender"        SMALLINT     NOT NULL DEFAULT 0,
    "qq"            VARCHAR(64)  NOT NULL DEFAULT '',
    "wechat"        VARCHAR(128) NOT NULL DEFAULT '',
    "idcard"        VARCHAR(64)  NOT NULL DEFAULT '',
    "avatar"        VARCHAR(500) NOT NULL DEFAULT '',
    "source"        VARCHAR(128) NOT NULL DEFAULT '',
    "grade"         VARCHAR(64)  NOT NULL DEFAULT '',
    "college"       VARCHAR(128) NOT NULL DEFAULT '',
    "college_code"  VARCHAR(64)  NOT NULL DEFAULT '',
    "campus"        VARCHAR(128) NOT NULL DEFAULT '',
    "school_system" VARCHAR(64)  NOT NULL DEFAULT '',
    "graduated_at"  VARCHAR(32)  NOT NULL DEFAULT '',
    "major"         VARCHAR(128) NOT NULL DEFAULT '',
    "class_name"    VARCHAR(128) NOT NULL DEFAULT '',
    "face"          VARCHAR(500) NOT NULL DEFAULT '',
    "created_by"    BIGINT       NOT NULL DEFAULT 0,
    "updated_by"    BIGINT       NOT NULL DEFAULT 0,
    "created_at"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE plugin_linapro_uidentity_cas_account_detail IS 'UIdentity account detail table';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_account_detail."birthday" IS 'Date-only birthday in YYYY-MM-DD format';

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_group (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT          NOT NULL DEFAULT 0,
    "name"       VARCHAR(128) NOT NULL DEFAULT '',
    "alias"      VARCHAR(128) NOT NULL DEFAULT '',
    "created_by" BIGINT       NOT NULL DEFAULT 0,
    "updated_by" BIGINT       NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP    NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_unit (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT          NOT NULL DEFAULT 0,
    "name"       VARCHAR(128) NOT NULL DEFAULT '',
    "alias"      VARCHAR(128) NOT NULL DEFAULT '',
    "code"       VARCHAR(128) NOT NULL DEFAULT '',
    "created_by" BIGINT       NOT NULL DEFAULT 0,
    "updated_by" BIGINT       NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP    NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_container (
    "id"            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"     INT          NOT NULL DEFAULT 0,
    "name"          VARCHAR(128) NOT NULL DEFAULT '',
    "alias"         VARCHAR(128) NOT NULL DEFAULT '',
    "account_count" INT          NOT NULL DEFAULT 0,
    "admin_count"   INT          NOT NULL DEFAULT 0,
    "created_by"    BIGINT       NOT NULL DEFAULT 0,
    "updated_by"    BIGINT       NOT NULL DEFAULT 0,
    "created_at"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"    TIMESTAMP    NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_application (
    "id"           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"    INT          NOT NULL DEFAULT 0,
    "name"         VARCHAR(128) NOT NULL DEFAULT '',
    "alias"        VARCHAR(128) NOT NULL DEFAULT '',
    "client_id"    VARCHAR(128) NOT NULL DEFAULT '',
    "secret_key"   VARCHAR(255) NOT NULL DEFAULT '',
    "access_model" VARCHAR(64)  NOT NULL DEFAULT '',
    "status"       SMALLINT     NOT NULL DEFAULT 0,
    "callback_url" VARCHAR(500) NOT NULL DEFAULT '',
    "whitelist"    TEXT         NOT NULL DEFAULT '',
    "created_by"   BIGINT       NOT NULL DEFAULT 0,
    "updated_by"   BIGINT       NOT NULL DEFAULT 0,
    "created_at"   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"   TIMESTAMP    NULL DEFAULT NULL
);

COMMENT ON COLUMN plugin_linapro_uidentity_cas_application."access_model" IS 'Application access model, for example cas/oauth/ldap';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_application."status" IS 'Application status: 0=disabled, 1=enabled';

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_account_group (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT       NOT NULL DEFAULT 0,
    "account_id" BIGINT    NOT NULL DEFAULT 0,
    "group_id"   BIGINT    NOT NULL DEFAULT 0,
    "created_by" BIGINT    NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_account_unit (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT       NOT NULL DEFAULT 0,
    "account_id" BIGINT    NOT NULL DEFAULT 0,
    "unit_id"    BIGINT    NOT NULL DEFAULT 0,
    "created_by" BIGINT    NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_account_app_role (
    "id"                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"            INT       NOT NULL DEFAULT 0,
    "give_account_id"      BIGINT    NOT NULL DEFAULT 0,
    "empowered_account_id" BIGINT    NOT NULL DEFAULT 0,
    "app_id"               BIGINT    NOT NULL DEFAULT 0,
    "expire_at"            TIMESTAMP NULL DEFAULT NULL,
    "created_by"           BIGINT    NOT NULL DEFAULT 0,
    "updated_by"           BIGINT    NOT NULL DEFAULT 0,
    "created_at"           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"           TIMESTAMP NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_account_app_blacklist (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT          NOT NULL DEFAULT 0,
    "name"       VARCHAR(128) NOT NULL DEFAULT '',
    "app_id"     BIGINT       NOT NULL DEFAULT 0,
    "account_id" BIGINT       NOT NULL DEFAULT 0,
    "effect_at"  TIMESTAMP    NULL DEFAULT NULL,
    "expire_at"  TIMESTAMP    NULL DEFAULT NULL,
    "created_by" BIGINT       NOT NULL DEFAULT 0,
    "updated_by" BIGINT       NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP    NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_group_app_blacklist (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT          NOT NULL DEFAULT 0,
    "name"       VARCHAR(128) NOT NULL DEFAULT '',
    "app_id"     BIGINT       NOT NULL DEFAULT 0,
    "group_id"   BIGINT       NOT NULL DEFAULT 0,
    "effect_at"  TIMESTAMP    NULL DEFAULT NULL,
    "expire_at"  TIMESTAMP    NULL DEFAULT NULL,
    "created_by" BIGINT       NOT NULL DEFAULT 0,
    "updated_by" BIGINT       NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP    NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_pass_rule (
    "id"              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"       INT          NOT NULL DEFAULT 0,
    "name"            VARCHAR(128) NOT NULL DEFAULT '',
    "capital"         SMALLINT     NOT NULL DEFAULT 0,
    "lower"           SMALLINT     NOT NULL DEFAULT 0,
    "number"          SMALLINT     NOT NULL DEFAULT 0,
    "symbol"          SMALLINT     NOT NULL DEFAULT 0,
    "length"          SMALLINT     NOT NULL DEFAULT 8,
    "interval_days"   INT          NOT NULL DEFAULT 0,
    "interval_status" SMALLINT     NOT NULL DEFAULT 0,
    "status"          SMALLINT     NOT NULL DEFAULT 1,
    "created_by"      BIGINT       NOT NULL DEFAULT 0,
    "updated_by"      BIGINT       NOT NULL DEFAULT 0,
    "created_at"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"      TIMESTAMP    NULL DEFAULT NULL
);

COMMENT ON COLUMN plugin_linapro_uidentity_cas_pass_rule."status" IS 'Rule status: 0=disabled, 1=enabled';

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_sms (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT         NOT NULL DEFAULT 0,
    "phone"      VARCHAR(64) NOT NULL DEFAULT '',
    "type"       VARCHAR(64) NOT NULL DEFAULT '',
    "content"    TEXT        NOT NULL DEFAULT '',
    "status"     SMALLINT    NOT NULL DEFAULT 0,
    "resp_msg"   TEXT        NOT NULL DEFAULT '',
    "created_by" BIGINT      NOT NULL DEFAULT 0,
    "updated_by" BIGINT      NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP   NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_cas_login_log (
    "id"                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"         INT          NOT NULL DEFAULT 0,
    "account_id"        BIGINT       NOT NULL DEFAULT 0,
    "choice_account_id" BIGINT       NOT NULL DEFAULT 0,
    "app_id"            BIGINT       NOT NULL DEFAULT 0,
    "ipaddr"            VARCHAR(128) NOT NULL DEFAULT '',
    "login_location"    VARCHAR(255) NOT NULL DEFAULT '',
    "browser"           VARCHAR(128) NOT NULL DEFAULT '',
    "os"                VARCHAR(128) NOT NULL DEFAULT '',
    "platform"          VARCHAR(128) NOT NULL DEFAULT '',
    "login_time"        TIMESTAMP    NULL DEFAULT NULL,
    "remark"            VARCHAR(500) NOT NULL DEFAULT '',
    "msg"               TEXT         NOT NULL DEFAULT '',
    "login_type"        VARCHAR(64)  NOT NULL DEFAULT '',
    "created_by"        BIGINT       NOT NULL DEFAULT 0,
    "updated_by"        BIGINT       NOT NULL DEFAULT 0,
    "created_at"        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"        TIMESTAMP    NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_oauth_log (
    "id"           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"    INT          NOT NULL DEFAULT 0,
    "user_id"      BIGINT       NOT NULL DEFAULT 0,
    "app_id"       BIGINT       NOT NULL DEFAULT 0,
    "redirect_uri" VARCHAR(500) NOT NULL DEFAULT '',
    "scope"        VARCHAR(255) NOT NULL DEFAULT '',
    "created_by"   BIGINT       NOT NULL DEFAULT 0,
    "updated_by"   BIGINT       NOT NULL DEFAULT 0,
    "created_at"   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"   TIMESTAMP    NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_oauth_token (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT          NOT NULL DEFAULT 0,
    "expired_at" TIMESTAMP    NULL DEFAULT NULL,
    "code"       VARCHAR(255) NOT NULL DEFAULT '',
    "access"     VARCHAR(255) NOT NULL DEFAULT '',
    "refresh"    VARCHAR(255) NOT NULL DEFAULT '',
    "data"       TEXT         NOT NULL DEFAULT '',
    "created_by" BIGINT       NOT NULL DEFAULT 0,
    "updated_by" BIGINT       NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP    NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_account_change_log (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT          NOT NULL DEFAULT 0,
    "account_id" BIGINT       NOT NULL DEFAULT 0,
    "table_name" VARCHAR(128) NOT NULL DEFAULT '',
    "action"     VARCHAR(64)  NOT NULL DEFAULT '',
    "data_old"   TEXT         NOT NULL DEFAULT '',
    "data_new"   TEXT         NOT NULL DEFAULT '',
    "err_msg"    TEXT         NOT NULL DEFAULT '',
    "err_number" VARCHAR(128) NOT NULL DEFAULT '',
    "created_by" BIGINT       NOT NULL DEFAULT 0,
    "updated_by" BIGINT       NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP    NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_account_active_log (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT          NOT NULL DEFAULT 0,
    "number"     VARCHAR(128) NOT NULL DEFAULT '',
    "phone"      VARCHAR(64)  NOT NULL DEFAULT '',
    "wechat"     VARCHAR(128) NOT NULL DEFAULT '',
    "type"       SMALLINT     NOT NULL DEFAULT 0,
    "created_by" BIGINT       NOT NULL DEFAULT 0,
    "updated_by" BIGINT       NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP    NULL DEFAULT NULL
);

COMMENT ON TABLE plugin_linapro_uidentity_cas_account_active_log IS 'Legacy-compatible account activation and Wechat binding audit log';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_account_active_log."type" IS 'Legacy activation log type: 0=activation or Wechat rebind callback, 1=union ID bind';

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_sys_job (
    "job_id"          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"       INT          NOT NULL DEFAULT 0,
    "job_name"        VARCHAR(255) NOT NULL DEFAULT '',
    "job_group"       VARCHAR(255) NOT NULL DEFAULT '',
    "job_type"        SMALLINT     NOT NULL DEFAULT 0,
    "cron_expression" VARCHAR(255) NOT NULL DEFAULT '',
    "invoke_target"   VARCHAR(500) NOT NULL DEFAULT '',
    "args"            VARCHAR(500) NOT NULL DEFAULT '',
    "misfire_policy"  SMALLINT     NOT NULL DEFAULT 0,
    "concurrent"      SMALLINT     NOT NULL DEFAULT 0,
    "status"          SMALLINT     NOT NULL DEFAULT 1,
    "entry_id"        BIGINT       NOT NULL DEFAULT 0,
    "created_by"      BIGINT       NOT NULL DEFAULT 0,
    "updated_by"      BIGINT       NOT NULL DEFAULT 0,
    "created_at"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"      TIMESTAMP    NULL DEFAULT NULL
);

COMMENT ON TABLE plugin_linapro_uidentity_cas_sys_job IS 'Legacy-compatible UIdentity job definition table';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_sys_job."job_type" IS 'Job type: 1=http, 2=exec or plugin-defined executor';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_sys_job."misfire_policy" IS 'Misfire policy copied from legacy sys_job semantics';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_sys_job."concurrent" IS 'Concurrent execution flag: 0=disallow, 1=allow';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_sys_job."status" IS 'Job status: 1=disabled, 2=enabled';
COMMENT ON COLUMN plugin_linapro_uidentity_cas_sys_job."entry_id" IS 'Runtime scheduler entry ID, 0 means not scheduled';

CREATE TABLE IF NOT EXISTS plugin_linapro_uidentity_cas_job_log (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT          NOT NULL DEFAULT 0,
    "job_id"     BIGINT       NOT NULL DEFAULT 0,
    "job_name"   VARCHAR(128) NOT NULL DEFAULT '',
    "start_at"   TIMESTAMP    NULL DEFAULT NULL,
    "end_at"     TIMESTAMP    NULL DEFAULT NULL,
    "create_num" BIGINT       NOT NULL DEFAULT 0,
    "update_num" BIGINT       NOT NULL DEFAULT 0,
    "delete_num" BIGINT       NOT NULL DEFAULT 0,
    "err_num"    BIGINT       NOT NULL DEFAULT 0,
    "created_by" BIGINT       NOT NULL DEFAULT 0,
    "updated_by" BIGINT       NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP    NULL DEFAULT NULL
);

COMMENT ON TABLE plugin_linapro_uidentity_cas_job_log IS 'Legacy-compatible UIdentity job execution log table';

CREATE UNIQUE INDEX IF NOT EXISTS uk_pluicas_account_number ON plugin_linapro_uidentity_cas_account ("tenant_id", "number") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_account_phone ON plugin_linapro_uidentity_cas_account ("tenant_id", "phone") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_account_status ON plugin_linapro_uidentity_cas_account ("tenant_id", "status") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_account_container ON plugin_linapro_uidentity_cas_account ("tenant_id", "container_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_account_unit ON plugin_linapro_uidentity_cas_account ("tenant_id", "unit_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_detail_tenant ON plugin_linapro_uidentity_cas_account_detail ("tenant_id", "account_id");
CREATE INDEX IF NOT EXISTS idx_pluicas_detail_graduated_at ON plugin_linapro_uidentity_cas_account_detail ("tenant_id", "graduated_at", "account_id");

CREATE UNIQUE INDEX IF NOT EXISTS uk_pluicas_group_name ON plugin_linapro_uidentity_cas_group ("tenant_id", "name") WHERE "deleted_at" IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_pluicas_unit_code ON plugin_linapro_uidentity_cas_unit ("tenant_id", "code") WHERE "deleted_at" IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_pluicas_container_name ON plugin_linapro_uidentity_cas_container ("tenant_id", "name") WHERE "deleted_at" IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_pluicas_app_client ON plugin_linapro_uidentity_cas_application ("tenant_id", "client_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_app_status ON plugin_linapro_uidentity_cas_application ("tenant_id", "status") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_app_access ON plugin_linapro_uidentity_cas_application ("tenant_id", "access_model") WHERE "deleted_at" IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uk_pluicas_account_group ON plugin_linapro_uidentity_cas_account_group ("tenant_id", "account_id", "group_id");
CREATE INDEX IF NOT EXISTS idx_pluicas_account_group_group ON plugin_linapro_uidentity_cas_account_group ("tenant_id", "group_id");
CREATE UNIQUE INDEX IF NOT EXISTS uk_pluicas_account_unit ON plugin_linapro_uidentity_cas_account_unit ("tenant_id", "account_id", "unit_id");
CREATE INDEX IF NOT EXISTS idx_pluicas_account_unit_unit ON plugin_linapro_uidentity_cas_account_unit ("tenant_id", "unit_id");

CREATE INDEX IF NOT EXISTS idx_pluicas_role_app ON plugin_linapro_uidentity_cas_account_app_role ("tenant_id", "app_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_role_give ON plugin_linapro_uidentity_cas_account_app_role ("tenant_id", "give_account_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_role_empowered ON plugin_linapro_uidentity_cas_account_app_role ("tenant_id", "empowered_account_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_account_bl_app ON plugin_linapro_uidentity_cas_account_app_blacklist ("tenant_id", "app_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_account_bl_account ON plugin_linapro_uidentity_cas_account_app_blacklist ("tenant_id", "account_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_group_bl_app ON plugin_linapro_uidentity_cas_group_app_blacklist ("tenant_id", "app_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_group_bl_group ON plugin_linapro_uidentity_cas_group_app_blacklist ("tenant_id", "group_id") WHERE "deleted_at" IS NULL;

CREATE INDEX IF NOT EXISTS idx_pluicas_pass_rule_status ON plugin_linapro_uidentity_cas_pass_rule ("tenant_id", "status") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_sms_phone ON plugin_linapro_uidentity_cas_sms ("tenant_id", "phone", "created_at") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_cas_log_time ON plugin_linapro_uidentity_cas_cas_login_log ("tenant_id", "login_time") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_cas_log_account ON plugin_linapro_uidentity_cas_cas_login_log ("tenant_id", "account_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_cas_log_app ON plugin_linapro_uidentity_cas_cas_login_log ("tenant_id", "app_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_oauth_log_user ON plugin_linapro_uidentity_cas_oauth_log ("tenant_id", "user_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_oauth_log_app ON plugin_linapro_uidentity_cas_oauth_log ("tenant_id", "app_id") WHERE "deleted_at" IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_pluicas_oauth_code ON plugin_linapro_uidentity_cas_oauth_token ("tenant_id", "code") WHERE "deleted_at" IS NULL AND "code" <> '';
CREATE UNIQUE INDEX IF NOT EXISTS uk_pluicas_oauth_access ON plugin_linapro_uidentity_cas_oauth_token ("tenant_id", "access") WHERE "deleted_at" IS NULL AND "access" <> '';
CREATE INDEX IF NOT EXISTS idx_pluicas_change_account ON plugin_linapro_uidentity_cas_account_change_log ("tenant_id", "account_id", "created_at") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_active_log_number ON plugin_linapro_uidentity_cas_account_active_log ("tenant_id", "number", "created_at") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_active_log_wechat ON plugin_linapro_uidentity_cas_account_active_log ("tenant_id", "wechat", "created_at") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_sys_job_status ON plugin_linapro_uidentity_cas_sys_job ("tenant_id", "status") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_sys_job_group ON plugin_linapro_uidentity_cas_sys_job ("tenant_id", "job_group") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_sys_job_entry ON plugin_linapro_uidentity_cas_sys_job ("tenant_id", "entry_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_job_log_job ON plugin_linapro_uidentity_cas_job_log ("tenant_id", "job_id", "created_at") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_pluicas_job_log_time ON plugin_linapro_uidentity_cas_job_log ("tenant_id", "start_at", "end_at") WHERE "deleted_at" IS NULL;

INSERT INTO plugin_linapro_uidentity_cas_pass_rule (
    "tenant_id", "name", "capital", "lower", "number", "symbol", "length",
    "interval_days", "interval_status", "status"
)
SELECT 0, 'Default password policy', 1, 1, 1, 0, 8, 0, 0, 1
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_linapro_uidentity_cas_pass_rule
    WHERE "tenant_id" = 0 AND "name" = 'Default password policy' AND "deleted_at" IS NULL
);
