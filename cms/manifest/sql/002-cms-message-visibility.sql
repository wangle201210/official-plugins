-- 002: CMS public message visibility setting
-- 002：CMS 公开留言展示配置

ALTER TABLE plugin_cms_site
    ADD COLUMN IF NOT EXISTS "show_messages" SMALLINT NOT NULL DEFAULT 1;

ALTER TABLE plugin_cms_site
    ALTER COLUMN "show_messages" SET DEFAULT 1;

COMMENT ON COLUMN plugin_cms_site."show_messages" IS 'Show approved visitor messages on public message page: 0=no, 1=yes';
