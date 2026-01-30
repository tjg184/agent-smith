# Configuration Guide

This document provides a comprehensive guide to configuring agent-smith.

## Table of Contents

- [Configuration Files Overview](#configuration-files-overview)
- [User Configuration (config.json)](#user-configuration-configjson)
- [Project Configuration (.ralphy/config.yaml)](#project-configuration-ralphyconfigyaml)
- [Environment Variables](#environment-variables)
- [Validation Rules](#validation-rules)
- [Troubleshooting](#troubleshooting)
- [Examples](#examples)

## Configuration Files Overview

Agent Smith uses multiple configuration files for different purposes:

### 1. User Configuration: `~/.agent-smith/config.json`

**Purpose**: Global user preferences and custom target definitions

**Location**: `~/.agent-smith/config.json`

**Format**: JSON

**Created**: Automatically when you add custom targets

**Use case**: Configuring custom editors/tools to link components to

### 2. Project Configuration: `.ralphy/config.yaml`

**Purpose**: Project-specific AI assistant rules and settings

**Location**: `<project-root>/.ralphy/config.yaml`

**Format**: YAML

**Created**: Manually by developers

**Use case**: Setting project conventions, boundaries, and build commands for AI assistants

### 3. Lock Files

**Purpose**: Track installed components and their versions

**Location**: `~/.agent-smith/` and `~/.agent-smith/profiles/<profile-name>/`

**Files**:
- `.skill-lock.json` - Installed skills
- `.agent-lock.json` - Installed agents  
- `.command-lock.json` - Installed commands

**Format**: JSON

**Created**: Automatically when installing components

## User Configuration (config.json)

### File Structure

```json
{
  "version": 1,
  "customTargets": [
    {
      "name": "target-name",
      "baseDir": "~/.target-dir",
      "skillsDir": "skills",
      "agentsDir": "agents",
      "commandsDir": "commands"
    }
  ]
}
```

### Field Descriptions

#### `version` (required)

**Type**: `integer`

**Description**: Configuration schema version. Currently must be `1`.

**Purpose**: Allows the application to handle future breaking changes to the config format.

**Example**: `"version": 1`

#### `customTargets` (required)

**Type**: `array` of `CustomTargetConfig` objects

**Description**: List of custom target configurations for linking components to additional editors or tools.

**Default**: `[]` (empty array)

### CustomTargetConfig Object

Each custom target configuration has the following fields:

#### `name` (required)

**Type**: `string`

**Description**: Unique identifier for the target.

**Constraints**:
- Cannot be empty
- Must match pattern: `^[a-zA-Z0-9_-]+$` (alphanumeric, hyphens, underscores only)
- Must be unique across all targets (case-insensitive)
- Cannot conflict with built-in targets: `opencode`, `claudecode`

**Examples**:
- Valid: `"cursor"`, `"vscode-insiders"`, `"my_editor"`, `"editor2"`
- Invalid: `"my.editor"`, `"editor#1"`, `"editor/name"`, `"OpenCode"`

#### `baseDir` (required)

**Type**: `string`

**Description**: Root directory where the target stores its configuration.

**Constraints**:
- Cannot be empty
- Must be a valid filesystem path
- Supports tilde (`~`) expansion for home directory
- Will be expanded to absolute path during validation

**Examples**:
- `"~/.cursor"`
- `"~/.vscode-insiders"`
- `"/opt/my-editor"`
- `"~/Library/Application Support/MyEditor"`

#### `skillsDir` (required)

**Type**: `string`

**Description**: Subdirectory name (relative to baseDir) where skills will be linked.

**Constraints**:
- Cannot be empty
- Must be a simple directory name (no path separators)
- Cannot be `.` or `..`

**Examples**:
- Valid: `"skills"`, `"custom-skills"`, `"my_skills"`
- Invalid: `"skills/custom"`, `"."`, `".."`

**Result**: Skills will be symlinked to `<baseDir>/<skillsDir>/`

#### `agentsDir` (required)

**Type**: `string`

**Description**: Subdirectory name (relative to baseDir) where agents will be linked.

**Constraints**: Same as `skillsDir`

**Examples**: `"agents"`, `"custom-agents"`

**Result**: Agents will be symlinked to `<baseDir>/<agentsDir>/`

#### `commandsDir` (required)

**Type**: `string`

**Description**: Subdirectory name (relative to baseDir) where commands will be linked.

**Constraints**: Same as `skillsDir`

**Examples**: `"commands"`, `"custom-commands"`

**Result**: Commands will be symlinked to `<baseDir>/<commandsDir>/`

### Managing Custom Targets

#### Add a Custom Target

```bash
# Interactive mode (prompts for all fields)
agent-smith target add cursor

# Non-interactive mode (provide all flags)
agent-smith target add cursor \
  --base-dir ~/.cursor \
  --skills-dir skills \
  --agents-dir agents \
  --commands-dir commands
```

#### List All Targets

```bash
agent-smith target list
```

Output includes:
- Built-in targets (OpenCode, ClaudeCode) - if available
- Custom targets from config.json

#### Remove a Custom Target

```bash
agent-smith target remove cursor
```

This removes the target from config.json and unlinks any components linked to it.

### Complete Example

```json
{
  "version": 1,
  "customTargets": [
    {
      "name": "cursor",
      "baseDir": "~/.cursor",
      "skillsDir": "skills",
      "agentsDir": "agents",
      "commandsDir": "commands"
    },
    {
      "name": "vscode-insiders",
      "baseDir": "~/.vscode-insiders/User/globalStorage",
      "skillsDir": "agent-skills",
      "agentsDir": "agent-agents",
      "commandsDir": "agent-commands"
    },
    {
      "name": "custom-tool",
      "baseDir": "/opt/mytool/config",
      "skillsDir": "extensions/skills",
      "agentsDir": "extensions/agents",
      "commandsDir": "extensions/commands"
    }
  ]
}
```

## Project Configuration (.ralphy/config.yaml)

The Ralphy configuration file defines project-specific rules and settings for AI assistants.

### File Location

Create this file in your project root: `<project>/.ralphy/config.yaml`

### File Structure

```yaml
# Ralphy Configuration
# https://github.com/michaelshimeles/ralphy

# Project info (auto-detected, edit if needed)
project:
  name: "your-project-name"
  language: "Go"
  framework: "cobra"
  description: "Brief project description"

# Commands (auto-detected from package.json/pyproject.toml)
commands:
  test: "go test ./..."
  lint: "golangci-lint run"
  build: "go build -o bin/agent-smith"

# Rules - instructions the AI MUST follow
# These are injected into every prompt
rules:
  - "Use Go 1.23 features and best practices"
  - "Follow the existing error handling patterns"
  - "All new features must include tests"
  - "Use the cobra CLI framework for commands"

# Boundaries - files/folders the AI should not modify
boundaries:
  never_touch:
    - "vendor/**"
    - "bin/**"
    - "*.lock"
    - "go.sum"
```

### Field Descriptions

#### `project`

**project.name**: Name of your project

**project.language**: Primary programming language (e.g., "Go", "TypeScript", "Python")

**project.framework**: Main framework used (e.g., "cobra", "React", "FastAPI")

**project.description**: Brief description of what the project does

#### `commands`

Common commands for the project:

**commands.test**: Command to run tests (e.g., `"go test ./..."`)

**commands.lint**: Command to run linter (e.g., `"golangci-lint run"`)

**commands.build**: Command to build the project (e.g., `"go build"`)

#### `rules`

Array of instructions that AI assistants MUST follow. These are injected into every prompt.

**Use for**:
- Code style requirements
- Architecture patterns to follow
- Required validations or checks
- Technology preferences

**Examples**:
```yaml
rules:
  - "Always use TypeScript strict mode"
  - "Follow the error handling pattern in src/utils/errors.ts"
  - "All API endpoints must have input validation with Zod"
  - "Use server actions instead of API routes in Next.js"
  - "Database queries must use prepared statements"
  - "New components require Storybook stories"
```

#### `boundaries`

**boundaries.never_touch**: Array of glob patterns for files/folders AI should never modify.

**Use for**:
- Generated files
- Lock files
- Legacy code
- Vendor dependencies
- Migration files

**Examples**:
```yaml
boundaries:
  never_touch:
    - "src/legacy/**"
    - "migrations/**"
    - "*.lock"
    - "node_modules/**"
    - "generated/**"
    - ".env*"
```

### Complete Example

```yaml
# Ralphy Configuration
# https://github.com/michaelshimeles/ralphy

project:
  name: "agent-smith"
  language: "Go"
  framework: "cobra"
  description: "Component management system for AI coding assistants"

commands:
  test: "go test -v ./..."
  lint: "golangci-lint run"
  build: "go build -o bin/agent-smith"

rules:
  - "Use Go 1.23 features and follow Go best practices"
  - "Follow the existing cobra command structure"
  - "All new commands must include help text and examples"
  - "Public functions must have GoDoc comments"
  - "Error messages should be actionable and user-friendly"
  - "New features require integration tests"
  - "Use the existing config validation patterns"

boundaries:
  never_touch:
    - "vendor/**"
    - "bin/**"
    - "go.sum"
    - ".git/**"
    - "*.lock.json"
```

## Environment Variables

### `AGENT_SMITH_TARGET`

**Purpose**: Override the default target for linking operations.

**Type**: `string`

**Values**: Name of any available target (built-in or custom)

**Usage**:
```bash
# Link all components to a specific target
export AGENT_SMITH_TARGET=cursor
agent-smith link all

# Or use inline
AGENT_SMITH_TARGET=cursor agent-smith link all
```

**Default behavior**: If not set, links to all available targets.

## Validation Rules

### Config Version Validation

- Config must have `"version": 1`
- Future versions may be supported in future releases
- Incompatible versions will cause an error

### Custom Target Validation

#### Name Validation
- ✅ Cannot be empty
- ✅ Must match pattern: `^[a-zA-Z0-9_-]+$`
- ✅ Must be unique (case-insensitive)
- ✅ Cannot be "opencode" or "claudecode"

#### Path Validation
- ✅ `baseDir` cannot be empty
- ✅ `baseDir` must be a valid path
- ✅ Tilde (`~`) expansion is supported
- ✅ Converted to absolute path during validation

#### Subdirectory Validation
- ✅ `skillsDir`, `agentsDir`, `commandsDir` cannot be empty
- ✅ Cannot contain path separators (`/` or `\`)
- ✅ Cannot be `.` or `..`
- ✅ Must be simple directory names

### JSON Format Validation
- ✅ Must be valid JSON
- ✅ Must match the expected schema
- ✅ Unknown fields are ignored (forward compatibility)

## Troubleshooting

### Config file not found

**Problem**: `~/.agent-smith/config.json` doesn't exist

**Solution**: This is normal! The file is created automatically when you add your first custom target:

```bash
agent-smith target add cursor
```

### Invalid JSON format

**Problem**: Error message: "failed to parse config file: invalid character..."

**Solution**: 
1. Open `~/.agent-smith/config.json` in a text editor
2. Validate JSON at https://jsonlint.com
3. Common issues:
   - Missing comma between array items
   - Trailing comma after last item
   - Unquoted strings
   - Missing closing brace/bracket

### Version mismatch

**Problem**: Error message: "unsupported config version X (expected 1)"

**Solution**: Your config file is from a future or incompatible version. Either:
1. Update agent-smith to the latest version
2. Or manually edit config.json and set `"version": 1`

### Duplicate target name

**Problem**: Error message: "duplicate target name: X (names are case-insensitive)"

**Solution**: Choose a unique name. Remember that "Cursor" and "cursor" are considered the same.

### Target name conflicts with built-in

**Problem**: Error message: "target name X conflicts with built-in target"

**Solution**: Don't use "opencode" or "claudecode" as custom target names. Choose a different name.

### Invalid target name characters

**Problem**: Error message: "target name X contains invalid characters"

**Solution**: Use only letters, numbers, hyphens, and underscores. Examples:
- ✅ `cursor`, `vscode-insiders`, `my_editor`
- ❌ `my.editor`, `editor#1`, `editor/name`

### Invalid subdirectory name

**Problem**: Error message: "skillsDir cannot contain path separators"

**Solution**: Use simple directory names without slashes:
- ✅ `skills`, `custom-skills`
- ❌ `skills/custom`, `./skills`

### Path expansion issues

**Problem**: Tilde (`~`) not expanding correctly

**Solution**: The application automatically expands `~` to your home directory. If you're having issues:
1. Use absolute paths instead: `/Users/username/.cursor`
2. Check that the path exists: `ls -la ~/.cursor`

### Permission denied

**Problem**: Error when writing config file

**Solution**: Check directory permissions:
```bash
ls -la ~/.agent-smith/
chmod 755 ~/.agent-smith/
```

## Examples

### Example 1: Basic Cursor Configuration

```bash
# Add Cursor as a target
agent-smith target add cursor --base-dir ~/.cursor

# Install and link components
agent-smith install skill mcp-builder
agent-smith link skill mcp-builder --target cursor
```

Resulting config.json:
```json
{
  "version": 1,
  "customTargets": [
    {
      "name": "cursor",
      "baseDir": "~/.cursor",
      "skillsDir": "skills",
      "agentsDir": "agents",
      "commandsDir": "commands"
    }
  ]
}
```

### Example 2: Multiple Custom Editors

```bash
# Add multiple editors
agent-smith target add cursor --base-dir ~/.cursor
agent-smith target add vscode-insiders --base-dir ~/.vscode-insiders

# Link all components to all targets
agent-smith link all
```

Resulting config.json:
```json
{
  "version": 1,
  "customTargets": [
    {
      "name": "cursor",
      "baseDir": "~/.cursor",
      "skillsDir": "skills",
      "agentsDir": "agents",
      "commandsDir": "commands"
    },
    {
      "name": "vscode-insiders",
      "baseDir": "~/.vscode-insiders",
      "skillsDir": "skills",
      "agentsDir": "agents",
      "commandsDir": "commands"
    }
  ]
}
```

### Example 3: Custom Directory Structure

```bash
# Add target with custom subdirectory names
agent-smith target add mytool \
  --base-dir /opt/mytool \
  --skills-dir extensions/skills \
  --agents-dir extensions/agents \
  --commands-dir extensions/commands
```

Resulting config.json:
```json
{
  "version": 1,
  "customTargets": [
    {
      "name": "mytool",
      "baseDir": "/opt/mytool",
      "skillsDir": "extensions/skills",
      "agentsDir": "extensions/agents",
      "commandsDir": "extensions/commands"
    }
  ]
}
```

### Example 4: Project-Specific Ralphy Configuration

For a TypeScript/React project:

```yaml
# .ralphy/config.yaml
project:
  name: "my-react-app"
  language: "TypeScript"
  framework: "React"
  description: "Customer dashboard application"

commands:
  test: "npm test"
  lint: "npm run lint"
  build: "npm run build"

rules:
  - "Use TypeScript strict mode"
  - "Follow React hooks best practices"
  - "All components must be functional components"
  - "Use Tailwind CSS for styling"
  - "API calls must use the fetchApi wrapper"
  - "New features require unit tests with Jest"
  - "Follow the component structure in src/components"

boundaries:
  never_touch:
    - "node_modules/**"
    - "build/**"
    - "*.lock"
    - "src/generated/**"
    - ".env*"
```

For a Go/CLI project:

```yaml
# .ralphy/config.yaml
project:
  name: "agent-smith"
  language: "Go"
  framework: "cobra"
  description: "Component management for AI assistants"

commands:
  test: "go test -v -race ./..."
  lint: "golangci-lint run"
  build: "go build -o bin/agent-smith"

rules:
  - "Use Go 1.23+ features"
  - "Follow the cobra command structure"
  - "All commands need help text and examples"
  - "Functions must have GoDoc comments"
  - "Use table-driven tests"
  - "Error handling must be explicit"
  - "Follow the config validation pattern"

boundaries:
  never_touch:
    - "vendor/**"
    - "bin/**"
    - "go.sum"
    - ".git/**"
```

## See Also

- [README.md](README.md) - Main documentation
- [TESTING.md](TESTING.md) - Testing guide
- [pkg/config/config.go](pkg/config/config.go) - Configuration implementation
- [Ralphy Documentation](https://github.com/michaelshimeles/ralphy) - AI assistant framework
