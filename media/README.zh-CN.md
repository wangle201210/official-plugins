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

## 完整最新上报

采集服务收到有效对象报文后，在以下表的可空`report_payload JSON`列保存其完整最新已定义协议内容。使用 PostgreSQL `JSON`是为了保留原始扩展字符串中的转义`U+0000`，`JSONB`会拒绝该值：

| 上报 | 数据表 |
| --- | --- |
| `MachineMetric` | `media_report_instance` |
| `NetworkMetric` | `plugin_media_report_network_link` |
| `StreamMetric` | `media_report_stream` |
| `SessionMetric` | `media_report_session` |

原始内容保留`protobuf`字段名、默认值、枚举数值、完整`extra`及全部嵌套`LinkHop`字段，直接从输入报文序列化，保留单位转换和标识合并之前的数据。`protobuf JSON`使用字符串保存`64`位整数以避免精度丢失，不保留未知的`protobuf wire`字段。流和会话另存`report_subcommand`，删除事件也记录完整内容；首次收到删除事件时创建已关闭记录，稀疏删除报文不会清空已有展示字段。

本功能按现有业务键保存一份最新有效报文，不提供逐次事件历史。机器、流和会话沿用按到达顺序覆盖的行为，网络链路拒绝较旧采样时间；共享现有链路键的不同探测 URL 仍共用一份最新内容。缺少必需业务键的无效报文继续跳过，缺少目的地址的网络报文不创建完整链路记录。关闭流和会话及其原始内容仍按既有`30`天规则清理。生命周期计数可能更新实例行的`report_time`，机器原始采样时间以`report_payload.timestamp`为准。

启动新版前需执行`manifest/sql/003-add-media-data-collection-server.sql`末尾新增的`ALTER TABLE`语句，已提前创建为`JSONB`的原始内容列也会转为`JSON`。旧行在下次上报前保持`NULL`，此前丢弃的数据无法回填。下述四类展示查询返回完整原始内容，其他摘要和拓扑投影仍排除不需要的原始字段。既有业务键长度、IP格式和数值投影约束仍执行，违反约束会返回错误，不会被报告为写入成功。

原始内容也保留磁盘 I/O 累计计数。当前上游采集的是累计字节，而既有看板投影把换算值标为速率；该独立的展示单位问题不会因增加原始存储而自动修正。

## 第三方上报查询

以下既有受保护`GET`接口在展示字段之外，通过`report_payload`返回完整最新上报`JSON`对象：

| 接口路径 | `data`中的位置 |
| --- | --- |
| `/api/v1/media/dashboard/instances` | `instance_list[].report_payload` |
| `/api/v1/media/dashboard/nodes/overview` | 各节点及子节点的`network_links[].report_payload` |
| `/api/v1/media/dashboard/streams` | `stream_list[].report_payload` |
| `/api/v1/media/dashboard/sessions` | `protocol_list[].active_session_list[].report_payload` |

流和会话还返回`report_subcommand`，未记录时为`0`。旧记录没有原始报文时返回`null`，不根据展示字段伪造。该值是`JSON`对象，不是二次编码的字符串；原字段名、完整`extra`、嵌套跳点、枚举值、零值和转义字符均保留。载荷中的`protobuf`64位整数仍为字符串，时间单位保持上报原值。

接口沿用既有`LinaPro`／铁塔`token`鉴权和管理权限声明，不能仅凭`mediaopen`的`X-Inner-Api-Key`调用。原筛选和`10000`条上限不变，会话接口仍必填`streamId`且只返回活跃会话。第三方应使用节点、租户、设备、实例和流筛选控制响应大小。原始内容与对应记录同批查询，不增加逐条数据库请求。本次接口返回补全无需新增`SQL`，前提是已经执行上文完整上报存储迁移。

媒体文档页面为`/api/v1/media/apidocs.html`，`OpenAPI JSON`为`/api/v1/media/openapi.json`。

## 开发入口

- 后端入口：`backend/plugin.go`
- API DTO：`backend/api/media/v1/`
- 业务服务：`backend/internal/service/media/`
- 前端页面：`frontend/pages/media-management.vue`
- PostgreSQL 安装 SQL：`manifest/sql/001-media-schema.sql`
- 演示案例数据：`manifest/sql/mock-data/001-media-mock-data.sql`
