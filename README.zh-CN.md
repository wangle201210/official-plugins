# LinaPro 插件目录

`official-plugins` 是 LinaPro 官方源码插件工作区。

挂载到主仓库 `linapro` 后，该工作区位于 `apps/lina-plugins/`。

在当前开源阶段，宿主只保留稳定的核心基础能力，例如用户管理、角色管理、菜单管理、字典管理、参数设置、文件管理、任务调度、插件治理和开发支持；非核心业务能力通过 `apps/lina-plugins/<plugin-id>/` 下的源码插件交付。

## 这里包含什么

当前目录下包含以下参考内容：

- `linapro-demo-source`：源码插件目录结构与开发方式样例
- `linapro-demo-dynamic`：动态 WASM 插件结构与生命周期样例
- 官方源码插件：通过显式接线编译进宿主的一方业务插件

## 作为 Submodule 使用

主仓库 `linapro` 会通过如下命令把本仓库挂载到 `apps/lina-plugins`：

```bash
git submodule update --init --recursive
```

在已经检出的 `linapro` 工作区里做本地 submodule 管理时，SSH 远端地址为：

```text
git@github.com:linaproai/official-plugins.git
```

## 官方源码插件列表

当前仓库内置以下一方源码插件：

- `linapro-ops-demo-guard`：演示环境只读保护
- `linapro-org-core`：部门管理、岗位管理
- `linapro-content-notice`：通知公告管理
- `linapro-monitor-online`：在线用户查询与强制下线治理
- `linapro-monitor-server`：服务监控采集、清理与查询
- `linapro-monitor-operlog`：操作日志落库与治理
- `linapro-monitor-loginlog`：登录日志落库与治理
- `cms`：CMS 站点、文章、分类、链接、轮播和留言管理
- `media`：媒体策略、策略绑定和流别名管理
- `water`：基于媒体策略的水印任务和预览处理

每个官方插件都使用统一的基础结构：

```text
apps/lina-plugins/<plugin-id>/
  backend/
    api/                插件 API DTO 与路由契约
    internal/
      controller/       插件 HTTP 控制器
      service/          插件业务服务
      dao/              插件存在数据库访问时生成的本地 DAO 工件
      model/do/         插件存在数据库访问时生成的本地 DO 工件
      model/entity/     插件存在数据库访问时生成的本地实体工件
    hack/config.yaml    插件本地 GoFrame codegen 配置
    plugin.go           插件后端注册入口
  frontend/pages/       由宿主菜单挂载的插件页面
  manifest/sql/         插件自有安装 SQL 资源
  manifest/sql/mock-data/ 插件自有可选`mock`/演示 SQL 资源
  manifest/sql/uninstall/ 插件自有卸载 SQL 资源
  hack/tests/e2e/       可选的插件自有 E2E TC 用例
  hack/tests/pages/     可选的插件自有 E2E 页面对象
  hack/tests/support/   可选的插件自有 E2E helper
  plugin.yaml           插件清单
  plugin_embed.go       嵌入资源注册入口
  README.md             英文说明
  README.zh-CN.md       中文说明
```

`backend/internal/service/` 是源码插件业务 `service` 的唯一合法目录，禁止再创建 `backend/service/`。

## 宿主与插件边界

当前源码插件方案强调通过稳定接缝解耦，而不是在宿主里散落大量 `if pluginEnabled` 判断。

- 宿主拥有稳定的一级目录骨架，例如 `dashboard`、`iam`、`setting`、`scheduler`、`extension`、`developer`。
- 插件通过自身 `plugin.yaml` 的 `parent_key` 自主选择菜单挂载点；宿主只在同步时解析该父级菜单记录，并拒绝缺失父级以避免产生孤儿菜单树。
- 官方插件的预期菜单位置由各插件自己的 `plugin.yaml` 维护；宿主不硬编码官方插件 ID，也不强制绑定特定父级目录。
- 宿主对插件发布稳定能力接缝，例如认证事件、审计事件、组织能力接口和插件生命周期 Hook。
- 插件自有的数据表、菜单、页面、Hook 和定时任务都保留在插件目录内，并通过插件生命周期完成安装与卸载。

## HTTP 路由与公开资源

源码插件的`HTTP`路由是插件后端代码注册的实现细节。不要在`plugin.yaml`中声明公开路由、门户路由、工作台`API`路由或路由分组。

源码插件`API`应使用`registrar.Routes().APIPrefix()`。该方法返回`/x/{plugin-id}`；这个插件自有命名空间是强制边界，其后的路径段由插件自行定义。插件可以在该前缀下选择`/api/v1`、`/api/v2`或`/interface/m1`等习惯性路径。源码插件仍可注册非保留公开路由，例如`/`、`/portal/*`或`/assets/*`，用于页面、门户、静态资源或插件自管的 fallback handler。

`plugin.yaml`的`menus`仍然是管理工作台导航与权限的事实来源。注册`HTTP`路由不会自动创建菜单、权限节点、`OpenAPI`条目或工作台路由元数据。

源码插件和动态插件可以通过`plugin.yaml`的`public_assets`声明公开静态资源目录。宿主只会从`/x-assets/{plugin-id}/{version}/...`提供已声明的公开资源，并把每个声明视为插件作者的显式发布边界。声明必须留在插件自有资源集合内；包含路径穿越、绝对路径、`URL`、通配符、重叠挂载或符号链接逃逸的配置都会被拒绝。

## 源码插件开发流程

1. 创建 `apps/lina-plugins/<plugin-id>/`。
2. 参考 `linapro-demo-source/` 的目录结构。
3. 在 `plugin.yaml` 中声明清单、菜单、页面、SQL 资源与可选 Hook。
4. 插件后端代码保留在插件目录中，业务逻辑统一放在 `backend/internal/service/` 下，并且只依赖宿主公开包。
5. 在 `apps/lina-plugins/lina-plugins.go` 中做显式接线。

## 插件自有 E2E 测试

源码插件应把插件专属 Playwright 覆盖放在 `apps/lina-plugins/<plugin-id>/hack/tests/e2e/` 下。
插件页面对象和辅助 helper 应分别保留在同级的 `hack/tests/pages/` 与 `hack/tests/support/` 中。
宿主测试运行器会通过通用 `plugins` 范围发现这些测试；单个插件也可以通过 `pnpm -C hack/tests test:module -- plugin:<plugin-id>` 直接运行，不需要为每个插件在执行清单里新增专属 scope。

## 源码插件版本升级

当某个源码插件已经在宿主中安装完成，而你又提升了它的 `plugin.yaml` 版本时，源码扫描不再自动替换当前生效版本。

- 当前生效版本仍然固定在 `sys_plugin.version` 与 `release_id`。
- 新发现的更高源码版本会写入一个 `prepared` 状态的 `sys_plugin_release`。
- 在宿主允许启动前，必须通过受支持的插件工作区更新流程处理源码插件。
- 如果跳过这一步，宿主启动会直接失败，并输出仍需处理的插件。

## 动态插件说明

动态 WASM 插件仍然适用于运行时托管交付场景。如果插件需要通过上传、安装、启用、停用和卸载完成生命周期管理，请参考 `linapro-demo-dynamic/`。

## 参考入口

- `apps/lina-plugins/linapro-demo-source/README.md`
- `apps/lina-plugins/linapro-demo-dynamic/README.md`
