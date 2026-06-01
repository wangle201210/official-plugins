// This file embeds the backend-only source plugin manifest and manifest resources.

package uidentitycas

import "embed"

// EmbeddedFiles contains the plugin manifest and manifest resources.
//
//go:embed plugin.yaml manifest
var EmbeddedFiles embed.FS
