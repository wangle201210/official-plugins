# Manifest Resources

This directory stores the install and uninstall SQL assets for `linapro-org-core`.

## Contents

- `sql/001-linapro-org-core-schema.sql`: creates department, post, user-dept, and user-post tables
- `sql/uninstall/001-linapro-org-core-schema.sql`: drops the plugin-owned organization tables during uninstall purge
