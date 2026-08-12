# sicau-niu

`sicau-niu`是 LinaPro 的川农 120 周年校庆「寻牛」活动插件。以`platform_only`源码插件形态交付（活动为单部署、不涉及多租户），单语言（中文，不启用`i18n`）。

本仓库交付活动**后端 API**、**运营后台页面**与**H5 数字纪念墙**。微信小程序 App 为独立交付线，本插件负责其后端 API 契约。

## 迭代

活动按一系列 OpenSpec 变更（`C1`–`C7`，见`openspec/changes/`）推进。本目录当前实现：

### C1 `niu-identity` —— 玩家身份与院系字典

| 维度 | 取值 |
|------|------|
| 类型 | `source` |
| 作用域 | `platform_only` |
| 安装模式 | `global` |
| `i18n` | 未启用（单语言） |

能力：

- **玩家身份** —— 微信登录（`code`换 openid、首登建档）、手机号授权绑定（一机一号）、身份资料（身份标签/院系/年级/毕业年），与宿主管理员用户隔离。
- **院系字典** —— 运营维护的院系清单，供玩家资料选择与后续院系榜聚合复用。
- 运营只读玩家查询。

### 云搬牛

- **云搬牛** —— 持久团关系、玩家唯一有效团绑定、位置上报贡献事实、每日 12 次限额、72 小时不活跃团失效、聚合统计、受保护坐标审计和仅运营端改名。

### 鉴权模型

| 接口面 | 守卫 |
|--------|------|
| `POST /{api}/api/v1/plugins/sicau-niu/player/login` | 公开 |
| `/{api}/api/v1/plugins/sicau-niu/player/*`、`/colleges` | 插件自有玩家 token 中间件 |
| `/{api}/api/v1/plugins/sicau-niu/admin/*` | 宿主`Auth + Tenancy + Permission`（`sicau-niu:*`） |

玩家 token 由插件自签（stdlib HMAC-SHA256）；微信访问通过可 Mock 的网关接缝（`wechat.mock`），无需真实凭证即可走通流程。

## 目录结构

```text
sicau-niu/
  backend/
    api/{player,admin}/     API DTO 与路由契约
    internal/
      controller/{player,admin}/  请求处理与响应投影
      service/{identity,college,irontransport,...}/  业务逻辑与外部集成
      middleware/            插件玩家鉴权中间件
      dao|model/             生成的 DAO/DO/Entity（请勿手改）
    hack/config.yaml         插件本地 DAO 代码生成配置
    plugin.go                源码插件注册与路由绑定
  frontend/pages/            运营后台页面（含云搬牛管理）
  manifest/
    config/                  微信 / token 配置示例
    sql/                     活动自有数据的幂等安装 DDL
    sql/uninstall/           卸载清理
  plugin.yaml                插件清单
  plugin_embed.go            嵌入资源注册入口
```

## 代码生成

```bash
make -C apps/lina-plugins dao  p=sicau-niu   # 基于数据库结构重新生成 DAO/DO/Entity
make -C apps/lina-plugins ctrl p=sicau-niu   # 基于 api DTO 重新生成控制器骨架
```
