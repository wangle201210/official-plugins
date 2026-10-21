# media

`media` 是 LinaPro 的媒体管理源码插件，用于维护媒体策略、策略绑定、流别名、节点配置、设备节点、租户流配置和租户白名单。

本模块按用户要求使用中文-only 文案，不提供运行时 i18n JSON 或 apidoc i18n JSON。

## 能力范围

- 媒体策略增删查改与全局策略设置
- 设备、租户、租户设备三类策略绑定
- 按优先级预览设备当前生效策略
- 流别名增删查改
- 节点、设备节点与租户流配置增删查改
- 租户白名单增删查改
- `mediaopen`内部策略解析接口：`GET /api/v1/strategies/resolve?tenantId=<tenantId>&deviceId=<deviceId>`
- `mediaopen`策略鉴权接口：`GET /api/v1/strategies/user-device?token=<token>&deviceId=<deviceId>`
- `mediaopen`流别名配置接口：`GET /api/v1/stream-aliases/by-alias?alias=<alias>`
- `mediaopen`全量节点配置接口：`GET /api/v1/nodes/all`
- HotGo 兼容路由记忆接口：`PUT /api/v1/route-memories/{deviceCode}/{channelCode}`、`GET /api/v1/route-memories/{deviceCode}/{channelCode}`和`DELETE /api/v1/route-memories/{deviceCode}/{channelCode}`

## 配置说明

白名单接口`GET /api/v1/tenant-whites/ips?token=<token>`与策略接口统一使用`code`、`message`和`data`响应包装。成功时从`data.tenantId`读取客户编码，从`data.ips`读取启用的 IP 白名单；未配置白名单时返回空数组，同时保留租户 ID。客户端必须检查业务`code`，不能只依赖 HTTP 状态。

`mediaopen`和 HotGo 兼容接口使用类 HotGo 的内部 API Key 鉴权。请求必须携带`X-Inner-Api-Key`，插件会与`media`插件运行配置中的`innerapi.apiKey`比对。未配置`innerapi.apiKey`时，默认值与 HotGo 保持一致为`media`；当该配置显式为空时，为兼容旧部署会关闭 Key 校验。

基于 token 的媒体鉴权会调用上游 Tieta OpenAPI 服务。启用 Tieta token 接口前，需要在`media`插件运行配置中设置`tieta.baseUrl`。`tieta.timeout`缺失或非法时默认使用`3s`，`tieta.mock: true`仅用于本地开发的确定性响应。

媒体租户标识取自铁塔的`customerCode`，并去除首尾空白。既有`tenantId`接口字段和`tenant_id`数据库列保留名称及字符串类型。租户策略、租户设备策略、租户白名单、节点并发限制、内部策略请求以及流和会话上报都应填写同一客户编码。客户编码缺失时拒绝租户相关鉴权，不使用`customerId`兜底；缓存用户身份也按同一规则从`CustomerCode`派生`TenantId`。上游设备权限请求仍使用 token 和设备 ID。

切换已有数据时，应根据核实的`customerId`到`customerCode`映射更新租户标识，并协调上报客户端。采集端按原值保存上报的租户标识，无法推断该映射。本次代码变更不调整表结构，也不自动改写既有数据库记录。

路由记忆复用宿主 `pluginhost.HostServices.Cache()` 服务，默认保留 12 小时，逻辑键格式与 HotGo 保持一致：`route_data:<deviceCode>:<channelCode>`。

插件不定义专属 Redis 配置命名空间，也不维护插件自有缓存后端。

## 节点网络指标

`GET /api/v1/media/dashboard/nodes/overview?rootNodeId=<nodeId>`返回指定节点子树，沿用既有管理端认证和`media:management:query`权限。每个节点包含：

- `network_links`：按来源`machine_id`、`source_ip`和`destination_ip`区分的最新原始探测数据。每项包含毫秒单位的`rtt`和`jitter`、`0～1`比例的`packet_loss`、`status_code`、可空的`reachable`及`Unix`毫秒时间戳`report_time`。当前`net-flux`探测器的`packet_loss`实际表示`HTTP`探测失败比例。`extra.reachable`缺省或非法时保留为`null`，失败固定值按原值返回。
- `instance_resources`：当前节点直属实例的资源汇总，包括实例数量、分配和已用`CPU`核数、`MB`单位的内存及`KB/s`单位的磁盘和网络速率，不包含子节点。既有顶层资源字段继续表示子树汇总。新的机器上报将`CPU`使用率百分比转换为已用核数，既有样本在下一次上报时刷新。
- `network_links_limited`：仅响应根节点表示子树的原始链路明细是否超过共享的`10000`条上限，子节点为`false`。链路按所选子树通过一次数据库查询加载，资源复用已有实例分组聚合查询。

`node_latency_map`继续提供目的地址到延迟的摘要，键是上报的`destination_ip`，不是节点`ID`。接口不计算质量评分、质量分级、候选路径排名或综合负载百分比。没有链路时返回空数组，没有直属实例时返回各字段为零的资源对象。

对于上游机器上报中`node_id`为空或等于`instance_id`的情况，资源聚合通过相同`machine_id`的最新网络上报确定业务节点，明确的实例节点优先保留。多目的端不会重复累加资源；最新采样同刻指向多个节点时保持归属未知。归属先在全部节点范围内确定，再按子树展示，不受链路明细上限影响。

部署本功能前需执行当前迭代的`manifest/sql/003-add-media-data-collection-server.sql`，幂等创建最新探测表`plugin_media_report_network_link`。新链路明细在后续网络上报后产生，旧的纯延迟记录无法恢复丢包率和抖动。

## 开发入口

- 后端入口：`backend/plugin.go`
- API DTO：`backend/api/media/v1/`
- 业务服务：`backend/internal/service/media/`
- 前端页面：`frontend/pages/media-management.vue`
- PostgreSQL 安装 SQL：`manifest/sql/001-media-schema.sql`
- 演示案例数据：`manifest/sql/mock-data/001-media-mock-data.sql`
