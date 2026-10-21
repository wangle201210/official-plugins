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

The whitelist endpoint `GET /api/v1/tenant-whites/ips?token=<token>` uses the same `code`/`message`/`data` response envelope as the strategy endpoints. On success, read the customer code from `data.tenantId` and the enabled IP list from `data.ips`; an unconfigured whitelist returns an empty array while preserving the tenant ID. Clients must check the business `code` instead of relying only on HTTP status.

Mediaopen and HotGo-compatible APIs use a HotGo-style inner API key gate. Requests must send `X-Inner-Api-Key`, and the plugin compares it with `innerapi.apiKey` from the `media` plugin runtime config. When `innerapi.apiKey` is absent, the HotGo-compatible default is `media`; when it is explicitly blank, the key check is disabled for compatibility.

Token-based media authorization calls the upstream Tieta OpenAPI service. Configure `tieta.baseUrl` in the `media` plugin runtime config before enabling Tieta-backed token APIs. `tieta.timeout` defaults to `3s` when omitted or invalid, and `tieta.mock: true` is only for deterministic local development.

Media tenant keys come from Tieta's `customerCode`, with surrounding whitespace removed. Existing `tenantId` API fields and `tenant_id` columns keep their names and string types. Use the same customer code in tenant and tenant-device strategy bindings, tenant whitelists, node concurrency limits, internal strategy requests, and stream/session reports. A missing customer code rejects tenant-scoped authorization; `customerId` is never a fallback. Cached user identities derive `TenantId` from `CustomerCode` using the same rule. The upstream device-permission request continues to use the token and device ID.

When switching existing data, update tenant keys using a verified `customerId`-to-`customerCode` mapping and coordinate the reporting clients. The collector stores reported tenant keys as supplied; it cannot infer that mapping. This code change does not alter table structure or rewrite existing database rows.

Route memory reuses the host `pluginhost.HostServices.Cache()` service and keeps entries for 12 hours with HotGo-compatible logical keys in the form `route_data:<deviceCode>:<channelCode>`.

The plugin does not define a plugin-specific Redis configuration namespace or maintain a plugin-owned cache backend.

## Node network metrics

`GET /api/v1/media/dashboard/nodes/overview?rootNodeId=<nodeId>` returns the selected node subtree and requires the existing management authentication and `media:management:query` permission. Each node includes:

- `network_links`: latest raw probes, distinguished by source `machine_id`, `source_ip`, and `destination_ip`. Each item includes `rtt` and `jitter` in milliseconds, `packet_loss` as a `0–1` fraction, `status_code`, nullable `reachable`, and `report_time` as a Unix millisecond timestamp. The current `net-flux` prober reports HTTP probe failure ratio as `packet_loss`. Missing or invalid `extra.reachable` remains `null`; failure sentinel values remain unchanged.
- `instance_resources`: totals for instances attached directly to this node, including instance count, allocated/used CPU cores, memory in MB, and disk/network rates in KB/s. Child-node resources are excluded. Existing top-level resource fields continue to represent subtree totals. CPU usage percentages are converted to used cores on new machine reports; existing samples refresh with the next report.
- `network_links_limited`: on the response root, indicates that the subtree's raw link details exceeded the shared 10,000-item limit. Child nodes report `false`. A single database query loads the selected subtree's links; resource aggregation reuses existing grouped instance queries.

`node_latency_map` remains a destination-address-to-delay summary; its keys are the reported `destination_ip`, not node IDs. No quality score, quality grade, candidate-path ranking, or combined load percentage is calculated. Empty link results are arrays, and absent direct instance resources are zero-valued objects.

For upstream machine reports whose `node_id` is empty or equals `instance_id`, resource aggregation resolves the business node from the latest network probes with matching `machine_id`. Explicit instance node IDs take priority. Multiple destinations do not duplicate resources; conflicting nodes at the same latest timestamp leave ownership unknown. Ownership is resolved across all nodes before subtree selection and is independent of the link-detail limit.

Before deploying this feature, apply the current iteration's `manifest/sql/003-add-media-data-collection-server.sql`, which creates the idempotent latest-probe table `plugin_media_report_network_link`. The new link details become available after subsequent network reports; old latency-only records cannot reconstruct packet loss or jitter.

## Development

- Backend entry: `backend/plugin.go`
- API DTOs: `backend/api/media/v1/`
- Service implementation: `backend/internal/service/media/`
- Frontend page: `frontend/pages/media-management.vue`
- PostgreSQL install SQL: `manifest/sql/001-media-schema.sql`
- Mock demo data: `manifest/sql/mock-data/001-media-mock-data.sql`
