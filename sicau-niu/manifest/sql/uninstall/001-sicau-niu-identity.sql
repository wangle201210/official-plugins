-- ------------------------------------------------------------
-- 001 sicau-niu identity uninstall SQL file
-- 001 sicau-niu 身份能力卸载 SQL 文件
-- Purpose: Drop player and college tables on uninstall when the operator purges storage data.
-- 用途：卸载并清除存储数据时删除玩家与院系表。
-- ------------------------------------------------------------

DROP TABLE IF EXISTS plugin_sicau_niu_user;
DROP TABLE IF EXISTS plugin_sicau_niu_college;
