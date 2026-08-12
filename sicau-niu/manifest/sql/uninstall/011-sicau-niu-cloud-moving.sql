-- 011 sicau-niu cloud-moving uninstall
DELETE FROM sys_dict_data
WHERE "tenant_id" = 0
  AND "dict_type" IN (
    'sicau_niu_transport_team_status',
    'sicau_niu_transport_member_role',
    'sicau_niu_transport_invalid_reason'
  );
DELETE FROM sys_dict_type
WHERE "tenant_id" = 0
  AND "type" IN (
    'sicau_niu_transport_team_status',
    'sicau_niu_transport_member_role',
    'sicau_niu_transport_invalid_reason'
  );
DROP TABLE IF EXISTS plugin_sicau_niu_transport_report;
