package plugin

import (
	"testing"
	"testing/fstest"
)

// TestValidPluginID проверяет, что идентификатор плагина не выводит
// распаковку и удаление за пределы каталога плагинов.
func TestValidPluginID(t *testing.T) {
	valid := []string{"tester", "example-plugin", "my_plugin", "plugin.v2", "a1"}
	for _, id := range valid {
		if !ValidPluginID(id) {
			t.Errorf("ValidPluginID(%q) = false, want true", id)
		}
	}

	invalid := []string{
		"",
		".",
		"..",
		"../../etc",
		"../evil",
		"a/b",
		`a\b`,
		`..\..\evil`,
		"/absolute",
	}
	for _, id := range invalid {
		if ValidPluginID(id) {
			t.Errorf("ValidPluginID(%q) = true, want false", id)
		}
	}
}

// TestLoadManifestRejectsBadID проверяет, что манифест с опасным id не загружается.
func TestLoadManifestRejectsBadID(t *testing.T) {
	vfs := fstest.MapFS{
		"info.yaml": &fstest.MapFile{Data: []byte("id: \"../../etc\"\nname: \"Evil\"\n")},
	}

	if _, err := LoadManifestFromFS(vfs); err == nil {
		t.Fatal("manifest with traversal id must be rejected")
	}
}

// TestLoadManifestAcceptsNormalID проверяет, что обычные id продолжают грузиться.
func TestLoadManifestAcceptsNormalID(t *testing.T) {
	vfs := fstest.MapFS{
		"info.yaml": &fstest.MapFile{Data: []byte("id: \"example-plugin\"\nname: \"Example\"\n")},
	}

	m, err := LoadManifestFromFS(vfs)
	if err != nil {
		t.Fatalf("LoadManifestFromFS failed: %v", err)
	}
	if m.ID != "example-plugin" {
		t.Errorf("id = %q, want example-plugin", m.ID)
	}
}
