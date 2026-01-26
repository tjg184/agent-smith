# PRD: Simplify Install Output

## Introduction

Reduce visual clutter in the agent-smith install output by consolidating repetitive success messages into concise summaries. Currently, each component installation outputs 2-3 lines of information (download confirmation, storage path, component count), which creates excessive verbosity especially during bulk operations.

## Goals

- Minimize repetitive messages when installing multiple components
- Provide one-line summaries for single component installations
- Show progress indicators during bulk installations with summary tables at end
- Maintain clear success/failure feedback without verbose details
- Improve user experience with cleaner terminal output

## User Stories

- [x] Story-001: As a user, I want single component installations to show a concise one-line summary so that my terminal stays clean and readable.

  **Acceptance Criteria:**
  - Single installations display format: "Installed: component-name ✓"
  - Storage path information is removed from default output
  - Success checkmark (✓) or similar indicator shows installation succeeded
  - Failed installations show "Failed: component-name ✗" with error message
  
  **Testing Criteria:**
  **Unit Tests:**
  - Output formatting function tests
  - Success/failure message generation tests
  
  **Integration Tests:**
  - Single component installation output validation
  - Error handling and message format tests
  
  **Component Browser Tests:**
  - N/A (CLI tool, no browser components)

- [x] Story-002: As a user, I want bulk installations to show a progress indicator so that I know the system is working without verbose per-component messages.

  **Acceptance Criteria:**
  - Display "Installing X components..." at start of bulk operation
  - Show progress indicator (spinner or progress bar) during installation
  - Per-component output is minimal or hidden during installation
  - Progress indicator updates as each component completes
  
  **Testing Criteria:**
  **Unit Tests:**
  - Progress indicator state management tests
  - Component count tracking tests
  
  **Integration Tests:**
  - Bulk installation progress tracking tests
  - Progress indicator update sequence tests
  
  **Component Browser Tests:**
  - N/A (CLI tool, no browser components)

- [x] Story-003: As a user, I want to see a summary table after bulk installations so that I can quickly review what was installed successfully.

  **Acceptance Criteria:**
  - Summary table displays after all installations complete
  - Table shows component names grouped by type (skills, agents, commands)
  - Each row indicates success ✓ or failure ✗ status
  - Summary includes totals: "Successfully installed: X/Y components"
  - Failed components show brief error reason in summary
  
  **Testing Criteria:**
  **Unit Tests:**
  - Summary table formatting tests
  - Component grouping logic tests
  - Success/failure counting tests
  
  **Integration Tests:**
  - End-to-end bulk installation with summary generation
  - Mixed success/failure scenario summary tests
  
  **Component Browser Tests:**
  - N/A (CLI tool, no browser components)

- [x] Story-004: As a developer, I want to refactor output formatting logic into a centralized module so that output formatting is consistent across all downloaders.

  **Acceptance Criteria:**
  - Create new output formatter package/module
  - Centralize success/failure message formatting
  - Centralize progress indicator logic
  - Centralize summary table generation
  - Update command.go, bulk.go, skill.go, agent.go to use formatter
  
  **Testing Criteria:**
  **Unit Tests:**
  - Formatter module public API tests
  - Message formatting function tests
  - Table generation function tests
  
  **Integration Tests:**
  - All downloaders use centralized formatter
  - Consistent output format across component types
  
  **Component Browser Tests:**
  - N/A (CLI tool, no browser components)

## Functional Requirements

- FR-1: Single component installations must display a one-line summary in format "Installed: component-name ✓"
- FR-2: Bulk installations must show a progress indicator during installation
- FR-3: Bulk installations must display a summary table after completion
- FR-4: Summary table must group components by type (skills, agents, commands)
- FR-5: Summary table must show success ✓ or failure ✗ status for each component
- FR-6: Summary must include totals line showing "Successfully installed: X/Y components"
- FR-7: Failed installations must show brief error reasons in the summary
- FR-8: Output formatting must be centralized in a dedicated formatter module
- FR-9: All downloader types (command, skill, agent, bulk) must use the centralized formatter

## Non-Goals (Out of Scope)

- No --verbose or --quiet flags (simplified output is always used)
- No JSON or machine-readable output format
- No colored output or advanced terminal formatting (beyond checkmarks)
- No detailed storage path information in default output
- No component count details in output
- No URL information in success messages
- No download progress bars for individual files (only overall progress for bulk)
