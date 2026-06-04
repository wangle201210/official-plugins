# CMS Manifest

This directory contains installable CMS plugin resources:

- `sql/`: PostgreSQL schema and dictionaries loaded during normal plugin installation.
- `sql/mock-data/`: optional reference-site demo reset data used for local validation and runtime sample loading.
- `sql/uninstall/`: uninstall SQL for plugin-owned tables.
- `i18n/`: runtime translation resources for plugin menus, dictionaries, errors, API docs, and frontend text.

`sql/mock-data/002-cms-starter-content.sql` is the optional starter dataset. It
gives local validation and the management sample loader a populated public site
with categories, slides, links, rich article bodies, and visible approved
visitor messages without coupling normal plugin installation to demo content.

The CMS management `Clear Data` action removes user-facing CMS content and
resets the default site record for a fresh build. It is implemented by the
plugin service rather than SQL migration files because it is a runtime
administrator operation. Host-owned uploaded files are intentionally left
untouched.

The CMS management `Load Sample Data` action first clears the same CMS-owned
business data and then replays `sql/mock-data/002-cms-starter-content.sql` from
embedded plugin resources. It does not depend on external filesystem paths after
the plugin has been built.

All SQL files are plugin-owned and must remain idempotent.
