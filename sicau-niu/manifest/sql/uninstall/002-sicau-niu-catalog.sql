-- ------------------------------------------------------------
-- 002 sicau-niu catalog uninstall SQL file
-- 002 sicau-niu 内容资产卸载 SQL 文件
-- Purpose: Drop cattle, iron-cow, card and quote tables on uninstall when purging storage data.
-- 用途:卸载并清除存储数据时删除牛、铁牛、卡片、金句表。
-- ------------------------------------------------------------

DROP TABLE IF EXISTS plugin_sicau_niu_card;
DROP TABLE IF EXISTS plugin_sicau_niu_quote;
DROP TABLE IF EXISTS plugin_sicau_niu_iron;
DROP TABLE IF EXISTS plugin_sicau_niu_niu;
