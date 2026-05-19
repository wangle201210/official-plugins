-- 001: linapro-content-notice schema uninstall
-- 001：linapro-content-notice 数据结构卸载

DELETE FROM sys_dict_data WHERE "dict_type" IN ('sys_notice_type', 'sys_notice_status');
DELETE FROM sys_dict_type WHERE "type" IN ('sys_notice_type', 'sys_notice_status');
DROP TABLE IF EXISTS plugin_linapro_content_notice;
