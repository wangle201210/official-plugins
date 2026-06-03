# sicau-niu

`sicau-niu`是 LinaPro 的最小源码插件示例，演示一个可用源码插件的最小形态：一个左侧菜单页面、一个公开路由和一个受权限保护的路由，不包含数据库、`SQL`或`i18n`资源。

## 能力说明

| 维度 | 取值 |
|------|------|
| 类型 | `source` |
| 作用域 | `tenant_aware` |
| 安装模式 | `tenant_scoped` |
| 数据库 | 无（返回内存中的静态示例数据） |
| `i18n` | 未启用（单语言示例） |

## 路由

| 方法 | 路径 | 鉴权 | 权限 | 用途 |
|------|------|------|------|------|
| `GET` | `/portal/sicau-niu/ping` | 公开 | 无 | 纯文本公开路由连通性检查 |
| `GET` | `/{api}/api/v1/plugins/sicau-niu/ping` | 公开 | 无 | 演示路由注册的`JSON`公开 ping |
| `GET` | `/{api}/api/v1/plugins/sicau-niu/cattle` | 需要 | `sicau-niu:example:view` | 返回有界的静态示例牛只列表 |

受保护列表返回固定且数量很小的内存数据集，因此不触发任何数据库访问，也不存在`N+1`查询路径。

## 目录结构

```text
sicau-niu/
  backend/
    api/niu/                API DTO 与路由契约
    internal/controller/niu 请求处理与响应投影
    internal/service/niu    静态示例数据编排
    plugin.go               源码插件注册与路由绑定
  frontend/pages/           插件自有菜单页面与 API 客户端
  manifest/                 生命周期资源目录（本示例不包含任何资源）
  plugin.yaml               插件清单
  plugin_embed.go           嵌入资源注册入口
```

## 代码生成

```bash
make -C apps/lina-plugins ctrl p=sicau-niu
```

该命令会基于`backend/api/niu/v1`下的 DTO 重新生成`backend/api/niu/niu.go`和控制器骨架。
