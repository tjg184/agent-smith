---
name: find-skills
description: Helps users discover and install agent skills using the agent-smith CLI package manager. Use when users ask "how do I do X", "find a skill for X", "is there a skill that can...", "can you do X", or express interest in extending capabilities with skills, tools, templates, or workflows for specific domains (design, testing, deployment, etc.).
---

# Find Skills

This skill helps you discover and install skills from the open agent skills ecosystem using the agent-smith CLI.

**Browse all skills:** https://skills.sh/

## Core Commands

```bash
agent-smith find skill [query]          # Search for skills
agent-smith install skill <repo> <name> # Install a skill
agent-smith status                      # Check active profile
```

## Workflow

### Step 1: Understand the User's Need

Identify the domain and specific task:
- Domain: React, testing, design, deployment, etc.
- Task: Writing tests, creating animations, reviewing PRs, etc.

### Step 2: Search for Skills

Run `agent-smith find skill [query]` with relevant keywords.

**Examples:**
- "How do I make my React app faster?" → `agent-smith find skill react performance`
- "Can you help with PR reviews?" → `agent-smith find skill pr review`
- "I need to create a changelog" → `agent-smith find skill changelog`

**For search optimization tips:** See [references/search-tips.md](references/search-tips.md)

### Step 3: Present Options

When you find relevant skills, present:
1. Skill name and what it does
2. Install command
3. Link to skills.sh for more info

**Example response:**
```
I found the "vercel-react-best-practices" skill - it provides React and Next.js 
performance optimization guidelines from Vercel Engineering.

To install:
agent-smith install skill vercel-labs/agent-skills vercel-react-best-practices

Learn more: https://skills.sh/vercel-labs/agent-skills/vercel-react-best-practices
```

### Step 4: Install the Skill

If the user wants to proceed, follow the profile selection workflow.

**For complete profile management details:** See [references/profiles.md](references/profiles.md)

**Quick workflow:**
1. Run `agent-smith status` to check active profile
2. Ask user which profile to install to (use `question` tool)
3. Install with: `agent-smith install skill <owner/repo> <skill-name>`
4. For different profile: Add `--profile <profile-name>` flag
