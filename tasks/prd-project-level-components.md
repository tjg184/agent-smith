# PRD: Project-Level Component Management

## Introduction

Implement project-level component management to allow users to define dependencies in a manifest file (`agent-smith.json`), install them locally to the project, and automatically generate an XML-based `agents.md` file for AI context. This moves `agent-smith` closer to an `npm`-like experience for AI tools.

## Goals

- Enable per-project dependency definition via `agent-smith.json`.
- Support local installation of skills, agents, and commands into a project-local `.agents` directory.
- Automatically generate a structured `agents.md` file in XML format to expose these tools to LLMs.
- Simplify the CLI experience with a bare `agent-smith install` command.

## User Stories

- [ ] Story-001: As a developer, I want to define my project's AI tool dependencies in an `agent-smith.json` file so that I can share the setup with my team.
  
  **Acceptance Criteria:**
  - Create `pkg/manifest` to handle `agent-smith.json` parsing and saving.
  - Support `dependencies` (skills, agents, commands) and `config` (installDir, docsFile) sections.
  - Default `installDir` to `.agents` and `docsFile` to `agents.md`.
  - Validate the JSON structure on load.

  **Testing Criteria:**
  **Unit Tests:**
  - Test valid/invalid JSON parsing.
  - Test default value application.
  - Test serialization (save) logic.

- [ ] Story-002: As a developer, I want `agent-smith install` to download the dependencies to a local directory so that they are isolated to this project.

  **Acceptance Criteria:**
  - Refactor `SkillDownloader`, `AgentDownloader`, and `CommandDownloader` to accept a custom `baseDir`.
  - Implement `handleInstallManifest` in `main.go`.
  - Iterate through manifest dependencies and download them to the configured `installDir`.
  - **Overwrite** existing components in the local directory (no skipping/prompting).
  - Ensure lockfiles (if generated) are stored locally or ignored if not needed for this mode.

  **Testing Criteria:**
  **Unit Tests:**
  - Test new Downloader constructors with custom paths.
  **Integration Tests:**
  - Mock a manifest and verify files appear in the correct local folder.
  - Verify existing files are overwritten.

- [ ] Story-003: As a developer, I want an `agents.md` file generated automatically after install so that my AI agent knows how to use these tools.

  **Acceptance Criteria:**
  - Create `internal/generator` package.
  - Scan the `installDir` for installed components.
  - Generate XML output matching strict formats for each type:
    
    **Skills:**
    ```xml
    <skills_system priority="1">
      <usage>...</usage>
      <available_skills>
        <skill>
          <name>...</name>
          <description>...</description>
          <location>project</location>
        </skill>
      </available_skills>
    </skills_system>
    ```

    **Agents:**
    ```xml
    <agents_system priority="1">
      <usage>...</usage>
      <available_agents>
        <agent>
          <name>...</name>
          <description>...</description>
          <location>project</location>
        </agent>
      </available_agents>
    </agents_system>
    ```

    **Commands:**
    ```xml
    <commands_system priority="1">
      <usage>...</usage>
      <available_commands>
        <command>
          <name>...</name>
          <description>...</description>
          <location>project</location>
        </command>
      </available_commands>
    </commands_system>
    ```
  - Extract descriptions from `SKILL.md`/`AGENT.md` frontmatter or content (heuristic).
  - Write this content to the configured `docsFile` (default `agents.md`).

  **Testing Criteria:**
  **Unit Tests:**
  - Test XML generation string formatting for all 3 types.
  - Test metadata extraction from component files.
  **Integration Tests:**
  - Verify `agents.md` is created/updated after `install`.

- [ ] Story-004: As a user, I want to run `agent-smith install` without arguments to trigger this workflow.

  **Acceptance Criteria:**
  - Update `cmd/root.go` to allow 0 arguments for the `install` command.
  - Detect `agent-smith.json` in the current working directory.
  - Trigger the manifest installation flow if the file exists and no args are provided.
  - Provide clear CLI output (success/failure) for the batch operation.

  **Testing Criteria:**
  **Component Browser Tests:**
  - CLI manual testing: Run `agent-smith install` in a folder with/without the manifest.

## Functional Requirements

- FR-1: The system must support `agent-smith.json` with `dependencies` and `config`.
- FR-2: The system must allow configuring the installation directory (default `.agents`).
- FR-3: The system must allow configuring the output documentation file (default `agents.md`).
- FR-4: The system must overwrite existing components during project-level install.
- FR-5: The generated documentation must be in strict XML format, distinct for each component type (skills, agents, commands).
- FR-6: The `install` command must auto-detect the manifest in the CWD.

## Non-Goals

- No semantic versioning or complex lockfile logic for this iteration (simple URL list only).
- No "skip if exists" logic; always overwrite.
- No support for documentation formats other than the specified XML.
- No "uninstall" or "prune" command for the manifest yet.

