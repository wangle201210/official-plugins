# sicau-niu

`sicau-niu` is a minimal source-plugin example for LinaPro. It demonstrates the
smallest viable source plugin: one left-sidebar menu page, one public route, and
one permission-protected route, with no database, `SQL`, or `i18n` assets.

## Capability

| Aspect | Value |
|--------|-------|
| Type | `source` |
| Scope | `tenant_aware` |
| Install mode | `tenant_scoped` |
| Database | None (returns static in-memory sample data) |
| `i18n` | Not enabled (single-language sample) |

## Routes

| Method | Path | Auth | Permission | Purpose |
|--------|------|------|------------|---------|
| `GET` | `/portal/sicau-niu/ping` | Public | None | Plain-text public route smoke check |
| `GET` | `/{api}/api/v1/plugins/sicau-niu/ping` | Public | None | `JSON` public ping for route registration demo |
| `GET` | `/{api}/api/v1/plugins/sicau-niu/cattle` | Required | `sicau-niu:example:view` | Returns a bounded static list of sample cattle records |

The protected list returns a fixed, small in-memory dataset, so it triggers no
database access and no `N+1` query path.

## Layout

```text
sicau-niu/
  backend/
    api/niu/                API DTOs and route contracts
    internal/controller/niu Request handling and response projection
    internal/service/niu    Static sample data orchestration
    plugin.go               Source-plugin registration and route binding
  frontend/pages/           Plugin-owned menu page and API client
  manifest/                 Lifecycle asset directory (no assets in this sample)
  plugin.yaml               Plugin manifest
  plugin_embed.go           Embedded asset registration entry
```

## Code Generation

```bash
make -C apps/lina-plugins ctrl p=sicau-niu
```

The command regenerates `backend/api/niu/niu.go` and controller scaffolding from
the `backend/api/niu/v1` DTOs.
