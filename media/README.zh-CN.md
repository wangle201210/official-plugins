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

`mediaopen`和 HotGo 兼容接口使用类 HotGo 的内部 API Key 鉴权。请求必须携带`X-Inner-Api-Key`，插件会与`media`插件运行配置中的`innerapi.apiKey`比对。未配置`innerapi.apiKey`时，默认值与 HotGo 保持一致为`media`；当该配置显式为空时，为兼容旧部署会关闭 Key 校验。

基于 token 的媒体鉴权会调用上游 Tieta OpenAPI 服务。启用 Tieta token 接口前，需要在`media`插件运行配置中设置`tieta.baseUrl`。`tieta.timeout`缺失或非法时默认使用`3s`，`tieta.mock: true`仅用于本地开发的确定性响应。

媒体租户标识取自铁塔的`customerCode`，并去除首尾空白。既有`tenantId`接口字段和`tenant_id`数据库列保留名称及字符串类型。租户策略、租户设备策略、租户白名单、节点并发限制、内部策略请求以及流和会话上报都应填写同一客户编码。客户编码缺失时拒绝租户相关鉴权，不使用`customerId`兜底；缓存用户身份也按同一规则从`CustomerCode`派生`TenantId`。上游设备权限请求仍使用 token 和设备 ID。

切换已有数据时，应根据核实的`customerId`到`customerCode`映射更新租户标识，并协调上报客户端。采集端按原值保存上报的租户标识，无法推断该映射。本次代码变更不调整表结构，也不自动改写既有数据库记录。

路由记忆复用宿主 `pluginhost.HostServices.Cache()` 服务，默认保留 12 小时，逻辑键格式与 HotGo 保持一致：`route_data:<deviceCode>:<channelCode>`。

插件不定义专属 Redis 配置命名空间，也不维护插件自有缓存后端。

## 开发入口

- 后端入口：`backend/plugin.go`
- API DTO：`backend/api/media/v1/`
- 业务服务：`backend/internal/service/media/`
- 前端页面：`frontend/pages/media-management.vue`
- PostgreSQL 安装 SQL：`manifest/sql/001-media-schema.sql`
- 演示案例数据：`manifest/sql/mock-data/001-media-mock-data.sql`
