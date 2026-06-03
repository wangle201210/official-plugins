-- ------------------------------------------------------------
-- 001 sicau-niu identity SQL file
-- 001 sicau-niu 身份能力 SQL 文件
-- Purpose: Player accounts and college dictionary for the find-the-cow activity (C1 niu-identity).
-- 用途：寻牛活动的玩家账户与院系字典（C1 niu-identity）。
-- Dialect: PostgreSQL. Idempotent: safe to re-run.
-- ------------------------------------------------------------

-- 院系字典表（运营手动维护，供玩家身份选择与后续院系榜聚合复用）。
CREATE TABLE IF NOT EXISTS plugin_sicau_niu_college (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "name"       VARCHAR(64) NOT NULL DEFAULT '',
    "sort"       INT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP
);

COMMENT ON TABLE plugin_sicau_niu_college IS 'sicau-niu college dictionary table';
COMMENT ON COLUMN plugin_sicau_niu_college."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_sicau_niu_college."name" IS 'College name';
COMMENT ON COLUMN plugin_sicau_niu_college."sort" IS 'Display sort order, smaller first';
COMMENT ON COLUMN plugin_sicau_niu_college."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_sicau_niu_college."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_sicau_niu_college."deleted_at" IS 'Soft-delete time, NULL means active';

-- 院系名称在未删除集合内唯一。
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_college_name
    ON plugin_sicau_niu_college ("name") WHERE "deleted_at" IS NULL;
-- 院系下拉按排序读取。
CREATE INDEX IF NOT EXISTS idx_sicau_niu_college_sort
    ON plugin_sicau_niu_college ("sort");

-- 寻牛玩家账户表（微信 openid + 手机号，一机一号）。
CREATE TABLE IF NOT EXISTS plugin_sicau_niu_user (
    "id"                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "openid"             VARCHAR(64) NOT NULL DEFAULT '',
    "phone"              VARCHAR(20) NOT NULL DEFAULT '',
    "nickname"           VARCHAR(64) NOT NULL DEFAULT '',
    "avatar"             VARCHAR(500) NOT NULL DEFAULT '',
    "identity_type"      VARCHAR(20) NOT NULL DEFAULT '',
    "college_id"         BIGINT NOT NULL DEFAULT 0,
    "grade"              INT NOT NULL DEFAULT 0,
    "graduation_year"    INT NOT NULL DEFAULT 0,
    "device_fingerprint" VARCHAR(128) NOT NULL DEFAULT '',
    "created_at"         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"         TIMESTAMP
);

COMMENT ON TABLE plugin_sicau_niu_user IS 'sicau-niu player account table';
COMMENT ON COLUMN plugin_sicau_niu_user."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_sicau_niu_user."openid" IS 'WeChat openid, unique per active player';
COMMENT ON COLUMN plugin_sicau_niu_user."phone" IS 'Bound phone number, unique per active player (one-phone-one-account)';
COMMENT ON COLUMN plugin_sicau_niu_user."nickname" IS 'Player nickname';
COMMENT ON COLUMN plugin_sicau_niu_user."avatar" IS 'Player avatar URL';
COMMENT ON COLUMN plugin_sicau_niu_user."identity_type" IS 'Identity tag: student / alumni / friend';
COMMENT ON COLUMN plugin_sicau_niu_user."college_id" IS 'Selected college ID, 0 means none';
COMMENT ON COLUMN plugin_sicau_niu_user."grade" IS 'Grade number filled by student, 0 means unset';
COMMENT ON COLUMN plugin_sicau_niu_user."graduation_year" IS 'Graduation year filled by alumni, 0 means unset';
COMMENT ON COLUMN plugin_sicau_niu_user."device_fingerprint" IS 'Lightweight device fingerprint for risk control';
COMMENT ON COLUMN plugin_sicau_niu_user."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_sicau_niu_user."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_sicau_niu_user."deleted_at" IS 'Soft-delete time, NULL means active';

-- openid 在未删除玩家集合内唯一（登录建档锚点）。
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_user_openid
    ON plugin_sicau_niu_user ("openid") WHERE "deleted_at" IS NULL;
-- 手机号在已绑定且未删除的玩家集合内唯一（一机一号，空号不参与唯一约束）。
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_user_phone
    ON plugin_sicau_niu_user ("phone") WHERE "deleted_at" IS NULL AND "phone" <> '';
-- 运营按院系过滤/聚合与删除引用校验依赖该索引。
CREATE INDEX IF NOT EXISTS idx_sicau_niu_user_college
    ON plugin_sicau_niu_user ("college_id");
