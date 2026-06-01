-- 001: linapro-uidentity-cas backend schema uninstall
-- Purpose: Drops plugin-owned identity, CAS, OAuth, password policy, blacklist, and audit tables.

DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_account_change_log;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_job_log;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_sys_job;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_oauth_token;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_oauth_log;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_cas_login_log;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_sms;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_pass_rule;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_group_app_blacklist;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_account_app_blacklist;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_account_app_role;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_account_unit;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_account_group;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_application;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_container;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_unit;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_group;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_account_detail;
DROP TABLE IF EXISTS plugin_linapro_uidentity_cas_account;
