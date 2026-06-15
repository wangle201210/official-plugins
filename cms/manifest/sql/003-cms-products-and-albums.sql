-- 003: CMS product center and photo album schema with seed dictionaries
-- 003：CMS 产品中心与相册数据结构及字典种子

-- Purpose: Stores CMS products with gallery images and display pricing.
-- 用途：存储 CMS 产品及其多图与展示价格。
CREATE TABLE IF NOT EXISTS plugin_cms_product (
    "id"           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "category_id"  BIGINT        NOT NULL DEFAULT 0,
    "name"         VARCHAR(255)  NOT NULL DEFAULT '',
    "slug"         VARCHAR(128)  NOT NULL DEFAULT '',
    "summary"      VARCHAR(1000) NOT NULL DEFAULT '',
    "cover"        VARCHAR(500)  NOT NULL DEFAULT '',
    "gallery"      TEXT          NOT NULL DEFAULT '',
    "price"        VARCHAR(128)  NOT NULL DEFAULT '',
    "spec"         VARCHAR(500)  NOT NULL DEFAULT '',
    "content"      TEXT          NOT NULL DEFAULT '',
    "keywords"     VARCHAR(500)  NOT NULL DEFAULT '',
    "description"  VARCHAR(1000) NOT NULL DEFAULT '',
    "sort"         INT           NOT NULL DEFAULT 0,
    "status"       SMALLINT      NOT NULL DEFAULT 0,
    "is_top"       SMALLINT      NOT NULL DEFAULT 0,
    "is_recommend" SMALLINT      NOT NULL DEFAULT 0,
    "views"        BIGINT        NOT NULL DEFAULT 0,
    "published_at" TIMESTAMP     NULL DEFAULT NULL,
    "created_by"   BIGINT        NOT NULL DEFAULT 0,
    "updated_by"   BIGINT        NOT NULL DEFAULT 0,
    "created_at"   TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"   TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"   TIMESTAMP     NULL DEFAULT NULL
);

COMMENT ON TABLE plugin_cms_product IS 'CMS product';
COMMENT ON COLUMN plugin_cms_product."id" IS 'Product ID';
COMMENT ON COLUMN plugin_cms_product."category_id" IS 'Category ID';
COMMENT ON COLUMN plugin_cms_product."name" IS 'Product name';
COMMENT ON COLUMN plugin_cms_product."slug" IS 'Public URL slug';
COMMENT ON COLUMN plugin_cms_product."summary" IS 'Product summary';
COMMENT ON COLUMN plugin_cms_product."cover" IS 'Cover image URL';
COMMENT ON COLUMN plugin_cms_product."gallery" IS 'Gallery image URL list as JSON array text';
COMMENT ON COLUMN plugin_cms_product."price" IS 'Display price text';
COMMENT ON COLUMN plugin_cms_product."spec" IS 'Specification summary';
COMMENT ON COLUMN plugin_cms_product."content" IS 'Product detail HTML';
COMMENT ON COLUMN plugin_cms_product."keywords" IS 'SEO keywords';
COMMENT ON COLUMN plugin_cms_product."description" IS 'SEO description';
COMMENT ON COLUMN plugin_cms_product."sort" IS 'Display order';
COMMENT ON COLUMN plugin_cms_product."status" IS 'Status: 0=draft, 1=published';
COMMENT ON COLUMN plugin_cms_product."is_top" IS 'Top flag: 0=no, 1=yes';
COMMENT ON COLUMN plugin_cms_product."is_recommend" IS 'Recommend flag: 0=no, 1=yes';
COMMENT ON COLUMN plugin_cms_product."views" IS 'View count';
COMMENT ON COLUMN plugin_cms_product."published_at" IS 'Publication time';
COMMENT ON COLUMN plugin_cms_product."created_by" IS 'Creator user ID';
COMMENT ON COLUMN plugin_cms_product."updated_by" IS 'Updater user ID';
COMMENT ON COLUMN plugin_cms_product."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_cms_product."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_cms_product."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_cms_product_slug ON plugin_cms_product ("slug");
CREATE INDEX IF NOT EXISTS idx_plugin_cms_product_category ON plugin_cms_product ("category_id");
CREATE INDEX IF NOT EXISTS idx_plugin_cms_product_status_publish ON plugin_cms_product ("status", "published_at");
CREATE INDEX IF NOT EXISTS idx_plugin_cms_product_sort ON plugin_cms_product ("sort");

-- Purpose: Stores CMS photo albums and presentation metadata.
-- 用途：存储 CMS 相册及展示元数据。
CREATE TABLE IF NOT EXISTS plugin_cms_album (
    "id"          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "category_id" BIGINT        NOT NULL DEFAULT 0,
    "name"        VARCHAR(255)  NOT NULL DEFAULT '',
    "cover"       VARCHAR(500)  NOT NULL DEFAULT '',
    "description" VARCHAR(1000) NOT NULL DEFAULT '',
    "sort"        INT           NOT NULL DEFAULT 0,
    "status"      SMALLINT      NOT NULL DEFAULT 1,
    "created_by"  BIGINT        NOT NULL DEFAULT 0,
    "updated_by"  BIGINT        NOT NULL DEFAULT 0,
    "created_at"  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"  TIMESTAMP     NULL DEFAULT NULL
);

COMMENT ON TABLE plugin_cms_album IS 'CMS photo album';
COMMENT ON COLUMN plugin_cms_album."id" IS 'Album ID';
COMMENT ON COLUMN plugin_cms_album."category_id" IS 'Category ID';
COMMENT ON COLUMN plugin_cms_album."name" IS 'Album name';
COMMENT ON COLUMN plugin_cms_album."cover" IS 'Cover image URL';
COMMENT ON COLUMN plugin_cms_album."description" IS 'Album description';
COMMENT ON COLUMN plugin_cms_album."sort" IS 'Display order';
COMMENT ON COLUMN plugin_cms_album."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN plugin_cms_album."created_by" IS 'Creator user ID';
COMMENT ON COLUMN plugin_cms_album."updated_by" IS 'Updater user ID';
COMMENT ON COLUMN plugin_cms_album."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_cms_album."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_cms_album."deleted_at" IS 'Deletion time';

CREATE INDEX IF NOT EXISTS idx_plugin_cms_album_category ON plugin_cms_album ("category_id");
CREATE INDEX IF NOT EXISTS idx_plugin_cms_album_status_sort ON plugin_cms_album ("status", "sort");

-- Purpose: Stores ordered CMS album images replaced as a whole on album save.
-- 用途：存储相册图片，整册保存时全量替换。
CREATE TABLE IF NOT EXISTS plugin_cms_album_image (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "album_id"   BIGINT       NOT NULL DEFAULT 0,
    "url"        VARCHAR(500) NOT NULL DEFAULT '',
    "title"      VARCHAR(255) NOT NULL DEFAULT '',
    "sort"       INT          NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE plugin_cms_album_image IS 'CMS album image';
COMMENT ON COLUMN plugin_cms_album_image."id" IS 'Image ID';
COMMENT ON COLUMN plugin_cms_album_image."album_id" IS 'Album ID';
COMMENT ON COLUMN plugin_cms_album_image."url" IS 'Image URL';
COMMENT ON COLUMN plugin_cms_album_image."title" IS 'Image title';
COMMENT ON COLUMN plugin_cms_album_image."sort" IS 'Display order';
COMMENT ON COLUMN plugin_cms_album_image."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_cms_album_image."updated_at" IS 'Update time';

CREATE INDEX IF NOT EXISTS idx_plugin_cms_album_image_album_sort ON plugin_cms_album_image ("album_id", "sort");

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_category_type', '产品栏目', '4', 4, 'processing', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_category_type', '相册栏目', '5', 5, 'purple', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('CMS 产品状态', 'cms_product_status', 1, 1, 'CMS product status options', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_product_status', '草稿', '0', 1, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_product_status', '已发布', '1', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
