-- 001: CMS plugin schema and seed dictionaries
-- 001：CMS 插件数据结构与字典种子

CREATE TABLE IF NOT EXISTS plugin_cms_site (
    "id"          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "site_key"    VARCHAR(64)  NOT NULL DEFAULT 'default',
    "name"        VARCHAR(128) NOT NULL DEFAULT '',
    "logo"        VARCHAR(500) NOT NULL DEFAULT '',
    "weixin"      VARCHAR(500) NOT NULL DEFAULT '',
    "domain"      VARCHAR(255) NOT NULL DEFAULT '',
    "slogan"      VARCHAR(255) NOT NULL DEFAULT '',
    "keywords"    VARCHAR(500) NOT NULL DEFAULT '',
    "description" VARCHAR(1000) NOT NULL DEFAULT '',
    "icp"         VARCHAR(128) NOT NULL DEFAULT '',
    "contact"     VARCHAR(128) NOT NULL DEFAULT '',
    "phone"       VARCHAR(64)  NOT NULL DEFAULT '',
    "email"       VARCHAR(128) NOT NULL DEFAULT '',
    "address"     VARCHAR(255) NOT NULL DEFAULT '',
    "status"      SMALLINT     NOT NULL DEFAULT 1,
    "created_by"  BIGINT       NOT NULL DEFAULT 0,
    "updated_by"  BIGINT       NOT NULL DEFAULT 0,
    "created_at"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"  TIMESTAMP    NULL DEFAULT NULL
);

COMMENT ON TABLE plugin_cms_site IS 'CMS site settings';
COMMENT ON COLUMN plugin_cms_site."id" IS 'Site ID';
COMMENT ON COLUMN plugin_cms_site."site_key" IS 'Stable site key';
COMMENT ON COLUMN plugin_cms_site."name" IS 'Site name';
COMMENT ON COLUMN plugin_cms_site."logo" IS 'Site logo URL';
COMMENT ON COLUMN plugin_cms_site."weixin" IS 'WeChat QR code image URL';
COMMENT ON COLUMN plugin_cms_site."domain" IS 'Primary site domain';
COMMENT ON COLUMN plugin_cms_site."slogan" IS 'Site slogan';
COMMENT ON COLUMN plugin_cms_site."keywords" IS 'SEO keywords';
COMMENT ON COLUMN plugin_cms_site."description" IS 'SEO description';
COMMENT ON COLUMN plugin_cms_site."icp" IS 'ICP record number';
COMMENT ON COLUMN plugin_cms_site."contact" IS 'Contact person';
COMMENT ON COLUMN plugin_cms_site."phone" IS 'Contact phone';
COMMENT ON COLUMN plugin_cms_site."email" IS 'Contact email';
COMMENT ON COLUMN plugin_cms_site."address" IS 'Contact address';
COMMENT ON COLUMN plugin_cms_site."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN plugin_cms_site."created_by" IS 'Creator user ID';
COMMENT ON COLUMN plugin_cms_site."updated_by" IS 'Updater user ID';
COMMENT ON COLUMN plugin_cms_site."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_cms_site."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_cms_site."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_cms_site_key ON plugin_cms_site ("site_key");
CREATE INDEX IF NOT EXISTS idx_plugin_cms_site_status ON plugin_cms_site ("status");

CREATE TABLE IF NOT EXISTS plugin_cms_category (
    "id"               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "parent_id"        BIGINT       NOT NULL DEFAULT 0,
    "code"             VARCHAR(64)  NOT NULL DEFAULT '',
    "name"             VARCHAR(128) NOT NULL DEFAULT '',
    "type"             SMALLINT     NOT NULL DEFAULT 1,
    "path"             VARCHAR(255) NOT NULL DEFAULT '',
    "list_template"    VARCHAR(128) NOT NULL DEFAULT '',
    "content_template" VARCHAR(128) NOT NULL DEFAULT '',
    "cover"            VARCHAR(500) NOT NULL DEFAULT '',
    "outlink"          VARCHAR(500) NOT NULL DEFAULT '',
    "title"            VARCHAR(255) NOT NULL DEFAULT '',
    "keywords"         VARCHAR(500) NOT NULL DEFAULT '',
    "description"      VARCHAR(1000) NOT NULL DEFAULT '',
    "sort"             INT          NOT NULL DEFAULT 0,
    "status"           SMALLINT     NOT NULL DEFAULT 1,
    "created_by"       BIGINT       NOT NULL DEFAULT 0,
    "updated_by"       BIGINT       NOT NULL DEFAULT 0,
    "created_at"       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"       TIMESTAMP    NULL DEFAULT NULL
);

COMMENT ON TABLE plugin_cms_category IS 'CMS category tree';
COMMENT ON COLUMN plugin_cms_category."id" IS 'Category ID';
COMMENT ON COLUMN plugin_cms_category."parent_id" IS 'Parent category ID';
COMMENT ON COLUMN plugin_cms_category."code" IS 'Stable category code';
COMMENT ON COLUMN plugin_cms_category."name" IS 'Category name';
COMMENT ON COLUMN plugin_cms_category."type" IS 'Category type: 1=list, 2=single page, 3=external link';
COMMENT ON COLUMN plugin_cms_category."path" IS 'Public category path';
COMMENT ON COLUMN plugin_cms_category."list_template" IS 'Public list template file';
COMMENT ON COLUMN plugin_cms_category."content_template" IS 'Public content/detail template file';
COMMENT ON COLUMN plugin_cms_category."cover" IS 'Category cover image URL';
COMMENT ON COLUMN plugin_cms_category."outlink" IS 'External link URL';
COMMENT ON COLUMN plugin_cms_category."title" IS 'SEO title';
COMMENT ON COLUMN plugin_cms_category."keywords" IS 'SEO keywords';
COMMENT ON COLUMN plugin_cms_category."description" IS 'SEO description';
COMMENT ON COLUMN plugin_cms_category."sort" IS 'Display order';
COMMENT ON COLUMN plugin_cms_category."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN plugin_cms_category."created_by" IS 'Creator user ID';
COMMENT ON COLUMN plugin_cms_category."updated_by" IS 'Updater user ID';
COMMENT ON COLUMN plugin_cms_category."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_cms_category."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_cms_category."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_cms_category_code ON plugin_cms_category ("code");
CREATE INDEX IF NOT EXISTS idx_plugin_cms_category_parent ON plugin_cms_category ("parent_id");
CREATE INDEX IF NOT EXISTS idx_plugin_cms_category_status_sort ON plugin_cms_category ("status", "sort");

CREATE TABLE IF NOT EXISTS plugin_cms_article (
    "id"            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "category_id"   BIGINT       NOT NULL DEFAULT 0,
    "title"         VARCHAR(255) NOT NULL DEFAULT '',
    "subtitle"      VARCHAR(255) NOT NULL DEFAULT '',
    "slug"          VARCHAR(128) NOT NULL DEFAULT '',
    "summary"       VARCHAR(1000) NOT NULL DEFAULT '',
    "cover"         VARCHAR(500) NOT NULL DEFAULT '',
    "author"        VARCHAR(128) NOT NULL DEFAULT '',
    "source"        VARCHAR(128) NOT NULL DEFAULT '',
    "content"       TEXT         NOT NULL DEFAULT '',
    "tags"          VARCHAR(500) NOT NULL DEFAULT '',
    "keywords"      VARCHAR(500) NOT NULL DEFAULT '',
    "description"   VARCHAR(1000) NOT NULL DEFAULT '',
    "sort"          INT          NOT NULL DEFAULT 0,
    "status"        SMALLINT     NOT NULL DEFAULT 0,
    "is_top"        SMALLINT     NOT NULL DEFAULT 0,
    "is_recommend"  SMALLINT     NOT NULL DEFAULT 0,
    "views"         BIGINT       NOT NULL DEFAULT 0,
    "published_at"  TIMESTAMP    NULL DEFAULT NULL,
    "created_by"    BIGINT       NOT NULL DEFAULT 0,
    "updated_by"    BIGINT       NOT NULL DEFAULT 0,
    "created_at"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"    TIMESTAMP    NULL DEFAULT NULL
);

COMMENT ON TABLE plugin_cms_article IS 'CMS article content';
COMMENT ON COLUMN plugin_cms_article."id" IS 'Article ID';
COMMENT ON COLUMN plugin_cms_article."category_id" IS 'Category ID';
COMMENT ON COLUMN plugin_cms_article."title" IS 'Article title';
COMMENT ON COLUMN plugin_cms_article."subtitle" IS 'Article subtitle';
COMMENT ON COLUMN plugin_cms_article."slug" IS 'Public URL slug';
COMMENT ON COLUMN plugin_cms_article."summary" IS 'Article summary';
COMMENT ON COLUMN plugin_cms_article."cover" IS 'Cover image URL';
COMMENT ON COLUMN plugin_cms_article."author" IS 'Author name';
COMMENT ON COLUMN plugin_cms_article."source" IS 'Content source';
COMMENT ON COLUMN plugin_cms_article."content" IS 'Article body HTML';
COMMENT ON COLUMN plugin_cms_article."tags" IS 'Comma-separated tag names';
COMMENT ON COLUMN plugin_cms_article."keywords" IS 'SEO keywords';
COMMENT ON COLUMN plugin_cms_article."description" IS 'SEO description';
COMMENT ON COLUMN plugin_cms_article."sort" IS 'Display order';
COMMENT ON COLUMN plugin_cms_article."status" IS 'Status: 0=draft, 1=published';
COMMENT ON COLUMN plugin_cms_article."is_top" IS 'Top flag: 0=no, 1=yes';
COMMENT ON COLUMN plugin_cms_article."is_recommend" IS 'Recommend flag: 0=no, 1=yes';
COMMENT ON COLUMN plugin_cms_article."views" IS 'View count';
COMMENT ON COLUMN plugin_cms_article."published_at" IS 'Publication time';
COMMENT ON COLUMN plugin_cms_article."created_by" IS 'Creator user ID';
COMMENT ON COLUMN plugin_cms_article."updated_by" IS 'Updater user ID';
COMMENT ON COLUMN plugin_cms_article."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_cms_article."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_cms_article."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_cms_article_slug ON plugin_cms_article ("slug");
CREATE INDEX IF NOT EXISTS idx_plugin_cms_article_category ON plugin_cms_article ("category_id");
CREATE INDEX IF NOT EXISTS idx_plugin_cms_article_status_publish ON plugin_cms_article ("status", "published_at");
CREATE INDEX IF NOT EXISTS idx_plugin_cms_article_sort ON plugin_cms_article ("sort");
CREATE INDEX IF NOT EXISTS idx_plugin_cms_article_title ON plugin_cms_article ("title");

CREATE TABLE IF NOT EXISTS plugin_cms_article_tag (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "name"       VARCHAR(64)  NOT NULL DEFAULT '',
    "slug"       VARCHAR(64)  NOT NULL DEFAULT '',
    "sort"       INT          NOT NULL DEFAULT 0,
    "status"     SMALLINT     NOT NULL DEFAULT 1,
    "created_by" BIGINT       NOT NULL DEFAULT 0,
    "updated_by" BIGINT       NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP    NULL DEFAULT NULL
);

COMMENT ON TABLE plugin_cms_article_tag IS 'CMS article tag';
COMMENT ON COLUMN plugin_cms_article_tag."id" IS 'Tag ID';
COMMENT ON COLUMN plugin_cms_article_tag."name" IS 'Tag name';
COMMENT ON COLUMN plugin_cms_article_tag."slug" IS 'Tag slug';
COMMENT ON COLUMN plugin_cms_article_tag."sort" IS 'Display order';
COMMENT ON COLUMN plugin_cms_article_tag."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN plugin_cms_article_tag."created_by" IS 'Creator user ID';
COMMENT ON COLUMN plugin_cms_article_tag."updated_by" IS 'Updater user ID';
COMMENT ON COLUMN plugin_cms_article_tag."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_cms_article_tag."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_cms_article_tag."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_cms_article_tag_slug ON plugin_cms_article_tag ("slug");
CREATE INDEX IF NOT EXISTS idx_plugin_cms_article_tag_status_sort ON plugin_cms_article_tag ("status", "sort");

CREATE TABLE IF NOT EXISTS plugin_cms_link (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "group_code" VARCHAR(32)  NOT NULL DEFAULT '',
    "name"       VARCHAR(128) NOT NULL DEFAULT '',
    "url"        VARCHAR(500) NOT NULL DEFAULT '',
    "logo"       VARCHAR(500) NOT NULL DEFAULT '',
    "sort"       INT          NOT NULL DEFAULT 0,
    "status"     SMALLINT     NOT NULL DEFAULT 1,
    "created_by" BIGINT       NOT NULL DEFAULT 0,
    "updated_by" BIGINT       NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP    NULL DEFAULT NULL
);

COMMENT ON TABLE plugin_cms_link IS 'CMS friendly link';
COMMENT ON COLUMN plugin_cms_link."id" IS 'Link ID';
COMMENT ON COLUMN plugin_cms_link."group_code" IS 'Display group code';
COMMENT ON COLUMN plugin_cms_link."name" IS 'Link name';
COMMENT ON COLUMN plugin_cms_link."url" IS 'Link URL';
COMMENT ON COLUMN plugin_cms_link."logo" IS 'Logo URL';
COMMENT ON COLUMN plugin_cms_link."sort" IS 'Display order';
COMMENT ON COLUMN plugin_cms_link."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN plugin_cms_link."created_by" IS 'Creator user ID';
COMMENT ON COLUMN plugin_cms_link."updated_by" IS 'Updater user ID';
COMMENT ON COLUMN plugin_cms_link."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_cms_link."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_cms_link."deleted_at" IS 'Deletion time';

CREATE INDEX IF NOT EXISTS idx_plugin_cms_link_group_status_sort ON plugin_cms_link ("group_code", "status", "sort");
CREATE INDEX IF NOT EXISTS idx_plugin_cms_link_status_sort ON plugin_cms_link ("status", "sort");

CREATE TABLE IF NOT EXISTS plugin_cms_slide (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "group_code" VARCHAR(32)  NOT NULL DEFAULT '',
    "title"      VARCHAR(128) NOT NULL DEFAULT '',
    "subtitle"   VARCHAR(255) NOT NULL DEFAULT '',
    "image"      VARCHAR(500) NOT NULL DEFAULT '',
    "link"       VARCHAR(500) NOT NULL DEFAULT '',
    "sort"       INT          NOT NULL DEFAULT 0,
    "status"     SMALLINT     NOT NULL DEFAULT 1,
    "created_by" BIGINT       NOT NULL DEFAULT 0,
    "updated_by" BIGINT       NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP    NULL DEFAULT NULL
);

COMMENT ON TABLE plugin_cms_slide IS 'CMS slide';
COMMENT ON COLUMN plugin_cms_slide."id" IS 'Slide ID';
COMMENT ON COLUMN plugin_cms_slide."group_code" IS 'Display group code';
COMMENT ON COLUMN plugin_cms_slide."title" IS 'Slide title';
COMMENT ON COLUMN plugin_cms_slide."subtitle" IS 'Slide subtitle';
COMMENT ON COLUMN plugin_cms_slide."image" IS 'Slide image URL';
COMMENT ON COLUMN plugin_cms_slide."link" IS 'Click target URL';
COMMENT ON COLUMN plugin_cms_slide."sort" IS 'Display order';
COMMENT ON COLUMN plugin_cms_slide."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN plugin_cms_slide."created_by" IS 'Creator user ID';
COMMENT ON COLUMN plugin_cms_slide."updated_by" IS 'Updater user ID';
COMMENT ON COLUMN plugin_cms_slide."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_cms_slide."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_cms_slide."deleted_at" IS 'Deletion time';

CREATE INDEX IF NOT EXISTS idx_plugin_cms_slide_group_status_sort ON plugin_cms_slide ("group_code", "status", "sort");
CREATE INDEX IF NOT EXISTS idx_plugin_cms_slide_status_sort ON plugin_cms_slide ("status", "sort");

CREATE TABLE IF NOT EXISTS plugin_cms_message (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "name"       VARCHAR(128) NOT NULL DEFAULT '',
    "mobile"     VARCHAR(64)  NOT NULL DEFAULT '',
    "email"      VARCHAR(128) NOT NULL DEFAULT '',
    "content"    VARCHAR(1000) NOT NULL DEFAULT '',
    "reply"      VARCHAR(1000) NOT NULL DEFAULT '',
    "status"     SMALLINT     NOT NULL DEFAULT 0,
    "user_ip"    VARCHAR(64)  NOT NULL DEFAULT '',
    "user_agent" VARCHAR(500) NOT NULL DEFAULT '',
    "created_by" BIGINT      NOT NULL DEFAULT 0,
    "updated_by" BIGINT      NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP   NULL DEFAULT NULL
);

COMMENT ON TABLE plugin_cms_message IS 'CMS visitor message';
COMMENT ON COLUMN plugin_cms_message."id" IS 'Message ID';
COMMENT ON COLUMN plugin_cms_message."name" IS 'Visitor name';
COMMENT ON COLUMN plugin_cms_message."mobile" IS 'Visitor mobile';
COMMENT ON COLUMN plugin_cms_message."email" IS 'Visitor email';
COMMENT ON COLUMN plugin_cms_message."content" IS 'Message content';
COMMENT ON COLUMN plugin_cms_message."reply" IS 'Reply content';
COMMENT ON COLUMN plugin_cms_message."status" IS 'Status: 0=pending, 1=approved, 2=rejected';
COMMENT ON COLUMN plugin_cms_message."user_ip" IS 'Visitor IP';
COMMENT ON COLUMN plugin_cms_message."user_agent" IS 'Visitor user agent';
COMMENT ON COLUMN plugin_cms_message."created_by" IS 'Creator user ID';
COMMENT ON COLUMN plugin_cms_message."updated_by" IS 'Updater user ID';
COMMENT ON COLUMN plugin_cms_message."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_cms_message."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_cms_message."deleted_at" IS 'Deletion time';

CREATE INDEX IF NOT EXISTS idx_plugin_cms_message_status ON plugin_cms_message ("status");

INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('CMS 栏目类型', 'cms_category_type', 1, 1, 'CMS category type options', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('CMS 内容状态', 'cms_article_status', 1, 1, 'CMS article status options', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('CMS 留言状态', 'cms_message_status', 1, 1, 'CMS message status options', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('CMS 显示状态', 'cms_status', 1, 1, 'CMS enabled status options', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('CMS 推荐标识', 'cms_yes_no', 1, 1, 'CMS boolean flag options', NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_category_type', '列表栏目', '1', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_category_type', '单页栏目', '2', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_category_type', '外链栏目', '3', 3, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_article_status', '草稿', '0', 1, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_article_status', '已发布', '1', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_message_status', '待处理', '0', 1, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_message_status', '已通过', '1', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_message_status', '已拒绝', '2', 3, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_status', '启用', '1', 1, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_status', '停用', '0', 2, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_yes_no', '否', '0', 1, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cms_yes_no', '是', '1', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO plugin_cms_site ("site_key", "name", "slogan", "keywords", "description", "weixin", "status", "created_at", "updated_at")
SELECT 'default', 'LinaPro CMS', 'AI-native full-stack delivery framework', 'LinaPro,CMS,AI-native', 'LinaPro CMS demo site', '', 1, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM plugin_cms_site WHERE "site_key" = 'default'
);
