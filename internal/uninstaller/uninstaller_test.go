package uninstaller

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tjg184/agent-smith/internal/metadata"
	"github.com/tjg184/agent-smith/pkg/paths"
)

func writeLockEntry(t *testing.T, baseDir, componentType, name, sourceURL, filesystemName string) {
	t.Helper()
	err := metadata.SaveComponentEntry(
		baseDir,
		componentType,
		name,
		sourceURL,
		"git",
		sourceURL,
		"abc123",
		"",
		metadata.ComponentEntryOptions{
			UpdatedAt:      time.Now().Format(time.RFC3339),
			FilesystemName: filesystemName,
		},
	)
	if err != nil {
		t.Fatalf("writeLockEntry: %v", err)
	}
}

func TestUninstallComponent_InvalidType(t *testing.T) {
	tempDir := t.TempDir()
	u := NewUninstaller(tempDir, nil)

	err := u.UninstallComponent("invalid", "my-skill", "")
	if err == nil {
		t.Fatal("expected error for invalid component type, got nil")
	}
}

func TestUninstallComponent_RemovesDirectory(t *testing.T) {
	tempDir := t.TempDir()

	sourceURL := "https://github.com/test/repo"
	writeLockEntry(t, tempDir, "skills", "my-skill", sourceURL, "my-skill")

	componentDir := filepath.Join(tempDir, "skills", "my-skill")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		t.Fatalf("failed to create component dir: %v", err)
	}
	dummyFile := filepath.Join(componentDir, "skill.md")
	if err := os.WriteFile(dummyFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}

	u := NewUninstaller(tempDir, nil)
	if err := u.UninstallComponent("skills", "my-skill", sourceURL); err != nil {
		t.Fatalf("UninstallComponent returned error: %v", err)
	}

	if _, err := os.Stat(componentDir); !os.IsNotExist(err) {
		t.Error("expected component directory to be removed, but it still exists")
	}
}

func TestUninstallComponent_KeepsDirectoryWhenShared(t *testing.T) {
	tempDir := t.TempDir()

	source1 := "https://github.com/test/repo1"
	source2 := "https://github.com/test/repo2"

	writeLockEntry(t, tempDir, "skills", "shared-skill", source1, "shared-skill")
	writeLockEntry(t, tempDir, "skills", "shared-skill", source2, "shared-skill")

	componentDir := filepath.Join(tempDir, "skills", "shared-skill")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		t.Fatalf("failed to create component dir: %v", err)
	}

	u := NewUninstaller(tempDir, nil)
	if err := u.UninstallComponent("skills", "shared-skill", source1); err != nil {
		t.Fatalf("UninstallComponent returned error: %v", err)
	}

	if _, err := os.Stat(componentDir); os.IsNotExist(err) {
		t.Error("expected shared component directory to be kept, but it was removed")
	}
}

func TestUninstallComponent_UpdatesLockFile(t *testing.T) {
	tempDir := t.TempDir()

	sourceURL := "https://github.com/test/repo"
	writeLockEntry(t, tempDir, "skills", "my-skill", sourceURL, "my-skill")

	componentDir := filepath.Join(tempDir, "skills", "my-skill")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		t.Fatalf("failed to create component dir: %v", err)
	}

	u := NewUninstaller(tempDir, nil)
	if err := u.UninstallComponent("skills", "my-skill", sourceURL); err != nil {
		t.Fatalf("UninstallComponent returned error: %v", err)
	}

	lockPath := paths.GetComponentLockPath(tempDir, "skills")
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		return // lock file removed entirely is fine
	}
	_, err := metadata.LoadLockFileEntry(tempDir, "skills", "my-skill")
	if err == nil {
		t.Error("expected component to be removed from lock file, but it was still found")
	}
}

func TestUninstallComponent_NotInstalled(t *testing.T) {
	tempDir := t.TempDir()

	u := NewUninstaller(tempDir, nil)
	err := u.UninstallComponent("skills", "nonexistent-skill", "")
	if err == nil {
		t.Fatal("expected error for non-installed component, got nil")
	}
}

func TestUninstallAllFromSource_Force(t *testing.T) {
	tempDir := t.TempDir()

	sourceURL := "https://github.com/test/repo"
	components := []struct {
		ctype string
		name  string
	}{
		{"skills", "skill-a"},
		{"agents", "agent-b"},
	}

	for _, c := range components {
		writeLockEntry(t, tempDir, c.ctype, c.name, sourceURL, c.name)
		dir := filepath.Join(tempDir, c.ctype, c.name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	u := NewUninstaller(tempDir, nil)
	if err := u.UninstallAllFromSource(sourceURL, true); err != nil {
		t.Fatalf("UninstallAllFromSource returned error: %v", err)
	}

	for _, c := range components {
		dir := filepath.Join(tempDir, c.ctype, c.name)
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Errorf("expected %s/%s to be removed, but it still exists", c.ctype, c.name)
		}
	}
}

func TestUninstallAllFromSource_NoComponents(t *testing.T) {
	tempDir := t.TempDir()

	u := NewUninstaller(tempDir, nil)
	err := u.UninstallAllFromSource("https://github.com/nothing/here", true)
	if err != nil {
		t.Fatalf("expected no error when no components from source, got: %v", err)
	}
}

func TestIsDirectorySharedByOtherSource_NotShared(t *testing.T) {
	tempDir := t.TempDir()

	sourceURL := "https://github.com/test/repo"
	writeLockEntry(t, tempDir, "skills", "my-skill", sourceURL, "my-skill")

	u := NewUninstaller(tempDir, nil)
	shared, err := u.isDirectorySharedByOtherSource("skills", "my-skill", "my-skill", sourceURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shared {
		t.Error("expected directory to not be shared, but isDirectorySharedByOtherSource returned true")
	}
}

func TestIsDirectorySharedByOtherSource_Shared(t *testing.T) {
	tempDir := t.TempDir()

	source1 := "https://github.com/test/repo1"
	source2 := "https://github.com/test/repo2"
	writeLockEntry(t, tempDir, "skills", "shared-skill", source1, "shared-skill")
	writeLockEntry(t, tempDir, "skills", "shared-skill", source2, "shared-skill")

	u := NewUninstaller(tempDir, nil)
	shared, err := u.isDirectorySharedByOtherSource("skills", "shared-skill", "shared-skill", source1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !shared {
		t.Error("expected directory to be shared, but isDirectorySharedByOtherSource returned false")
	}
}

func TestNormalizeURLForComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://github.com/test/repo.git", "https://github.com/test/repo"},
		{"https://github.com/test/repo/", "https://github.com/test/repo"},
		{"HTTPS://GITHUB.COM/TEST/REPO", "https://github.com/test/repo"},
		{"  https://github.com/test/repo  ", "https://github.com/test/repo"},
		{"https://github.com/test/repo.git/", "https://github.com/test/repo"},
	}

	for _, tt := range tests {
		got := normalizeURLForComparison(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeURLForComparison(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMatchesSourceURL(t *testing.T) {
	u := NewUninstaller("", nil)

	tests := []struct {
		url1     string
		url2     string
		expected bool
	}{
		{"https://github.com/test/repo.git", "https://github.com/test/repo", true},
		{"https://github.com/test/repo/", "https://github.com/test/repo", true},
		{"https://github.com/a/b", "https://github.com/c/d", false},
		{"", "https://github.com/test/repo", false},
		{"https://github.com/test/repo", "", false},
	}

	for _, tt := range tests {
		got := u.matchesSourceURL(tt.url1, tt.url2)
		if got != tt.expected {
			t.Errorf("matchesSourceURL(%q, %q) = %v, want %v", tt.url1, tt.url2, got, tt.expected)
		}
	}
}
