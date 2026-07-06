# media

`media` is a LinaPro source plugin for media strategy, strategy binding, stream alias, node config, device-node, tenant stream config, and tenant whitelist management.

This module is intentionally Chinese-only for user-facing copy. It does not ship runtime i18n JSON or apidoc i18n JSON.

## Capabilities

- Media strategy CRUD and global strategy selection
- Device, tenant, and tenant-device strategy bindings
- Effective strategy preview with priority resolution
- Stream alias CRUD
- Node, device-node, and tenant stream config CRUD
- Tenant whitelist CRUD
- Mediaopen internal strategy resolution API: `GET /api/v1/strategies/resolve?tenantId=<tenantId>&deviceId=<deviceId>`
- Mediaopen strategy authorization API: `GET /api/v1/strategies/user-device?token=<token>&deviceId=<deviceId>`
- Mediaopen stream alias config API: `GET /api/v1/stream-aliases/by-alias?alias=<alias>`
- Mediaopen full node config API: `GET /api/v1/nodes/all`
- HotGo-compatible route memory APIs: `PUT /api/v1/route-memories/{deviceCode}/{channelCode}`, `GET /api/v1/route-memories/{deviceCode}/{channelCode}`, and `DELETE /api/v1/route-memories/{deviceCode}/{channelCode}`

## Configuration

Mediaopen and HotGo-compatible APIs use a HotGo-style inner API key gate. Requests must send `X-Inner-Api-Key`, and the plugin compares it with `innerapi.apiKey` from the `media` plugin runtime config. When `innerapi.apiKey` is absent, the HotGo-compatible default is `media`; when it is explicitly blank, the key check is disabled for compatibility.

Token-based media authorization calls the upstream Tieta OpenAPI service. Configure `tieta.baseUrl` in the `media` plugin runtime config before enabling Tieta-backed token APIs. `tieta.timeout` defaults to `3s` when omitted or invalid, and `tieta.mock: true` is only for deterministic local development.

Route memory reuses the host `pluginhost.HostServices.Cache()` service and keeps entries for 12 hours with HotGo-compatible logical keys in the form `route_data:<deviceCode>:<channelCode>`.

The plugin does not define a plugin-specific Redis configuration namespace or maintain a plugin-owned cache backend.

## Development

- Backend entry: `backend/plugin.go`
- API DTOs: `backend/api/media/v1/`
- Service implementation: `backend/internal/service/media/`
- Frontend page: `frontend/pages/media-management.vue`
- PostgreSQL install SQL: `manifest/sql/001-media-schema.sql`
- Mock demo data: `manifest/sql/mock-data/001-media-mock-data.sql`
