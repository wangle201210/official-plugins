# Manifest 资源

本目录用于存放`sicau-niu`源码插件自有的生命周期资源（`SQL`、Mock 数据、`i18n`）。

`sicau-niu`示例刻意不包含任何生命周期资源：

- 无`sql/`安装资源——该插件不维护任何数据库表。
- 无`sql/mock-data/`或`sql/uninstall/`资源。
- 无`i18n/`资源——该插件为单语言插件。

菜单声明保留在`plugin.yaml`中；本目录为后续插件自有数据生命周期变更预留。
