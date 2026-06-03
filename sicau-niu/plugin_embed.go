package pluginsicauniu

import "embed"

// EmbeddedFiles contains the plugin manifest, convention-based SQL assets, and
// frontend source resources that the host packs when compiling the sicau-niu
// source plugin.
//
//go:embed plugin.yaml frontend manifest
var EmbeddedFiles embed.FS
