# CMS Manifest

该目录保存 CMS 插件的可安装资源：

- `sql/`：PostgreSQL 数据结构和必要初始化数据。
- `sql/mock-data/`：本地校验使用的参考站点演示数据。
- `sql/uninstall/`：插件自有数据表卸载 SQL。
- `i18n/`：插件菜单、字典、错误、接口文档和前端文案的运行时翻译资源。

所有 SQL 文件均属于插件自身资源，并且必须保持幂等。
