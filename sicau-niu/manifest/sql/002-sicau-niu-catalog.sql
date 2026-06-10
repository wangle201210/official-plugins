-- ------------------------------------------------------------
-- 002 sicau-niu catalog SQL file
-- 002 sicau-niu 内容资产 SQL 文件
-- Purpose: Cattle, iron-cow, card and quote content assets for the activity (C2 niu-catalog-admin).
-- 用途:寻牛活动的牛、铁牛、卡片、校史金句内容资产(C2 niu-catalog-admin)。
-- Dialect: PostgreSQL. Idempotent: safe to re-run.
-- ------------------------------------------------------------

-- 牛表:活动虚拟牛,含类型、特殊子类、关联院系、GPS 锚点、上线时间与状态。
CREATE TABLE IF NOT EXISTS plugin_sicau_niu_niu (
    "id"               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "code"             VARCHAR(32) NOT NULL DEFAULT '',
    "niu_type"         VARCHAR(16) NOT NULL DEFAULT '',
    "special_subtype"  VARCHAR(16) NOT NULL DEFAULT '',
    "name"             VARCHAR(64) NOT NULL DEFAULT '',
    "college_id"       BIGINT NOT NULL DEFAULT 0,
    "lat"              DOUBLE PRECISION NOT NULL DEFAULT 0,
    "lng"              DOUBLE PRECISION NOT NULL DEFAULT 0,
    "online_at"        TIMESTAMPTZ,
    "visible_weekdays" VARCHAR(20) NOT NULL DEFAULT '',
    "visible_start"    VARCHAR(5) NOT NULL DEFAULT '',
    "visible_end"      VARCHAR(5) NOT NULL DEFAULT '',
    "status"           VARCHAR(16) NOT NULL DEFAULT 'inactive',
    "created_at"       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"       TIMESTAMP
);

COMMENT ON TABLE plugin_sicau_niu_niu IS 'sicau-niu activity cattle table';
COMMENT ON COLUMN plugin_sicau_niu_niu."code" IS 'Cattle serial code, unique among active rows';
COMMENT ON COLUMN plugin_sicau_niu_niu."niu_type" IS 'Cattle type: common, special';
COMMENT ON COLUMN plugin_sicau_niu_niu."special_subtype" IS 'Special subtype: college, contribution, alumni, spirit; empty for common';
COMMENT ON COLUMN plugin_sicau_niu_niu."name" IS 'Cattle name, used by special cattle';
COMMENT ON COLUMN plugin_sicau_niu_niu."college_id" IS 'Linked college ID for college cattle, 0 means none';
COMMENT ON COLUMN plugin_sicau_niu_niu."lat" IS 'GPS latitude anchor';
COMMENT ON COLUMN plugin_sicau_niu_niu."lng" IS 'GPS longitude anchor';
COMMENT ON COLUMN plugin_sicau_niu_niu."online_at" IS 'Scheduled online time; NULL means not yet online';
COMMENT ON COLUMN plugin_sicau_niu_niu."visible_weekdays" IS 'Optional visible weekdays, e.g. 1,3,5';
COMMENT ON COLUMN plugin_sicau_niu_niu."visible_start" IS 'Optional visible window start HH:MM';
COMMENT ON COLUMN plugin_sicau_niu_niu."visible_end" IS 'Optional visible window end HH:MM';
COMMENT ON COLUMN plugin_sicau_niu_niu."status" IS 'Cattle status: inactive, active; managed by activation flow';
COMMENT ON COLUMN plugin_sicau_niu_niu."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_sicau_niu_niu."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_sicau_niu_niu."deleted_at" IS 'Soft-delete time, NULL means active';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_niu_code
    ON plugin_sicau_niu_niu ("code") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_sicau_niu_niu_online
    ON plugin_sicau_niu_niu ("online_at") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_sicau_niu_niu_college
    ON plugin_sicau_niu_niu ("college_id");

-- 铁牛表:带定位装置的实体铁牛,实时经纬由喂草加成能力(C4)写入。
CREATE TABLE IF NOT EXISTS plugin_sicau_niu_iron (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "code"       VARCHAR(64) NOT NULL DEFAULT '',
    "name"       VARCHAR(64) NOT NULL DEFAULT '',
    "last_lat"   DOUBLE PRECISION,
    "last_lng"   DOUBLE PRECISION,
    "located_at" TIMESTAMP,
    "remark"     VARCHAR(255) NOT NULL DEFAULT '',
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP
);

COMMENT ON TABLE plugin_sicau_niu_iron IS 'sicau-niu iron-cow registry table';
COMMENT ON COLUMN plugin_sicau_niu_iron."code" IS 'Iron-cow device identifier, unique among active rows';
COMMENT ON COLUMN plugin_sicau_niu_iron."name" IS 'Iron-cow display name';
COMMENT ON COLUMN plugin_sicau_niu_iron."last_lat" IS 'Latest GPS latitude pulled by the bonus flow (C4)';
COMMENT ON COLUMN plugin_sicau_niu_iron."last_lng" IS 'Latest GPS longitude pulled by the bonus flow (C4)';
COMMENT ON COLUMN plugin_sicau_niu_iron."located_at" IS 'Latest location pull time';
COMMENT ON COLUMN plugin_sicau_niu_iron."remark" IS 'Iron-cow remark';
COMMENT ON COLUMN plugin_sicau_niu_iron."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_sicau_niu_iron."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_sicau_niu_iron."deleted_at" IS 'Soft-delete time, NULL means active';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_iron_code
    ON plugin_sicau_niu_iron ("code") WHERE "deleted_at" IS NULL;

-- 卡片表:校史主卡,与牛保持 1 牛 1 卡关联。
CREATE TABLE IF NOT EXISTS plugin_sicau_niu_card (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "niu_id"     BIGINT NOT NULL DEFAULT 0,
    "category"   VARCHAR(16) NOT NULL DEFAULT '',
    "title"      VARCHAR(128) NOT NULL DEFAULT '',
    "content"    TEXT NOT NULL DEFAULT '',
    "image_path" VARCHAR(500) NOT NULL DEFAULT '',
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP
);

COMMENT ON TABLE plugin_sicau_niu_card IS 'sicau-niu campus-history card table';
COMMENT ON COLUMN plugin_sicau_niu_card."niu_id" IS 'Owning cattle ID; one card per cattle';
COMMENT ON COLUMN plugin_sicau_niu_card."category" IS 'Card category: person, event, research, college, spirit';
COMMENT ON COLUMN plugin_sicau_niu_card."title" IS 'Card title';
COMMENT ON COLUMN plugin_sicau_niu_card."content" IS 'Card content text';
COMMENT ON COLUMN plugin_sicau_niu_card."image_path" IS 'Card image storage path uploaded via host file management';
COMMENT ON COLUMN plugin_sicau_niu_card."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_sicau_niu_card."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_sicau_niu_card."deleted_at" IS 'Soft-delete time, NULL means active';

-- 每头牛活跃集合内至多一张主卡(niu_id=0 表示未归属,不参与唯一约束)。
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_card_niu
    ON plugin_sicau_niu_card ("niu_id") WHERE "deleted_at" IS NULL AND "niu_id" > 0;
CREATE INDEX IF NOT EXISTS idx_sicau_niu_card_category
    ON plugin_sicau_niu_card ("category");

-- 校史金句表:喂草/点击随机播放的金句池。
CREATE TABLE IF NOT EXISTS plugin_sicau_niu_quote (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "content"    VARCHAR(500) NOT NULL DEFAULT '',
    "enabled"    INT NOT NULL DEFAULT 1,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP
);

COMMENT ON TABLE plugin_sicau_niu_quote IS 'sicau-niu campus-history quote pool table';
COMMENT ON COLUMN plugin_sicau_niu_quote."content" IS 'Quote text';
COMMENT ON COLUMN plugin_sicau_niu_quote."enabled" IS 'Whether the quote participates in random playback: 1 enabled, 0 disabled';
COMMENT ON COLUMN plugin_sicau_niu_quote."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_sicau_niu_quote."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_sicau_niu_quote."deleted_at" IS 'Soft-delete time, NULL means active';

CREATE INDEX IF NOT EXISTS idx_sicau_niu_quote_enabled
    ON plugin_sicau_niu_quote ("enabled");
