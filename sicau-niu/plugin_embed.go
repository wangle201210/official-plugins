package pluginsicauniu

import "embed"

// EmbeddedFiles contains the plugin manifest and frontend source resources that
// the host packs when compiling the sicau-niu source plugin. This sample owns no
// SQL or i18n assets, so only plugin.yaml and the frontend directory are embedded.
//
//go:embed plugin.yaml frontend
var EmbeddedFiles embed.FS
