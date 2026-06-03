# Manifest 资源

`sicau-niu`插件自有的生命周期资源。

## 内容

- `config/config.example.yaml` —— 插件配置示例：微信网关（`wechat.appId/secret/mock/mockOpenid`）与玩家 token（`token.secret/ttl`）。
- `sql/001-sicau-niu-identity.sql` —— 安装 DDL，创建玩家表（`plugin_sicau_niu_user`）与院系字典表（`plugin_sicau_niu_college`）。幂等（PostgreSQL）。
- `sql/uninstall/001-sicau-niu-identity.sql` —— 卸载并清除存储数据时删除玩家表与院系表。

无`i18n/`资源：插件为单语言。Mock 数据 SQL 在后续迭代按需添加。
