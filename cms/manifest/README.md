# CMS Manifest

This directory contains installable CMS plugin resources:

- `sql/`: PostgreSQL schema and required seed data.
- `sql/mock-data/`: reference-site demo data used for local validation.
- `sql/uninstall/`: uninstall SQL for plugin-owned tables.
- `i18n/`: runtime translation resources for plugin menus, dictionaries, errors, API docs, and frontend text.

All SQL files are plugin-owned and must remain idempotent.
