// This file embeds the CMS plugin manifest, management pages, public site assets, and lifecycle resources.

package cms

import "embed"

// EmbeddedFiles contains the CMS plugin manifest, management pages, public site assets, and manifest resources.
//
//go:embed plugin.yaml frontend public manifest
var EmbeddedFiles embed.FS
