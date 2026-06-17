// This file verifies embedded media plugin manifest metadata that gates source
// plugin runtime upgrades.

package media

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// TestEmbeddedManifestVersionCoversCollectionServerSQL verifies the manifest
// version is high enough to trigger the data-collection SQL migration.
func TestEmbeddedManifestVersionCoversCollectionServerSQL(t *testing.T) {
	const expectedVersion = "v0.1.2"

	content, err := EmbeddedFiles.ReadFile("plugin.yaml")
	if err != nil {
		t.Fatalf("read embedded plugin.yaml: %v", err)
	}
	var manifest struct {
		Version string `yaml:"version"`
	}
	if err = yaml.Unmarshal(content, &manifest); err != nil {
		t.Fatalf("parse embedded plugin.yaml: %v", err)
	}
	if manifest.Version != expectedVersion {
		t.Fatalf("expected media plugin version %s for 003 collection SQL, got %s", expectedVersion, manifest.Version)
	}
	if _, err = EmbeddedFiles.ReadFile("manifest/sql/003-add-media-data-collection-server.sql"); err != nil {
		t.Fatalf("expected 003 collection SQL to be embedded with version %s: %v", expectedVersion, err)
	}
}
