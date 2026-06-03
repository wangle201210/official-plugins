# sicau-niu

`sicau-niu` is the SICAU 120th-anniversary "find-the-cow" (寻牛) activity plugin for
LinaPro. It is delivered as a `platform_only` source plugin (the activity is a single
deployment, not multi-tenant) and is single-language (Chinese, no `i18n`).

This repository delivers the activity **backend API**, the **operator console pages**,
and the **H5 memorial wall**. The WeChat mini-program app is a separate delivery; this
plugin owns its backend API contract.

## Iterations

The activity is built as a sequence of OpenSpec changes (`C1`–`C7`, see
`openspec/changes/`). This directory currently implements:

### C1 `niu-identity` — player identity & college dictionary

| Aspect | Value |
|--------|-------|
| Type | `source` |
| Scope | `platform_only` |
| Install mode | `global` |
| `i18n` | Not enabled (single-language) |

Capabilities:

- **Player identity** — WeChat login (`code` → openid, first-login provisioning), phone
  authorization binding under a one-phone-one-account constraint, identity profile
  (identity tag / college / grade / graduation year), all isolated from host admin users.
- **College dictionary** — operator-maintained college list used by player profile
  selection and future college-ranking aggregation.
- Operator read-only player query.

### Auth model

| Surface | Guard |
|---------|-------|
| `POST /{api}/api/v1/plugins/sicau-niu/player/login` | Public |
| `/{api}/api/v1/plugins/sicau-niu/player/*`, `/colleges` | Plugin-owned player token middleware |
| `/{api}/api/v1/plugins/sicau-niu/admin/*` | Host `Auth + Tenancy + Permission` (`sicau-niu:*`) |

Player tokens are plugin self-signed (stdlib HMAC-SHA256); WeChat access uses a
mockable gateway seam (`wechat.mock`) so the flow runs without live credentials.

## Layout

```text
sicau-niu/
  backend/
    api/{player,admin}/     API DTOs and route contracts
    internal/
      controller/{player,admin}/  Request handling and response projection
      service/{identity,college,token,wechat}/  Business logic, token, WeChat seam
      middleware/            Plugin player-auth middleware
      dao|model/             Generated DAO/DO/Entity (do not edit)
    hack/config.yaml         Plugin-local DAO codegen config
    plugin.go                Source-plugin registration and route binding
  frontend/pages/            Operator console pages (college dictionary, player query)
  manifest/
    config/                  WeChat / token configuration example
    sql/                     Install DDL (player + college tables)
    sql/uninstall/           Uninstall cleanup
  plugin.yaml                Plugin manifest
  plugin_embed.go            Embedded asset registration entry
```

## Code Generation

```bash
make -C apps/lina-plugins dao  p=sicau-niu   # regenerate DAO/DO/Entity from DB schema
make -C apps/lina-plugins ctrl p=sicau-niu   # regenerate controller scaffolding from api DTOs
```
