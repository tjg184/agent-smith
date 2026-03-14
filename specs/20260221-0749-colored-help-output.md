# PRD: Colored Help Output for All Agent-Smith Commands

**Status:** ✅ COMPLETED  
**Priority:** High  
**Actual Time:** ~2 hours  
**Created:** 2026-02-21  
**Completed:** 2026-02-21

---

## 📋 Overview

Implement colored, visually enhanced help output for all agent-smith commands using custom Cobra templates. The goal is to extend the beautiful colored welcome screen aesthetic to all `--help` output, making commands, examples, and sections visually distinct and easier to read.

---

## 🎯 Goals

1. **Primary Goal:** Add colored output to all command help screens matching the welcome screen aesthetic
2. **Consistency:** Ensure uniform color scheme across all ~28 commands and subcommands
3. **Accessibility:** Respect TTY detection and NO_COLOR environment variable
4. **Maintainability:** Implement once via custom Cobra templates (no per-command changes needed)

---

## 🎨 Design Specifications

### Color Scheme (matching `cmd/root.go` welcome screen)

| Element | Color | Package Function | Example |
|---------|-------|-----------------|---------|
| Section Headers | **Cyan Bold** | `colors.InfoBold` | `USAGE:`, `EXAMPLES:`, `FLAGS:` |
| Command Examples | **Green** | `colors.Success` | `agent-smith install all` |
| Parameters | **Yellow** | `colors.Warning` | `<repository-url>`, `<skill-name>` |
| Comments | **Gray/Muted** | `colors.Muted` | `# Install a skill from GitHub` |
| URLs/Paths | **Cyan** | `colors.Info` | `https://github.com/owner/repo` |
| Command Names (list) | **Green** | `colors.Success` | In "Available Commands" section |
| Flag Names | **Cyan** | `colors.Info` | `--profile`, `--debug`, `-h` |
| Hint Text | **Gray/Muted** | `colors.Muted` | "Use ... for more information" |

### Visual Examples

**Before:**
```
USAGE:
  agent-smith install skill <repository-url> <skill-name>

EXAMPLES:
  # Install a specific skill from GitHub
  agent-smith install skill openai/cookbook gpt-skill
```

**After (with colors):**
```
USAGE:                                          [Cyan Bold]
  agent-smith install skill <repository-url> <skill-name>
      └─ Green ──┘         └── Yellow ──┘

EXAMPLES:                                       [Cyan Bold]
  # Install a specific skill from GitHub        [Gray/Muted]
  agent-smith install skill openai/cookbook gpt-skill
      └────── Green ──────┘
```

---

## 🏗️ Architecture

### Package Structure
```
pkg/help/
├── template.go       # Custom Cobra templates with color injection (100 lines)
├── formatter.go      # Text parsing and colorization logic (200 lines)
└── formatter_test.go # Comprehensive tests (150 lines)
```

### Integration Point
```
cmd/root.go
├── Import: github.com/tjg184/agent-smith/pkg/help
└── init() function: Add help.SetupCustomTemplates(rootCmd)
```

### Core Components

1. **Template Engine** (`pkg/help/template.go`)
   - `SetupCustomTemplates(rootCmd *cobra.Command)` - Main entry point
   - `getHelpTemplate() string` - Custom help template with colorization
   - `getUsageTemplate() string` - Custom usage template for command lists
   - Template function registration via `cobra.AddTemplateFunc()`

2. **Text Formatter** (`pkg/help/formatter.go`)
   - `ColorizeText(text string) string` - Main colorization function
   - `colorizeLine(line string) string` - Line-by-line processing
   - Pattern detection: Section headers, commands, parameters, comments, URLs
   - Pattern colorization: Apply colors via `pkg/colors` package
   - Edge case handling: Multi-pattern lines, indentation preservation, emoji preservation

3. **Tests** (`pkg/help/formatter_test.go`)
   - Pattern detection tests
   - Colorization logic tests
   - Edge case tests (multi-pattern lines, indentation, emojis)
   - Color disabling tests (NO_COLOR, TTY detection)

---

## 📝 Implementation Tasks

### Task 1: Create Help Package Structure
**File:** `pkg/help/template.go`  
**Time:** 15 minutes  
**Status:** Pending

**Requirements:**
- Create `SetupCustomTemplates()` function to register templates
- Implement `getHelpTemplate()` with custom colorization
- Implement `getUsageTemplate()` for command lists
- Register template functions via `cobra.AddTemplateFunc()`

**Acceptance Criteria:**
- [ ] `pkg/help/template.go` created with all functions
- [ ] Template functions registered in `init()`
- [ ] Templates use Cobra's template syntax correctly
- [ ] Templates inject colorization via custom functions

---

### Task 2: Implement Text Formatter
**File:** `pkg/help/formatter.go`  
**Time:** 60 minutes (30 min basic + 30 min advanced)  
**Status:** Pending

**Requirements:**
- Implement `ColorizeText()` main entry point
- Implement `colorizeLine()` for line-by-line processing
- Pattern detection functions:
  - `isSectionHeader(line string) bool` - Detect `UPPERCASE:` headers
  - `isComment(line string) bool` - Detect `#` comments
  - `isCommandExample(line string) bool` - Detect `agent-smith` commands
  - `hasURL(line string) bool` - Detect URLs
- Colorization functions:
  - `colorizeSection(line string) string` - Apply cyan bold
  - `colorizeComment(line string) string` - Apply gray/muted
  - `colorizeCommand(line string) string` - Apply green + yellow for params
  - `colorizeURLs(line string) string` - Apply cyan to URLs
- Edge case handlers:
  - `colorizeMultiPattern(line string) string` - Handle multiple patterns on one line
  - `getIndentation(line string) string` - Preserve indentation

**Acceptance Criteria:**
- [ ] All pattern detection functions implemented with regex
- [ ] All colorization functions implemented using `pkg/colors`
- [ ] Respects `colors.IsEnabled()` for TTY detection
- [ ] Preserves indentation and formatting
- [ ] Handles multi-pattern lines correctly
- [ ] Preserves emoji characters

---

### Task 3: Integrate with Root Command
**File:** `cmd/root.go`  
**Time:** 10 minutes  
**Status:** Pending

**Requirements:**
- Add import: `"github.com/tjg184/agent-smith/pkg/help"`
- Add one line in `init()`: `help.SetupCustomTemplates(rootCmd)`
- Ensure it runs before command definitions

**Acceptance Criteria:**
- [ ] Import added to `cmd/root.go`
- [ ] `help.SetupCustomTemplates(rootCmd)` called in `init()`
- [ ] Placement is correct (before command definitions)
- [ ] No build errors

---

### Task 4: Create Comprehensive Tests
**File:** `pkg/help/formatter_test.go`  
**Time:** 25 minutes  
**Status:** Pending

**Requirements:**
- Test pattern detection functions
- Test colorization functions
- Test edge cases (multi-pattern, indentation, emojis)
- Test color disabling (NO_COLOR, colors.Disable())
- Test with actual help text samples

**Test Cases:**
```go
// Pattern Detection Tests
TestIsSectionHeader()
TestIsComment()
TestIsCommandExample()
TestHasURL()

// Colorization Tests
TestColorizeSection()
TestColorizeComment()
TestColorizeCommand()
TestColorizeURLs()
TestColorizeMultiPattern()

// Edge Case Tests
TestPreserveIndentation()
TestPreserveEmojis()
TestMultiplePatternsSameLine()

// Integration Tests
TestColorizeText()
TestColorizeWithColorsDisabled()
```

**Acceptance Criteria:**
- [ ] All test cases pass
- [ ] Test coverage > 80%
- [ ] Tests verify color disabling works
- [ ] Tests verify indentation preservation
- [ ] Tests verify emoji preservation

---

### Task 5: Test All Commands
**Testing:** Manual verification  
**Time:** 30 minutes  
**Status:** Pending

**Requirements:**
- Test all 8 top-level commands
- Test ~20 subcommands
- Verify color scheme matches specification
- Test color disabling (NO_COLOR, pipe, redirect)

**Test Commands:**
```bash
# Top-level commands
./agent-smith install --help
./agent-smith link --help
./agent-smith unlink --help
./agent-smith profile --help
./agent-smith materialize --help
./agent-smith target --help
./agent-smith update --help
./agent-smith uninstall --help
./agent-smith status --help

# Subcommands (sampling)
./agent-smith install skill --help
./agent-smith install all --help
./agent-smith link skill --help
./agent-smith link status --help
./agent-smith profile create --help
./agent-smith profile list --help

# Color disabling
NO_COLOR=1 ./agent-smith install --help
./agent-smith install --help | cat
./agent-smith install --help > /tmp/help.txt && cat /tmp/help.txt
```

**Visual Checklist:**
- [ ] Section headers are Cyan Bold
- [ ] Command examples are Green
- [ ] Parameters are Yellow
- [ ] Comments are Gray/Muted
- [ ] URLs are Cyan
- [ ] Available Commands list shows names in Green
- [ ] Flags are Cyan
- [ ] Indentation preserved
- [ ] Emojis intact (📦, 👤, →)
- [ ] Colors disabled when piped
- [ ] Colors disabled when redirected
- [ ] NO_COLOR env var respected

**Acceptance Criteria:**
- [ ] All top-level commands show correct colors
- [ ] All tested subcommands show correct colors
- [ ] Color disabling works in all scenarios
- [ ] No formatting regressions
- [ ] No missing or broken help text

---

### Task 6: Build Verification
**Time:** 10 minutes  
**Status:** Pending

**Requirements:**
- Run `go build` and verify no errors
- Run `go test ./...` and verify all tests pass
- Run `go vet ./...` and verify no issues
- Run `golangci-lint run` (if available)

**Acceptance Criteria:**
- [ ] `go build` succeeds
- [ ] `go test ./...` all tests pass
- [ ] `go vet ./...` reports no issues
- [ ] No linting errors

---

## 📊 Success Metrics

### Functional Requirements
- ✅ All help screens show colored output
- ✅ Color scheme matches welcome screen aesthetic
- ✅ Colors automatically disabled when output is piped/redirected
- ✅ NO_COLOR environment variable is respected
- ✅ All existing help text structure is preserved
- ✅ No regression in help functionality

### Technical Requirements
- ✅ Tests pass for all colorization logic (>80% coverage)
- ✅ No new dependencies required
- ✅ Single integration point (cmd/root.go init())
- ✅ Indentation and formatting preserved
- ✅ Works for all 8 top-level commands and ~20 subcommands

### Code Quality
- ✅ Clean separation of concerns (template vs formatter)
- ✅ Comprehensive test coverage
- ✅ Well-documented functions
- ✅ Follows Go conventions

---

## 🔄 Dependencies

**Existing Packages (no new dependencies):**
- `github.com/spf13/cobra` v1.10.2 - CLI framework
- `github.com/fatih/color` v1.18.0 - Terminal colors
- `github.com/tjg184/agent-smith/pkg/colors` - Centralized color system

---

## 📂 File Changes Summary

### New Files (3)
1. `pkg/help/template.go` (~100 lines)
2. `pkg/help/formatter.go` (~200 lines)
3. `pkg/help/formatter_test.go` (~150 lines)

### Modified Files (1)
1. `cmd/root.go` (~2 lines added)

**Total new code:** ~450 lines  
**Total modified code:** ~2 lines

---

## ⏱️ Time Estimate

| Task | Time | Status |
|------|------|--------|
| Task 1: Create help package structure | 15 min | Pending |
| Task 2: Implement text formatter | 60 min | Pending |
| Task 3: Integrate with root command | 10 min | Pending |
| Task 4: Create comprehensive tests | 25 min | Pending |
| Task 5: Test all commands | 30 min | Pending |
| Task 6: Build verification | 10 min | Pending |
| **Total** | **2.5 hours** | **0% Complete** |

---

## 🚀 Implementation Order

1. **Task 1** → Create `pkg/help/template.go` skeleton
2. **Task 2** → Implement `pkg/help/formatter.go` with all logic
3. **Task 3** → Integrate with `cmd/root.go`
4. **Quick Test** → Run `./agent-smith install --help` to verify basic colorization
5. **Task 4** → Create tests and verify logic
6. **Task 5** → Comprehensive manual testing
7. **Task 6** → Final build verification

---

## 📝 Notes

- This implementation uses **Option 1: Custom Cobra Template** approach
- No changes needed to individual command definitions (all ~28 commands inherit automatically)
- Color system already exists in `pkg/colors/colors.go` with TTY detection
- Welcome screen colorization in `cmd/root.go` lines 52-94 serves as reference
- Section headers follow pattern: `^[A-Z][A-Z ]+:$` (e.g., `USAGE:`, `EXAMPLES:`)

---

## ✅ Acceptance Criteria

### Definition of Done
- [ ] All 6 tasks completed
- [ ] All tests pass
- [ ] All commands verified with colored output
- [ ] Color disabling verified (NO_COLOR, pipe, redirect)
- [ ] No build errors or warnings
- [ ] No regression in help functionality
- [ ] Code reviewed and follows Go conventions
- [ ] Documentation updated (if needed)

---

## 🔍 Testing Strategy

### Unit Tests (pkg/help/formatter_test.go)
- Pattern detection accuracy
- Colorization correctness
- Edge case handling
- Color disabling behavior

### Integration Tests (manual)
- All command help screens
- Color scheme consistency
- TTY detection
- Environment variable handling

### Regression Tests
- Verify no existing help text broken
- Verify indentation preserved
- Verify formatting preserved
- Verify emoji characters intact

---

## 🎯 Current Status

**Overall Progress:** 0% (0/6 tasks complete)

**Next Steps:**
1. Create `pkg/help/template.go`
2. Create `pkg/help/formatter.go`
3. Implement colorization logic
4. Integrate with `cmd/root.go`
5. Test and verify

**Blockers:** None

---

**Last Updated:** 2026-02-21 07:49  
**Updated By:** Agent (OpenCode)
