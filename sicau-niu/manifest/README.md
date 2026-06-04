# Manifest Resources

Plugin-owned lifecycle assets for the `sicau-niu` plugin.

## Contents

- `config/config.example.yaml` — example plugin configuration: WeChat gateway
  (`wechat.appId/secret/mock/mockOpenid`) and player token (`token.secret/ttl`).
- `sql/001-sicau-niu-identity.sql` — install DDL for the player table
  (`plugin_sicau_niu_user`) and the college dictionary table
  (`plugin_sicau_niu_college`). Idempotent (PostgreSQL).
- `sql/uninstall/001-sicau-niu-identity.sql` — drops the player and college
  tables when the operator purges storage data on uninstall.

No `i18n/` resources: the plugin is single-language. Mock-data SQL is added in
later iterations if needed.
