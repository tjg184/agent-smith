# PRD: Modernize CLI Output

**Created**: 2026-02-01 12:14 UTC

---

## Introduction

Enhance Agent Smith's command-line interface output to provide a more modern, visually appealing, and easier-to-scan user experience while maintaining backward compatibility and accessibility. The current CLI uses basic Unicode symbols without colors and simple text separators. This modernization will add subtle ANSI colors, box-drawing characters for tables, improved error formatting, and consistent visual hierarchy across all commands, following a minimalist design philosophy similar to GitHub CLI.

## Goals

- Add subtle ANSI color support (green for success, red for errors, yellow for warnings, cyan for info) with auto-detection and graceful degradation
- Upgrade table and list formatting to use Unicode box-drawing characters for a cleaner, more professional appearance
- Improve inline error messages with better context, indentation, and helpful suggestions
- Establish consistent visual hierarchy and spacing patterns across all commands
- Maintain backward compatibility with non-TTY outputs, `NO_COLOR` environment variable, and existing command syntax
- Keep current Unicode symbols (✓, ✗, ⚠️) while adding supplemental symbols (⟳, →, •) for enhanced clarity
- Preserve existing progress indicator approach (progressbar for bulk operations, inline status for individual items)
- Use fixed-width formatting (80 characters) for tables and bordered sections

## User Stories

- [x] Story-001: As a CLI user, I want colored success messages so that I can quickly identify successful operations without reading detailed text.

  **Acceptance Criteria:**
  - Success messages display with green ✓ symbol
  - Component type labels (skill, agent, command) display in gray/muted color
  - Component names display in regular terminal color
  - Colors automatically disable for non-TTY outputs (pipes, redirects)
  - `NO_COLOR` environment variable support implemented
  - Color output works in common terminals (iTerm2, Terminal.app, VS Code integrated terminal)
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Color enabling/disabling logic based on TTY detection
  - `NO_COLOR` environment variable handling
  - Color code generation and wrapping functions
  
  **Integration Tests:**
  - Install command success message displays with colors in TTY
  - Install command success message has no ANSI codes when piped to file
  - Setting `NO_COLOR=1` disables all color output
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-002: As a CLI user, I want colored error messages with context so that I can quickly understand what went wrong and how to fix it.

  **Acceptance Criteria:**
  - Error messages display with red ✗ symbol and bold error text
  - Error details indented and displayed in muted gray color
  - Optional "Try:" section with helpful suggestions in regular color
  - Multiline error output uses consistent indentation (2 spaces)
  - Errors maintain clear hierarchy (main message → details → suggestions)
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Error formatting function with various error types
  - Indentation and line wrapping logic
  - Suggestion generation for common error patterns
  
  **Integration Tests:**
  - Failed download displays error with context and suggestion
  - Failed link operation shows helpful troubleshooting steps
  - Broken link detection provides clear remediation guidance
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-003: As a CLI user, I want tables formatted with box-drawing characters so that tabular information is easier to read and more visually organized.

  **Acceptance Criteria:**
  - Tables use Unicode box-drawing characters (┌─┐│├─┤└─┘)
  - Table borders are 80 characters wide (fixed width)
  - Headers separated from content with horizontal line (├─┤)
  - Colored status indicators (✓ green, ✗ red, ⚠ yellow) within tables
  - Table padding maintains alignment across all rows
  - Empty tables display cleanly without broken borders
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Box-drawing character constants and border generation functions
  - Table row formatting with various content lengths
  - Table width calculation and padding logic
  - Empty table rendering
  
  **Integration Tests:**
  - Installation summary table displays correctly with box-drawing
  - Profile list displays with proper borders and alignment
  - Link status table maintains consistent formatting
  - Target list table shows built-in and custom targets correctly
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-004: As a CLI user, I want consistent visual hierarchy across all commands so that I can easily scan output and find relevant information.

  **Acceptance Criteria:**
  - Section headers use consistent formatting (bold or colored)
  - Key-value pairs aligned consistently (label: value format)
  - Bullet points use • symbol for lists
  - Indentation follows 2-space standard throughout
  - "Next steps" sections use consistent formatting across commands
  - Summary statistics appear below bordered sections consistently
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Section header formatting function
  - Key-value pair alignment logic
  - Bullet list formatting
  - Indentation helper functions
  
  **Integration Tests:**
  - Status command displays sections with consistent headers
  - Install command shows "Next steps" in standard format
  - Profile list shows consistent bullet formatting
  - All commands maintain 2-space indentation
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-005: As a CLI user, I want the status command output modernized so that I can quickly understand my current configuration at a glance.

  **Acceptance Criteria:**
  - Status displays in bordered box with title "Current Configuration"
  - Active profile shows with green ✓ symbol
  - Detected targets listed in single line, comma-separated
  - Component counts organized by category with • bullets
  - Active profile components shown in separate section
  - "For more details" section uses • bullets and cyan-colored commands
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Status data collection and formatting logic
  - Active profile detection and display formatting
  - Component counting across different directories
  
  **Integration Tests:**
  - Status command with active profile displays correctly
  - Status command with no active profile shows "None"
  - Status command with custom targets includes them in output
  - Component counts match actual directory contents
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-006: As a CLI user, I want the install command output modernized so that I can see installation progress and results clearly.

  **Acceptance Criteria:**
  - Installation starts with "Installing components from {repo}" header
  - Progress bar maintains current format and behavior
  - Summary table uses box-drawing with sections for skills/agents/commands
  - Success count displays with green ✓ symbol below table
  - "Next steps" section shows common follow-up commands with • bullets
  - Commands in next steps highlighted in cyan color
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Installation summary table generation
  - Component grouping by type for display
  - Success/failure counting logic
  
  **Integration Tests:**
  - Bulk install displays progress bar correctly
  - Summary table groups components by type
  - Failed installations show in red with error details
  - Next steps section always displays after successful install
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-007: As a CLI user, I want the link command output modernized so that I can see linking progress and results at a glance.

  **Acceptance Criteria:**
  - Linking starts with "Linking components to {target}" header
  - Individual link progress displays as: "Linking {type}: {name}... ✓ Done" with green
  - Failed links show: "Linking {type}: {name}... ✗ Failed" in red with indented error
  - Summary table uses box-drawing showing success/skip/failure counts
  - Failed components listed below table with • bullets and indented errors
  - Skipped components explanation provided when applicable
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Link progress message formatting
  - Link summary statistics calculation
  - Failed component list generation
  
  **Integration Tests:**
  - Link all displays progress for each component
  - Link summary shows correct counts
  - Broken links reported with clear error messages
  - Skipped monorepo components explained in summary
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-008: As a CLI user, I want the profile list command output modernized so that I can easily see all my profiles and their contents.

  **Acceptance Criteria:**
  - Profile list displays in bordered box with title "Available Profiles"
  - Active profile marked with green ✓ symbol
  - Component counts shown in parentheses: "(X agents, Y skills, Z commands)"
  - Legend displayed below table explaining ✓ symbol
  - Total count shows below legend
  - Empty profiles show "(empty)" instead of component counts
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Profile component counting logic
  - Active profile detection and marking
  - Component count formatting with plural handling
  
  **Integration Tests:**
  - Profile list shows all profiles with correct counts
  - Active profile correctly marked with ✓
  - Empty profiles display correctly
  - Legend always present regardless of profile count
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-009: As a CLI user, I want the update command output modernized so that I can see what's being updated and track progress clearly.

  **Acceptance Criteria:**
  - Update starts with "Updating components in: {path}" header
  - Progress shows: "[1/15] {type}/{name}... ✓ Up to date" in green
  - Updates show: "[2/15] {type}/{name}... ⟳ Updating" then "✓ Updated successfully" in green
  - Summary table uses box-drawing showing "up to date" and "updated" counts
  - Update errors display in red with indented error details
  - ⟳ symbol used for "updating/syncing" status
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Update progress message formatting with item numbers
  - Update status detection (up-to-date vs needs update)
  - Update summary statistics calculation
  
  **Integration Tests:**
  - Update all checks each component and displays status
  - Components requiring updates show updating indicator
  - Summary shows correct counts for updated vs up-to-date
  - Failed updates shown in red with error context
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-010: As a CLI user, I want the uninstall command output modernized so that I understand what will be removed before confirmation.

  **Acceptance Criteria:**
  - Components to remove displayed in bordered box with title
  - Components grouped by type (Skills, Agents, Commands) with counts
  - Each component listed with • bullet
  - Warning message shows: "⚠ This will unlink and delete these components." in yellow
  - Confirmation prompt displays: "Continue? [y/N]:"
  - Removal progress shows: "✓ Removed {type}: {name}" in green for each
  - Final summary: "✓ Successfully removed X components" in green
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Component grouping and listing logic
  - Confirmation prompt handling
  - Removal progress message formatting
  
  **Integration Tests:**
  - Bulk uninstall displays all components before confirmation
  - Confirmation prompt accepts y/n correctly
  - Removal executes only after confirmation
  - Each removed component displays success message
  - Final summary matches actual removal count
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-011: As a CLI user, I want the unlink command output modernized so that I can see what's being unlinked and from where.

  **Acceptance Criteria:**
  - Unlink starts with "Unlinking components from {target}" header
  - Individual unlink shows: "Unlinking {type}: {name}... ✓ Done" in green
  - Failed unlinks show: "Unlinking {type}: {name}... ✗ Failed" in red
  - Summary shows counts for successful and failed unlinks
  - Target-specific unlinking clearly indicates target name
  - Already unlinked components handled gracefully with appropriate message
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Unlink progress message formatting
  - Unlink summary statistics calculation
  - Target-specific vs all-target unlink logic
  
  **Integration Tests:**
  - Unlink all removes links from all detected targets
  - Target-specific unlink only affects specified target
  - Already unlinked components reported appropriately
  - Summary shows correct counts
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-012: As a CLI user, I want the target list command output modernized so that I can easily see all available targets and their status.

  **Acceptance Criteria:**
  - Target list displays with "Available Targets:" header
  - Built-in targets section with "Built-in Targets:" subheader
  - Custom targets section with "Custom Targets:" subheader (if any exist)
  - Each target shows: "{symbol} {name} {path} {status}"
  - Green ✓ for existing directories, gray - for missing
  - Status shows "(exists)" or "(not found)" in muted color
  - Legend below list explains symbols
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Target detection and status checking
  - Built-in vs custom target separation
  - Directory existence validation
  - Symbol selection based on status
  
  **Integration Tests:**
  - Target list shows all detected built-in targets
  - Custom targets appear in separate section
  - Missing target directories marked appropriately
  - Legend always displays regardless of target count
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-013: As a developer, I want a centralized color system so that all commands use consistent colors and handle TTY detection properly.

  **Acceptance Criteria:**
  - New file `internal/formatter/colors.go` created
  - Color constants defined: Red, Green, Yellow, Blue, Cyan, Gray
  - Style functions: Bold, Dim, Underline, Reset
  - TTY detection using `golang.org/x/term.IsTerminal()`
  - Colors auto-disable for non-TTY outputs
  - `NO_COLOR` environment variable support
  - `CLICOLOR` and `CLICOLOR_FORCE` environment variable support
  - Helper functions: Success(), Error(), Warning(), Info(), Muted()
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - TTY detection logic with mocked file descriptors
  - Environment variable handling (NO_COLOR, CLICOLOR, CLICOLOR_FORCE)
  - Color code wrapping functions
  - Color disabling when conditions not met
  - Style combination functions (e.g., Bold + Green)
  
  **Integration Tests:**
  - Colors appear in TTY output
  - No ANSI codes when output piped to file
  - NO_COLOR=1 disables colors
  - CLICOLOR=0 disables colors
  - CLICOLOR_FORCE=1 enables colors even for non-TTY
  
  **Component Browser Tests:**
  - Not applicable (internal library)

- [x] Story-014: As a developer, I want box-drawing utilities so that all commands can easily create consistent bordered sections.

  **Acceptance Criteria:**
  - New file `internal/formatter/boxes.go` created
  - Box-drawing constants: TopLeft, TopRight, BottomLeft, BottomRight, Horizontal, Vertical, LeftJoin, RightJoin, CrossJoin
  - Function: DrawBox(title, content, width) returns formatted box
  - Function: DrawHeader(text, width) returns top border with title
  - Function: DrawSeparator(width) returns horizontal separator (├─┤)
  - Function: DrawFooter(width) returns bottom border
  - Fixed width: 80 characters (default, configurable)
  - Proper padding calculation for content alignment
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Box-drawing character constants validation
  - Border generation with various widths
  - Title centering in header
  - Content padding and alignment
  - Multi-line content handling
  - Empty content handling
  
  **Integration Tests:**
  - DrawBox creates complete bordered section
  - DrawSeparator maintains width consistency
  - Content alignment preserved across multiple lines
  - Box rendering with colored content
  
  **Component Browser Tests:**
  - Not applicable (internal library)

- [ ] Story-015: As a developer, I want enhanced formatter methods so that all commands use consistent formatting patterns.

  **Acceptance Criteria:**
  - Update `internal/formatter/formatter.go` with new methods
  - Method: SuccessWithDetail(type, name, detail) for detailed success messages
  - Method: ErrorWithContext(message, error, suggestion) for contextual errors
  - Method: Section(title) for section headers
  - Method: Divider() for visual separators
  - Method: KeyValue(key, value) for aligned key-value pairs
  - Method: List(items) for bulleted lists with • symbol
  - Method: NextSteps(commands) for "Next steps" sections
  - All methods integrate color system from colors.go
  - All methods respect TTY detection and NO_COLOR
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Each formatter method with various inputs
  - Color integration in formatted output
  - Alignment and padding calculations
  - Empty input handling
  - Multi-line content wrapping
  
  **Integration Tests:**
  - Formatter methods used in actual commands produce expected output
  - Colors appear correctly in formatted output
  - Formatted output degrades gracefully without colors
  - Consistent spacing across different method calls
  
  **Component Browser Tests:**
  - Not applicable (internal library)

- [ ] Story-016: As a developer, I want updated table formatting so that summary tables use box-drawing and colors consistently.

  **Acceptance Criteria:**
  - Update `internal/formatter/tables.go` with box-drawing integration
  - DisplaySummaryTable() uses DrawBox() from boxes.go
  - Status indicators colored: ✓ green, ✗ red, ⚠ yellow
  - Component type headers bolded
  - Error details indented with → symbol in muted color
  - Summary statistics below table use colored symbols
  - Table width fixed at 80 characters
  - Proper column alignment maintained
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Table generation with various result sets
  - Component grouping by type
  - Status color application
  - Error detail indentation
  - Summary statistics calculation
  
  **Integration Tests:**
  - Installation summary displays with box-drawing
  - Link summary displays with box-drawing
  - Update summary displays with box-drawing
  - Tables render correctly with mixed success/failure results
  
  **Component Browser Tests:**
  - Not applicable (internal library)

- [ ] Story-017: As a developer, I want common style patterns extracted so that formatting is consistent and easy to maintain.

  **Acceptance Criteria:**
  - New file `internal/formatter/styles.go` created
  - Function: ProgressMessage(action, type, name, status) for "Linking skill: name... ✓ Done" format
  - Function: SummaryStats(success, skipped, failed) for consistent summary formatting
  - Function: ComponentCount(type, count) for "X agents, Y skills" format
  - Function: CommandHint(command, description) for cyan-colored command suggestions
  - All functions integrate colors and use consistent spacing
  - Plural handling for counts (1 agent vs 2 agents)
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Each style pattern function with edge cases
  - Plural handling with various counts (0, 1, 2+)
  - Color integration in styled output
  - Spacing and alignment consistency
  
  **Integration Tests:**
  - Style patterns used across multiple commands
  - Consistent appearance in install, link, update commands
  - Command hints display correctly in next steps sections
  
  **Component Browser Tests:**
  - Not applicable (internal library)

- [x] Story-018: As a developer, I want color support in the logger package so that log levels are visually distinct.

  **Acceptance Criteria:**
  - Update `pkg/logger/logger.go` with optional color support
  - [ERROR] tag displays in red when colors enabled
  - [WARN] tag displays in yellow when colors enabled
  - [INFO] tag displays in cyan when colors enabled
  - [DEBUG] tag displays in gray when colors enabled
  - Color support off by default, must be explicitly enabled
  - SetColorEnabled(bool) method to control color output
  - Logger respects TTY detection and NO_COLOR environment variable
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Logger color enabling/disabling
  - Level tag color application
  - TTY detection integration
  - NO_COLOR environment variable handling
  
  **Integration Tests:**
  - Colored log output in TTY with colors enabled
  - No ANSI codes when colors disabled
  - Logger colors respect NO_COLOR variable
  - Debug messages show with gray [DEBUG] tag
  
  **Component Browser Tests:**
  - Not applicable (internal library)

- [ ] Story-019: As a developer, I want configuration support for display settings so that users can control colors and formatting.

  **Acceptance Criteria:**
  - Update `~/.agent-smith/config.json` schema with display section
  - Config field: display.colors ("auto" | "always" | "never")
  - Config field: display.unicode ("auto" | "always" | "ascii")
  - Default values: colors "auto", unicode "auto"
  - Config loaded at startup and applied to formatter
  - "auto" respects TTY detection, "always" forces on, "never" forces off
  - Fallback to ASCII if unicode not supported (future enhancement)
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Config parsing for display section
  - Default value handling when section missing
  - Validation of valid option values
  - Invalid value handling (fallback to "auto")
  
  **Integration Tests:**
  - Config with colors "always" enables colors even in non-TTY
  - Config with colors "never" disables colors in TTY
  - Config with colors "auto" respects TTY detection
  - Missing display section uses defaults
  
  **Component Browser Tests:**
  - Not applicable (configuration)

- [ ] Story-020: As a developer, I want all fmt.Printf calls in main.go replaced with formatter calls so that output is consistent and maintainable.

  **Acceptance Criteria:**
  - All direct fmt.Printf/Println calls in main.go replaced with formatter methods
  - Profile list handler uses formatter.DrawBox() and formatter.List()
  - Profile show handler uses formatter.KeyValue() and formatter.Section()
  - Status handler uses formatter.DrawBox() and formatter.KeyValue()
  - Cherry-pick handler uses formatter.DrawBox() and formatter.List()
  - Target list handler uses formatter methods for consistent formatting
  - No raw fmt.Printf/Println for user-facing output (debug logs excepted)
  - Logger used for debug/info messages where appropriate
  
  **Testing Criteria:**
  
  **Unit Tests:**
  - Not applicable (integration-level changes)
  
  **Integration Tests:**
  - Profile list displays with box-drawing and colors
  - Status command output matches new format
  - Cherry-pick interface displays correctly
  - Target list shows with consistent formatting
  - All commands produce colored output in TTY
  - All commands produce clean output when piped
  
  **Component Browser Tests:**
  - Not applicable (CLI-only changes)

## Functional Requirements

- FR-1: The system SHALL implement ANSI color support with automatic TTY detection using `golang.org/x/term.IsTerminal()`
- FR-2: The system SHALL disable colors when `NO_COLOR` environment variable is set to any non-empty value
- FR-3: The system SHALL disable colors when output is redirected to a file or pipe (non-TTY)
- FR-4: The system SHALL support `CLICOLOR` environment variable (0=disable, 1=enable if TTY)
- FR-5: The system SHALL support `CLICOLOR_FORCE` environment variable (non-zero=force colors even for non-TTY)
- FR-6: The system SHALL use Unicode box-drawing characters (┌─┐│├─┤└─┘) for tables and bordered sections
- FR-7: The system SHALL use fixed-width formatting at 80 characters for tables and boxes
- FR-8: The system SHALL use green color for success indicators (✓ symbol and success messages)
- FR-9: The system SHALL use red color for error indicators (✗ symbol and error messages)
- FR-10: The system SHALL use yellow color for warning indicators (⚠ symbol and warning messages)
- FR-11: The system SHALL use cyan color for info labels and command hints
- FR-12: The system SHALL use gray/dim color for muted text (component type labels, error details)
- FR-13: The system SHALL maintain existing Unicode symbols (✓, ✗, ⚠️)
- FR-14: The system SHALL add supplemental symbols (⟳ for updating, → for indentation/linking, • for bullets)
- FR-15: The system SHALL preserve existing progress bar behavior from progressbar library
- FR-16: The system SHALL use 2-space indentation for hierarchical content
- FR-17: The system SHALL display "Next steps" sections after successful operations with common follow-up commands
- FR-18: The system SHALL format error messages with main message, indented details, and optional suggestions
- FR-19: The system SHALL group table content by component type (skills, agents, commands)
- FR-20: The system SHALL display summary statistics below tables using colored symbols
- FR-21: The system SHALL provide display configuration options in `~/.agent-smith/config.json`
- FR-22: The system SHALL support display.colors config option with values: "auto", "always", "never"
- FR-23: The system SHALL support display.unicode config option with values: "auto", "always", "ascii"
- FR-24: The system SHALL default to display.colors="auto" and display.unicode="auto"
- FR-25: The system SHALL apply config settings at startup before any output
- FR-26: The system SHALL use consistent key-value formatting for status and configuration displays
- FR-27: The system SHALL maintain backward compatibility with existing command syntax and flags
- FR-28: The system SHALL produce identical functional output with and without colors (only visual difference)
- FR-29: The system SHALL use bold text for error messages and important headers
- FR-30: The system SHALL display component counts with proper plural handling (1 agent vs 2 agents)

## Non-Goals (Out of Scope)

- No animated spinners or loading indicators (keeping current progress bar approach)
- No emoji additions beyond current Unicode symbols
- No Nerd Font icon requirements (basic Unicode only)
- No interactive prompts beyond existing confirmation dialogs
- No terminal resize detection or responsive width adjustment (fixed 80 characters)
- No color theme customization or custom color schemes
- No ASCII art or decorative elements
- No syntax highlighting for code snippets
- No mouse interaction support
- No terminal bell or notification sounds
- No progress percentage displays for bulk operations
- No JSON output format option
- No verbose output mode changes (keeping current --verbose and --debug flags)
- No changes to underlying command functionality (visual only)
- No changes to file formats or data structures
- No new command-line flags or options (except display config)

## Technical Considerations

### Dependencies

- **golang.org/x/term**: Already available (indirect dependency), used for TTY detection via `term.IsTerminal()`
- **No new external dependencies**: All color and formatting implemented using standard library and existing dependencies

### Backward Compatibility

- All changes are visual only, no functional behavior changes
- Commands continue to accept same arguments and flags
- Output piped to files or other commands remains clean (no ANSI codes)
- `NO_COLOR` environment variable respected per industry standard
- Config file changes are additive (missing display section uses defaults)

### Performance Impact

- Color code wrapping adds negligible overhead (string concatenation only)
- Box-drawing character rendering has no measurable performance impact
- TTY detection performed once at startup, cached for process lifetime
- No additional I/O operations beyond existing output

### Accessibility

- Colors enhance but don't replace symbols (✓, ✗, ⚠ always present)
- Color-blind friendly (symbols provide redundant information)
- Screen reader compatible (plain text with symbols)
- Works perfectly in monochrome mode (NO_COLOR or non-TTY)

### Testing Strategy

- Unit tests for all formatter functions (colors, boxes, styles)
- Integration tests for command output in TTY and non-TTY modes
- Manual testing in multiple terminal emulators (iTerm2, Terminal.app, VS Code)
- Test with NO_COLOR environment variable set
- Test with output redirection (> file.txt, | less)
- Test on different terminal color schemes (light and dark backgrounds)
- Test with narrow terminal widths to verify 80-character limit

### Migration Path

1. **Phase 1**: Create color and box-drawing infrastructure (colors.go, boxes.go, styles.go)
2. **Phase 2**: Enhance formatter package methods (formatter.go, tables.go)
3. **Phase 3**: Update command outputs one by one (install, link, status, etc.)
4. **Phase 4**: Replace direct fmt.Printf calls in main.go with formatter calls
5. **Phase 5**: Add display configuration support and documentation
6. **Phase 6**: Comprehensive testing and user feedback iteration

### Documentation Updates

- Update README.md with screenshots showing new output
- Document display configuration options in CONFIG.md
- Add environment variable documentation (NO_COLOR, CLICOLOR, CLICOLOR_FORCE)
- Update TESTING.md with display testing procedures
- Add accessibility section explaining color-blind compatibility

## Success Metrics

- All commands produce colored output when run in a TTY
- All commands produce clean output (no ANSI codes) when piped or redirected
- NO_COLOR environment variable completely disables colors
- All tables use box-drawing characters consistently
- Error messages include helpful suggestions for common issues
- User feedback indicates improved readability and ease of use
- No regression in command functionality or performance
- Test coverage maintained or improved for all modified code

## Timeline Estimate

- **Week 1**: Infrastructure (colors.go, boxes.go, styles.go) - 3-5 days
- **Week 2**: Formatter enhancements (formatter.go, tables.go) - 4-5 days
- **Week 3**: Command updates (install, link, status, profile commands) - 5-7 days
- **Week 4**: Remaining commands (uninstall, update, unlink, target commands) - 5-7 days
- **Week 5**: Main.go refactoring and configuration support - 3-4 days
- **Week 6**: Testing, documentation, and polish - 5-7 days

**Total estimated time**: 25-35 development days

## References

- GitHub CLI (gh) - Minimalist colored output inspiration
- Industry standard: https://no-color.org/ - NO_COLOR environment variable
- Unicode box-drawing characters: https://en.wikipedia.org/wiki/Box-drawing_character
- ANSI color codes: https://en.wikipedia.org/wiki/ANSI_escape_code
- golang.org/x/term package documentation
