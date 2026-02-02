# Story-004: Provenance Tracking Implementation

## Status: ✅ COMPLETE

## Summary

Story-004 requires provenance tracking for materialized components so that project maintainers know where each materialized component came from. This implementation was already complete in the codebase, and comprehensive tests have been added to validate all acceptance criteria.

## Acceptance Criteria Validation

### 1. ✅ .materializations.json file created in .opencode/ or .claude/

**Implementation:**
- File is created automatically in `pkg/project/materialization.go`
- `SaveMaterializationMetadata()` creates the file with proper structure
- File location: `.opencode/.materializations.json` or `.claude/.materializations.json`

**Test Coverage:**
- `TestMaterializeProvenance` verifies file creation
- `TestMaterializeAllComponentTypes/VerifyMetadata` validates file structure

### 2. ✅ Metadata includes: source repo URL, source type, commit hash, original path, materialization timestamp

**Implementation:**
```go
type MaterializedComponentMetadata struct {
    Source         string `json:"source"`          // Source repo URL
    SourceType     string `json:"sourceType"`      // github/gitlab/local
    CommitHash     string `json:"commitHash"`      // Git commit hash
    OriginalPath   string `json:"originalPath"`    // Original path in repo
    MaterializedAt string `json:"materializedAt"`  // RFC3339 timestamp
    // ... other fields
}
```

**Test Coverage:**
- `TestMaterializeProvenance` validates all metadata fields
- Verifies data is correctly extracted from lock files
- Confirms timestamp is in RFC3339 format
- Validates timestamp is within expected time range

### 3. ✅ Metadata includes sourceHash and currentHash for future sync detection

**Implementation:**
```go
type MaterializedComponentMetadata struct {
    // ... other fields
    SourceHash     string `json:"sourceHash"`      // SHA-256 at materialization
    CurrentHash    string `json:"currentHash"`     // SHA-256 current state
}
```

- Hashes calculated using `materializer.CalculateDirectoryHash()`
- Uses SHA-256 algorithm with "sha256:" prefix
- Includes all files in directory (sorted by path for consistency)

**Test Coverage:**
- `TestMaterializeProvenance` verifies both hashes are present
- Confirms SHA-256 format with proper prefix
- Validates sourceHash == currentHash immediately after materialization
- `TestMaterializeProvenanceMultipleComponents` confirms different components have different hashes

### 4. ✅ Metadata loaded from existing ~/.agent-smith/.skill-lock.json (or agent/command lock files)

**Implementation:**
- `metadata.LoadLockFileEntry()` reads from lock files
- Supports skills, agents, and commands
- Lock files: `.skill-lock.json`, `.agent-lock.json`, `.command-lock.json`
- Data extracted: source, sourceType, sourceUrl, commitHash, originalPath

**Test Coverage:**
- All tests create lock files with metadata
- `TestMaterializeProvenance` uses comprehensive lock file with all fields
- Verifies metadata is correctly transferred from lock file to materialization metadata

### 5. ✅ JSON formatted with indentation for readability and git diffing

**Implementation:**
```go
data, err := json.MarshalIndent(metadata, "", "  ")  // 2-space indentation
```

**Test Coverage:**
- `TestMaterializeProvenance` verifies JSON contains newlines and indentation
- Confirms JSON is parseable and properly formatted

## Example Metadata Output

```json
{
  "version": 1,
  "skills": {
    "provenance-test": {
      "source": "https://github.com/example/provenance-test",
      "sourceType": "github",
      "commitHash": "fedcba9876543210",
      "originalPath": "skills/provenance-test/SKILL.md",
      "materializedAt": "2026-02-01T21:26:58-06:00",
      "sourceHash": "sha256:168c9a004f2d0...",
      "currentHash": "sha256:168c9a004f2d0..."
    }
  },
  "agents": {},
  "commands": {}
}
```

## Test Suite

### New Tests Added

1. **TestMaterializeProvenance** - Comprehensive validation of all Story-004 acceptance criteria
   - Creates test skill with nested directory structure
   - Verifies all metadata fields are correctly populated
   - Validates timestamp format and timing
   - Confirms hash calculation and format
   - Checks JSON formatting with indentation

2. **TestMaterializeProvenanceMultipleComponents** - Multi-component tracking
   - Tests multiple skills and agents
   - Verifies each component has unique metadata
   - Confirms different components have different hashes
   - Validates metadata persistence across multiple materializations

### Existing Tests

- `TestMaterializeAllComponentTypes/VerifyMetadata` - Basic metadata validation
- All other materialize tests implicitly test provenance tracking

## Files Modified

### New Files
- `tests/integration/materialize_provenance_test.go` - Comprehensive Story-004 tests

### Existing Files (No Changes)
The implementation was already complete in:
- `pkg/project/materialization.go` - Metadata structures and persistence
- `internal/materializer/materializer.go` - Hash calculation
- `main.go` - Materialize command implementation
- `internal/metadata/lock.go` - Lock file reading

## Test Results

```bash
$ cd tests/integration && go test -tags integration -run "TestMaterialize" -v

=== RUN   TestMaterializeAllComponentTypes
--- PASS: TestMaterializeAllComponentTypes (6.69s)

=== RUN   TestMaterializeComponentNotFound
--- PASS: TestMaterializeComponentNotFound (6.78s)

=== RUN   TestMaterializeRecursiveDirectoryStructure
--- PASS: TestMaterializeRecursiveDirectoryStructure (6.72s)

=== RUN   TestMaterializeFromNestedSubdirectory
--- PASS: TestMaterializeFromNestedSubdirectory (6.62s)

=== RUN   TestMaterializeWithProjectDirOverride
--- PASS: TestMaterializeWithProjectDirOverride (6.67s)

=== RUN   TestMaterializeNoProjectFound
--- PASS: TestMaterializeNoProjectFound (6.74s)

=== RUN   TestMaterializeStopsAtHomeDirectory
--- PASS: TestMaterializeStopsAtHomeDirectory (6.96s)

=== RUN   TestMaterializeWithRelativeProjectDir
--- PASS: TestMaterializeWithRelativeProjectDir (6.94s)

=== RUN   TestMaterializeProvenance
--- PASS: TestMaterializeProvenance (6.73s)

=== RUN   TestMaterializeProvenanceMultipleComponents
--- PASS: TestMaterializeProvenanceMultipleComponents (6.82s)

PASS
```

All 10 materialize tests pass successfully!

## Future Work

While Story-004 is complete, the following related stories are planned:

- **Story-011**: `materialize info <type> <name>` command to display provenance
- **Story-013**: Profile support (sourceProfile field)
- **Future**: Sync detection using hash comparison

## Conclusion

Story-004 acceptance criteria are **fully implemented and thoroughly tested**. The provenance tracking system:

1. ✅ Creates .materializations.json in project directories
2. ✅ Captures all required metadata from lock files
3. ✅ Calculates hashes for future sync detection
4. ✅ Formats JSON for readability and version control
5. ✅ Supports all component types (skills, agents, commands)
6. ✅ Works with both opencode and claudecode targets

The implementation enables project maintainers to track the origin and state of all materialized components, providing full transparency and enabling future sync/update features.
