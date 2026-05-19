-- 001: linapro-monitor-operlog schema uninstall
-- 001：linapro-monitor-operlog 数据结构卸载

DELETE FROM sys_dict_data WHERE "dict_type" IN ('sys_oper_type', 'sys_oper_status');
DELETE FROM sys_dict_type WHERE "type" IN ('sys_oper_type', 'sys_oper_status');
DROP TABLE IF EXISTS plugin_linapro_monitor_operlog;
