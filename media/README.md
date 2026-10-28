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

## Complete node configuration queries

All node-management configurations are available through the internal API:

| Documentation group | API | Authentication |
| --- | --- | --- |
| 内部接口 | `GET /api/v1/nodes/all` | Existing `X-Inner-Api-Key` rules |

The endpoint queries the `media_node` configuration table, accepts no pagination or filter parameters, and returns the complete `data.list` ordered by `nodeNum` ascending. Empty results are `[]`. The unique node-number constraint and `0..255` range bound the set to 256 records; no default page size truncates it. Each item contains `id`, `nodeNum`, `name`, `qnUrl`, `basicUrl`, `dnUrl`, `creatorId`, `createTime`, `updaterId` and `updateTime`. Times are Unix millisecond numbers, or `null` when missing.

This endpoint returns node-management configuration. Use `/api/v1/media/dashboard/nodes/overview` for reported state and resource metrics. The original `/api/v1/media/nodes` remains the paginated management list. No database migration is required.

## Node network metrics

`GET /api/v1/media/dashboard/nodes/overview?rootNodeId=<nodeId>` returns the selected node subtree and requires the existing management authentication and `media:management:query` permission. Each node includes:

- `network_links`: latest raw probes, distinguished by source `machine_id`, `source_ip`, and `destination_ip`. Each item includes `rtt` and `jitter` in milliseconds, `packet_loss` as a `0–1` fraction, `status_code`, nullable `reachable`, and `report_time` as a Unix millisecond timestamp. The current `net-flux` prober reports HTTP probe failure ratio as `packet_loss`. Missing or invalid `extra.reachable` remains `null`; failure sentinel values remain unchanged.
- `instance_resources`: totals for instances attached directly to this node, including instance count, allocated/used CPU cores, memory in MB, and disk rates in KB/s and network rates in Kbps. Child-node resources are excluded. Existing top-level resource fields continue to represent subtree totals. CPU usage percentages are converted to used cores on new machine reports; existing samples refresh with the next report.
- `network_links_limited`: on the response root, indicates that the subtree's raw link details exceeded the shared 10,000-item limit. Child nodes report `false`. A single database query loads the selected subtree's links; resource aggregation reuses existing grouped instance queries.

`node_latency_map` remains a destination-address-to-delay summary; its keys are the reported `destination_ip`, not node IDs. No quality score, quality grade, candidate-path ranking, or combined load percentage is calculated. Empty link results are arrays, and absent direct instance resources are zero-valued objects.

For upstream machine reports whose `node_id` is empty or equals `instance_id`, resource aggregation resolves the business node from the latest network probes with matching `machine_id`. Explicit instance node IDs take priority. Multiple destinations do not duplicate resources; conflicting nodes at the same latest timestamp leave ownership unknown. Ownership is resolved across all nodes before subtree selection and is independent of the link-detail limit.

Before deploying this feature, apply the current iteration's `manifest/sql/003-add-media-data-collection-server.sql`, which creates the idempotent latest-probe table `plugin_media_report_network_link`. The new link details become available after subsequent network reports; old latency-only records cannot reconstruct packet loss or jitter.

Node overview fields (including all child nodes and `instance_resources`) and `GET /api/v1/media/dashboard/instances` expose `network_in/network_out` in Kbps, where `1000 bps = 1 Kbps`. The API multiplies existing KiB/s projections by `8.192` at the response boundary. For example, a projection of `1000 KiB/s` produces `8192 Kbps`. Existing projection storage rounding remains unchanged. Raw `report_payload` values remain in bps, disk units stay unchanged, and no database migration is required.

## Complete latest reports

The collector preserves each valid object's complete latest known protobuf report in the nullable `report_payload` JSON column. PostgreSQL `JSON` is used because it preserves escaped `U+0000` in original extension strings, which `JSONB` rejects:

| Report | Table |
| --- | --- |
| `MachineMetric` | `media_report_instance` |
| `NetworkMetric` | `plugin_media_report_network_link` |
| `StreamMetric` | `media_report_stream` |
| `SessionMetric` | `media_report_session` |

Payloads retain original protobuf field names, default values, numeric enums, full `extra` maps and all nested `LinkHop` fields. They are serialized from the incoming message before business normalization changes units or combines identifiers. Protobuf JSON represents 64-bit integers as strings to preserve precision. It does not preserve unknown protobuf wire fields. Stream and session rows also store `report_subcommand`, including delete events; a first-seen delete creates a closed row, and a sparse delete does not erase existing display fields.

This stores one latest accepted report per existing business key, not an event history. Machine, stream and session reports keep the existing arrival-order replacement behavior; network links reject older sample timestamps. Different probe URLs sharing the existing link key still share one latest record. Invalid reports without required business keys are skipped; network reports without a destination do not create a full link record. Closed stream and session rows, including payloads, retain the existing 30-day cleanup policy. Lifecycle counters may update an instance row's `report_time`; the machine sample's original time remains in `report_payload.timestamp`.

Apply the appended `ALTER TABLE` statements in `manifest/sql/003-add-media-data-collection-server.sql` before starting this version. They also convert any preliminary `report_payload JSONB` columns to `JSON`. Existing rows remain `NULL` until the next report; previously discarded data cannot be backfilled. The four display queries below return the complete raw payload; unrelated summaries and topology projections continue to exclude it. Existing business-key lengths, IP formats and numeric projection constraints still apply; violations return errors and are not reported as successful persistence.

The raw payload also preserves disk I/O counters. The current upstream collector reports cumulative disk bytes, while the existing dashboard projection labels the converted values as rates; that separate display-unit issue is not corrected by raw payload storage.

## Third-party report queries

The existing protected `GET` APIs return complete latest reports as JSON objects under `report_payload`, alongside their normalized display fields:

| API path | Payload location in `data` |
| --- | --- |
| `/api/v1/media/dashboard/instances` | `instance_list[].report_payload` |
| `/api/v1/media/dashboard/nodes/overview` | Each node's `network_links[].report_payload`, including child nodes |
| `/api/v1/media/dashboard/streams` | `stream_list[].report_payload` |
| `/api/v1/media/dashboard/sessions` | `protocol_list[].active_session_list[].report_payload` |

Stream and session items also include `report_subcommand` (`0` when not recorded). A missing original report returns `null`; it is never reconstructed from display fields. The value is an object, not a JSON-encoded string. Original names, complete `extra`, nested hop fields, enum values, zero values and escaped characters are preserved. Protobuf 64-bit integers remain strings and original timestamp units remain unchanged inside the payload.

These routes retain the existing LinaPro/Tieta token authentication and management permission declaration; the mediaopen `X-Inner-Api-Key` alone does not authorize them. Existing filters and 10,000-row limits remain. The session query still requires `streamId` and returns active sessions only. Use node, tenant, device, instance and stream filters to limit response size. Raw reports are loaded with their rows, without per-item database queries. No additional SQL migration is needed beyond the complete-report storage migration above.

The media documentation page is `/api/v1/media/apidocs.html`; the OpenAPI JSON is `/api/v1/media/openapi.json`.

Each instance-list item also exposes nullable Unix millisecond timestamps `report_time` and `updated_at`. `report_time` comes from the original machine report's `timestamp`, converting seconds to milliseconds; missing, non-positive or invalid values return `null`. It does not fall back to the database `report_time`, which can change on stream/session counter events. `updated_at` is the instance row's database update time, or `null` when absent. The original `report_payload.timestamp` keeps its reported type and unit. No new database columns are required.

## Stream status and list filters

`GET /api/v1/media/dashboard/streams` returns all matching latest stream projections by default, including closed records awaiting cleanup. An omitted or blank `status` adds no status filter; supplied values match the stored string exactly after trimming whitespace. Collector status values are:

| `status` | Meaning |
| --- | --- |
| `universal` | Unspecified or unknown |
| `running` | Running |
| `inactive` | Inactive |
| `error` | Error |
| `closed` | Closed |
| `timeout` | Timed out |
| `cancelled` | Cancelled |
| `failed` | Failed |

Example: `GET /api/v1/media/dashboard/streams?status=running`. This list filter checks status only; `summary` also excludes closed streams. The list returns at most 10,000 rows ordered by report time descending and stream ID ascending, without pagination or `total`.

`GET /api/v1/media/dashboard/sessions?streamId=<streamId>` requires `streamId` and returns only unclosed sessions. `protocolType` matches the stored string exactly. Standard values are `UNIVERSAL`, `RTMP`, `RTSP`, `HLS`, `HTTP_FLV`, `WS_FLV`, `HTTPS_FLV`, `WSS_FLV` and `GB28181`; `FLV` is not expanded automatically. Filters combine with AND. All protocols share a 10,000-row detail limit; `session_count` covers the full filtered scope and may exceed the `active_session_list` length. Protocol-group `active/inactive` only describes whether the matching session count is positive, not the stream status.

`sourceType=node` also restricts the stream's source type. Use `nodeId` to retrieve streams from all source types on a node. A standalone `sourceId` filters by node ID, while `sourceType=instance` interprets it as instance ID. Response `source_type/source_id` echo the request values.

## Active totals

`GET /api/v1/media/dashboard/summary` takes no parameters and returns platform-wide alive node, active stream and session counts. It uses the existing LinaPro/Tieta management authentication chain and declares `media:management:query`. This is a platform monitoring API without tenant filtering; `X-Inner-Api-Key` alone does not authorize access.

```json
{
  "active_stream_count": 128,
  "active_session_count": 1024,
  "alive_node_count": 14
}
```

These integer fields appear inside the standard response's `data` object. Streams must have `status=running` and `close_time IS NULL`. Sessions must be unclosed and refer to those running streams through `stream_id`. Sessions belonging to non-running or closed streams, and orphan sessions, are excluded. No matches return `0`; heartbeat age does not add an expiry condition. All three counts use one database statement, are unaffected by the 10,000-row list limit, and do not load details or raw payloads. `alive_node_count` counts nodes whose server-maintained `updated_at` falls within the last three minutes, including both bounds. Node status and client `report_time` do not affect the result, so delayed samples still refresh receipt-based liveness. Each node counts once without summing `alive_nodes` or requiring running streams. Missing, expired or future update times are excluded. Machine metrics update instance rows only; direct edits to the node table also affect its update time. Apply this index when deploying the optimization:

```sql
CREATE INDEX IF NOT EXISTS idx_media_report_node_updated_at
    ON media_report_node (updated_at DESC);
```


## Development

- Backend entry: `backend/plugin.go`
- API DTOs: `backend/api/media/v1/`
- Service implementation: `backend/internal/service/media/`
- Frontend page: `frontend/pages/media-management.vue`
- PostgreSQL install SQL: `manifest/sql/001-media-schema.sql`
- Mock demo data: `manifest/sql/mock-data/001-media-mock-data.sql`
