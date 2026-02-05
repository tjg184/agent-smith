# PRD: Link Status Legend Box Table Improvement

**Created**: 2026-02-05 12:10 UTC

---

## Introduction

Improve the visual consistency and professional appearance of the legend section in `link status` commands by replacing bullet-point list formatting with BoxTable formatting, matching the polished look of the `install all` summary tables.

Currently, the `link status --all-profiles` legend displays symbols as list items with bullet points (`  • ✓  Valid symlink`), which looks less polished compared to the tabular format used in `install all` command summaries. This inconsistency degrades the overall user experience and visual appeal of the CLI tool.

## Goals

- Replace bullet-point legend formatting with BoxTable formatting for professional appearance
- Ensure consistent visual style between `install all` summaries and `link status` legends
- Update both `ShowLinkStatus()` and `ShowAllProfilesLinkStatus()` methods for consistency
- Maintain all existing functionality while improving presentation
- Add comprehensive test coverage to prevent regression

## User Stories

- [ ] Story-001: As a CLI tool, I want to add a DisplayLegendTable method to the Formatter so that legends can be displayed in a consistent, professional box table format across the application.

  **Acceptance Criteria:**
  - New `DisplayLegendTable(items []LegendItem)` method added to `/internal/formatter/formatter.go`
  - LegendItem struct with `Symbol` and `Description` fields defined
  - Method creates a two-column BoxTable with "Symbol" and "Meaning" headers
  - Each legend item rendered as a table row with symbol and description
  - Color formatting preserved (green checkmarks, red X's, etc.)
  - Method integrated with existing Formatter interface
  
  **Testing Criteria:**
  **Unit Tests:**
  - `TestDisplayLegendTable_WithMultipleItems` verifies correct table creation with various symbols
  - `TestDisplayLegendTable_WithColoredSymbols` ensures color codes preserved in output
  - `TestDisplayLegendTable_EmptyList` handles edge case gracefully

- [ ] Story-002: As a user running `agent-smith link status`, I want the legend to be displayed in a box table format so that it looks professional and consistent with other command outputs.

  **Acceptance Criteria:**
  - `ShowLinkStatus()` method in `/internal/linker/linker.go` updated to use `DisplayLegendTable()`
  - Legend section at lines 992-998 replaced with new table-based formatting
  - All five symbols displayed: ✓ (Valid symlink), ◆ (Copied directory), ✗ (Broken link), - (Not linked), ? (Unknown status)
  - Color formatting maintained (green for ✓, red for ✗, etc.)
  - Table borders and alignment match existing BoxTable styling
  - No changes to matrix table or summary sections
  
  **Testing Criteria:**
  **Unit Tests:**
  - Note: Legend formatting tested via formatter unit tests (Story-001)
  
  **Integration Tests:**
  - `TestLinkStatus_LegendDisplaysBoxTable` verifies box-drawing characters in legend output
  - Test ensures legend contains table borders (┌─┬─┐, ├─┼─┤, └─┴─┘)
  - Test verifies all five symbols present in legend table
  - Test confirms "Symbol" and "Meaning" headers exist

- [ ] Story-003: As a user running `agent-smith link status --all-profiles`, I want the legend to be displayed in a box table format so that it matches the single-profile view and looks professional.

  **Acceptance Criteria:**
  - `ShowAllProfilesLinkStatus()` method in `/internal/linker/linker.go` updated to use `DisplayLegendTable()`
  - Legend section at lines 1282-1288 replaced with new table-based formatting
  - All five symbols displayed consistently with single-profile view
  - Color formatting maintained for all symbols
  - Table formatting identical to single-profile view legend
  - No changes to multi-profile matrix table or summary sections
  
  **Testing Criteria:**
  **Unit Tests:**
  - Note: Legend formatting tested via formatter unit tests (Story-001)
  
  **Integration Tests:**
  - `TestLinkStatusAllProfiles_LegendDisplaysBoxTable` verifies box table in all-profiles output
  - Test ensures legend table structure matches single-profile view
  - Test verifies all five symbols present with correct meanings
  - Test confirms consistent formatting between views

- [ ] Story-004: As a developer, I want comprehensive integration tests for the updated legend formatting so that regressions are detected early and visual consistency is maintained.

  **Acceptance Criteria:**
  - Integration test file `tests/integration/link_status_legend_test.go` created
  - Test verifies box-drawing characters present in legend output
  - Test ensures all symbols (✓, ◆, ✗, -, ?) appear in legend table
  - Test validates "Symbol" and "Meaning" column headers
  - Test confirms legend appears after main status table
  - Test covers both single-profile and all-profiles views
  - Existing integration tests updated to handle new legend format
  
  **Testing Criteria:**
  **Integration Tests:**
  - `TestLinkStatusLegend_SingleProfile` validates legend table in default view
  - `TestLinkStatusLegend_AllProfiles` validates legend table in all-profiles view
  - `TestLinkStatusLegend_ContainsAllSymbols` ensures completeness
  - Tests verify box table structure using string matching for borders

## Functional Requirements

- FR-1: The system SHALL provide a `DisplayLegendTable()` method in the Formatter that accepts a list of LegendItem structs and renders them as a two-column BoxTable

- FR-2: The `DisplayLegendTable()` method SHALL create a table with "Symbol" and "Meaning" column headers

- FR-3: The legend SHALL preserve all existing color formatting (green ✓, red ✗, etc.) when rendering in table format

- FR-4: The `ShowLinkStatus()` method SHALL replace lines 992-998 with a call to `DisplayLegendTable()` using the same five legend items

- FR-5: The `ShowAllProfilesLinkStatus()` method SHALL replace lines 1282-1288 with a call to `DisplayLegendTable()` using the same five legend items

- FR-6: The legend table SHALL use the same box-drawing characters and formatting style as other BoxTable instances in the codebase

- FR-7: Integration tests SHALL verify the presence of box table borders (┌─┬─┐, ├─┼─┤, └─┴─┘) in legend output

- FR-8: The visual appearance SHALL match the professional styling of the `install all` summary tables

## Non-Goals (Out of Scope)

- No changes to the main status matrix table formatting
- No changes to the summary statistics section
- No modifications to symbol meanings or color schemes
- No changes to other commands' legend displays (only link status commands)
- No internationalization or localization of legend text
- No configuration options for legend display format
- No changes to the help text or documentation (separate task)

## Implementation Notes

### Color Enhancement (Added during implementation)

During implementation, colored symbols were added to the legend table for improved visual appeal:
- ✓ (green) - Valid symlink
- ◆ (green) - Copied directory  
- ✗ (red) - Broken link
- \- (gray) - Not linked
- ? (yellow) - Unknown status

This enhancement uses the existing `colors` package functions:
- `colors.Success()` for green symbols (✓, ◆)
- `colors.Error()` for red symbols (✗)
- `colors.Muted()` for gray symbols (-)
- `colors.Warning()` for yellow symbols (?)

Colors are automatically disabled when:
- `NO_COLOR` environment variable is set
- Output is not a TTY (e.g., piped to a file)
- Terminal doesn't support colors

This matches the existing color behavior throughout the application and provides better visual hierarchy in the legend.
