# PRD: Simplify Add-All Command Output Messages

## Introduction

The add-all command currently displays misleading output messages that make it appear as if the repository is being cloned multiple times. When processing multiple components from the same repository, each component download shows the full repository URL, creating confusion about whether the optimization (single clone) is working correctly.

## Goals

- Improve user experience by reducing confusing output messages
- Make it clear that repository is only cloned once through simplified output
- Reduce verbosity of command output for cleaner logs
- Prioritize user experience improvement over tool compatibility

## User Stories

- [x] Story-001: As a user running add-all, I want clear output that doesn't confuse me about multiple clones so I can understand that the optimization is working correctly.

  **Acceptance Criteria:**
  - Success messages show only component type and name (not repository URL)
  - Initial repository context message is preserved for clarity
  - Error messages retain full repository URL for debugging
  - All output messages maintain consistency

## Functional Requirements

- FR-1: The system must simplify "Successfully downloaded" messages from "Successfully downloaded [type] '[name]' from [url]" to "Successfully downloaded [type]: [name]"
- FR-2: The system must simplify "Downloading" messages from "Downloading [type]: [name]" to "Downloading: [name]"
- FR-3: The system must preserve initial repository cloning message for user context
- FR-4: The system must keep error messages unchanged to maintain debugging capability
- FR-5: The system must maintain output consistency across all component types (skills, agents, commands)

## Non-Goals

- No changes to underlying functionality or performance optimization
- No modifications to error handling or repository cloning logic
- No changes to summary statistics or final reporting
- No consideration for backward compatibility with output parsing tools

## Expected Output Transformation

**Before (current):**

```
Downloading skill: root-skill
Successfully downloaded skill 'root-skill' from https://github.com/wshobson/agents/

Downloading agent: observability-engineer  
Successfully downloaded agent 'observability-engineer' from https://github.com/wshobson/agents/
Downloading agent: performance-engineer
Successfully downloaded agent 'performance-engineer' from https://github.com/wshobson/agents/
```

**After (simplified):**

```
Downloading: root-skill
Successfully downloaded skill: root-skill

Downloading: observability-engineer
Successfully downloaded agent: observability-engineer
Downloading: performance-engineer  
Successfully downloaded agent: performance-engineer
```

## Implementation Requirements

- Modified output messages in BulkDownloader.AddAll method
- Changes to all three component processing loops (skills, agents, commands)
- Preservation of existing error message formatting
- No changes to core functionality or performance optimizations

