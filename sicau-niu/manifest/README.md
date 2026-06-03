# Manifest Resources

This directory holds plugin-owned lifecycle assets (`SQL`, mock data, `i18n`) for
the `sicau-niu` source plugin.

The `sicau-niu` sample intentionally owns no lifecycle assets:

- No `sql/` install assets — the plugin keeps no database table.
- No `sql/mock-data/` or `sql/uninstall/` assets.
- No `i18n/` resources — the plugin is single-language.

Menus stay in `plugin.yaml`; this directory is reserved for future plugin-owned
data lifecycle changes.
