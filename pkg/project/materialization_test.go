package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/models"
)

func emptyLockFile() *models.ComponentLockFile {
	return &models.ComponentLockFile{
		Version:  models.CurrentLockFileVersion,
		Skills:   make(map[string]map[string]models.ComponentEntry),
		Agents:   make(map[string]map[string]models.ComponentEntry),
		Commands: make(map[string]map[string]models.ComponentEntry),
	}
}

func addEntryToLockFile(lf *models.ComponentLockFile, componentType, sourceURL, name, filesystemName string) {
	var targetMap map[string]map[string]models.ComponentEntry
	switch componentType {
	case "skills":
		targetMap = lf.Skills
	case "agents":
		targetMap = lf.Agents
	case "commands":
		targetMap = lf.Commands
	}
	if targetMap[sourceURL] == nil {
		targetMap[sourceURL] = make(map[string]models.ComponentEntry)
	}
	targetMap[sourceURL][name] = models.ComponentEntry{
		FilesystemName: filesystemName,
		SourceUrl:      sourceURL,
		Version:        models.CurrentLockFileVersion,
	}
}

// ResolveFilesystemName

func TestResolveFilesystemName_NoCollision(t *testing.T) {
	targetDir := t.TempDir()
	lf := emptyLockFile()

	got := ResolveFilesystemName(targetDir, "skills", "my-skill", "https://github.com/a/b", lf)
	if got != "my-skill" {
		t.Errorf("expected 'my-skill', got %q", got)
	}
}

func TestResolveFilesystemName_DiskCollisionGetsSuffix2(t *testing.T) {
	targetDir := t.TempDir()
	lf := emptyLockFile()

	// Create the base name on disk to force a collision
	if err := os.MkdirAll(filepath.Join(targetDir, "my-skill"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	got := ResolveFilesystemName(targetDir, "skills", "my-skill", "https://github.com/new/repo", lf)
	if got != "my-skill-2" {
		t.Errorf("expected 'my-skill-2', got %q", got)
	}
}

func TestResolveFilesystemName_DiskCollisionBothSuffixes(t *testing.T) {
	targetDir := t.TempDir()
	lf := emptyLockFile()

	// Occupy base name and -2 on disk
	if err := os.MkdirAll(filepath.Join(targetDir, "my-skill"), 0755); err != nil {
		t.Fatalf("mkdir base: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(targetDir, "my-skill-2"), 0755); err != nil {
		t.Fatalf("mkdir -2: %v", err)
	}

	got := ResolveFilesystemName(targetDir, "skills", "my-skill", "https://github.com/new/repo", lf)
	if got != "my-skill-3" {
		t.Errorf("expected 'my-skill-3', got %q", got)
	}
}

func TestResolveFilesystemName_MetadataOnlyCollision(t *testing.T) {
	targetDir := t.TempDir()
	lf := emptyLockFile()

	// Add base name to metadata only (no disk directory)
	addEntryToLockFile(lf, "skills", "https://github.com/existing/repo", "my-skill", "my-skill")

	got := ResolveFilesystemName(targetDir, "skills", "my-skill", "https://github.com/new/repo", lf)
	if got != "my-skill-2" {
		t.Errorf("expected 'my-skill-2' for metadata-only collision, got %q", got)
	}
}

func TestResolveFilesystemName_IdempotentForSameSource(t *testing.T) {
	targetDir := t.TempDir()
	lf := emptyLockFile()

	sourceURL := "https://github.com/existing/repo"
	addEntryToLockFile(lf, "skills", sourceURL, "my-skill", "my-skill-2")

	got := ResolveFilesystemName(targetDir, "skills", "my-skill", sourceURL, lf)
	if got != "my-skill-2" {
		t.Errorf("expected idempotent return of 'my-skill-2', got %q", got)
	}
}

// AddMaterializationEntry

func TestAddMaterializationEntry_StoresAllFields(t *testing.T) {
	lf := emptyLockFile()

	AddMaterializationEntry(lf, "skills", "my-skill", "https://github.com/a/b", "local", "dev", "abc123", "skills/my-skill", "src-hash", "cur-hash", "my-skill")

	entry, ok := lf.Skills["https://github.com/a/b"]["my-skill"]
	if !ok {
		t.Fatal("expected entry to exist in Skills map")
	}
	if entry.CommitHash != "abc123" {
		t.Errorf("CommitHash: got %q, want 'abc123'", entry.CommitHash)
	}
	if entry.SourceHash != "src-hash" {
		t.Errorf("SourceHash: got %q, want 'src-hash'", entry.SourceHash)
	}
	if entry.CurrentHash != "cur-hash" {
		t.Errorf("CurrentHash: got %q, want 'cur-hash'", entry.CurrentHash)
	}
	if entry.FilesystemName != "my-skill" {
		t.Errorf("FilesystemName: got %q, want 'my-skill'", entry.FilesystemName)
	}
	if entry.SourceProfile != "dev" {
		t.Errorf("SourceProfile: got %q, want 'dev'", entry.SourceProfile)
	}
	if entry.MaterializedAt == "" {
		t.Error("expected MaterializedAt to be set")
	}
}

func TestAddMaterializationEntry_UnknownTypeIsNoOp(t *testing.T) {
	lf := emptyLockFile()
	AddMaterializationEntry(lf, "unknown", "my-skill", "https://github.com/a/b", "local", "dev", "abc123", "", "", "", "my-skill")
	// Should not panic and should not modify any map
	if len(lf.Skills)+len(lf.Agents)+len(lf.Commands) != 0 {
		t.Error("expected no entries for unknown component type")
	}
}

// GetMaterializationComponentMap

func TestGetMaterializationComponentMap_ReturnsCorrectMap(t *testing.T) {
	lf := emptyLockFile()
	addEntryToLockFile(lf, "skills", "https://github.com/a/b", "s1", "s1")
	addEntryToLockFile(lf, "agents", "https://github.com/a/b", "a1", "a1")
	addEntryToLockFile(lf, "commands", "https://github.com/a/b", "c1", "c1")

	tests := []struct {
		componentType string
		expectedName  string
	}{
		{"skills", "s1"},
		{"agents", "a1"},
		{"commands", "c1"},
	}

	for _, tt := range tests {
		m := GetMaterializationComponentMap(lf, tt.componentType)
		if m == nil {
			t.Errorf("expected non-nil map for %s", tt.componentType)
			continue
		}
		found := false
		for _, entries := range m {
			if _, ok := entries[tt.expectedName]; ok {
				found = true
			}
		}
		if !found {
			t.Errorf("expected to find %q in %s map", tt.expectedName, tt.componentType)
		}
	}
}

func TestGetMaterializationComponentMap_UnknownTypeReturnsNil(t *testing.T) {
	lf := emptyLockFile()
	m := GetMaterializationComponentMap(lf, "unknown")
	if m != nil {
		t.Errorf("expected nil for unknown component type, got %v", m)
	}
}

// UpdateMaterializationEntry

func TestUpdateMaterializationEntry_UpdatesHashes(t *testing.T) {
	targetDir := t.TempDir()
	lf := emptyLockFile()

	sourceURL := "https://github.com/a/b"
	addEntryToLockFile(lf, "skills", sourceURL, "my-skill", "my-skill")

	// Also write a lock file on disk so UpdateMaterializationEntry can read commit hash
	lockContent := `{"version":5,"skills":{"https://github.com/a/b":{"my-skill":{"source":"https://github.com/a/b","sourceUrl":"https://github.com/a/b","commitHash":"newcommit","version":5}}},"agents":{},"commands":{}}`
	lockPath := filepath.Join(targetDir, ".component-lock.json")
	if err := os.WriteFile(lockPath, []byte(lockContent), 0644); err != nil {
		t.Fatalf("write lock file: %v", err)
	}

	err := UpdateMaterializationEntry(lf, targetDir, "skills", "my-skill", "new-src-hash", "new-cur-hash")
	if err != nil {
		t.Fatalf("UpdateMaterializationEntry error: %v", err)
	}

	entry := lf.Skills[sourceURL]["my-skill"]
	if entry.SourceHash != "new-src-hash" {
		t.Errorf("SourceHash: got %q, want 'new-src-hash'", entry.SourceHash)
	}
	if entry.CurrentHash != "new-cur-hash" {
		t.Errorf("CurrentHash: got %q, want 'new-cur-hash'", entry.CurrentHash)
	}
	if entry.MaterializedAt == "" {
		t.Error("expected MaterializedAt to be updated")
	}
}

func TestUpdateMaterializationEntry_UnknownType(t *testing.T) {
	lf := emptyLockFile()
	err := UpdateMaterializationEntry(lf, t.TempDir(), "unknown", "my-skill", "h1", "h2")
	if err == nil {
		t.Fatal("expected error for unknown component type, got nil")
	}
}

func TestUpdateMaterializationEntry_ComponentNotFound(t *testing.T) {
	lf := emptyLockFile()
	err := UpdateMaterializationEntry(lf, t.TempDir(), "skills", "nonexistent", "h1", "h2")
	if err == nil {
		t.Fatal("expected error for missing component, got nil")
	}
}
