# Agent Smith Profile Builder - Quick Start Guide

## What Changed?

The agent-smith-profile-builder skill now correctly uses agent-smith's `profile copy` command to copy components from existing profiles in `~/.agent-smith/profiles/` instead of a non-existent base directory.

## How It Works

### Component Discovery
1. Scans all profiles in `~/.agent-smith/profiles/`
2. Finds skills/agents/commands matching your template keywords
3. Identifies which profile each component lives in
4. Copies components from source profile to your new profile

### Example Flow

```bash
# You want to create a Java backend profile
# The skill will:

# 1. Scan for matching skills
found: api-design-principles (in wshobson-agents)
found: sql-optimization-patterns (in anthropics-skills)

# 2. Create new profile
./agent-smith profile create java-backend-dev

# 3. Copy each component from its source
./agent-smith profile copy skills wshobson-agents java-backend-dev api-design-principles
./agent-smith profile copy skills anthropics-skills java-backend-dev sql-optimization-patterns

# 4. Activate and link
./agent-smith profile activate java-backend-dev
./agent-smith link all
```

## Available Commands

### Correct Commands (Singular)
```bash
# Create profile
./agent-smith profile create <name>

# Copy component from another profile
./agent-smith profile copy skills <source> <target> <skill-name>
./agent-smith profile copy agents <source> <target> <agent-name>

# Activate profile
./agent-smith profile activate <name>

# Show profile info
./agent-smith profile status [name]

# Add component from base directory (rarely used)
./agent-smith profile add skills <name> <skill-name>
```

### Incorrect Commands (Don't Use)
```bash
# These DON'T exist:
./agent-smith profiles create    # ❌ (profiles is plural)
./agent-smith profiles add        # ❌ (profiles is plural)
./agent-smith profiles activate   # ❌ (profiles is plural)
```

## Quick Usage

### 1. Install the skill (if not already)
```bash
cp -r skills/agent-smith-profile-builder ~/.agent-smith/skills/
```

### 2. Use in Claude/OpenCode
```
"Use the agent-smith-profile-builder skill to create a Java backend profile"
```

### 3. Follow the prompts
- Choose "Quick Start" for predefined templates
- Choose "Custom Build" for tailored profiles
- Choose "Update Existing" to enhance a profile

## Available Templates

1. **java-backend** - Java + Spring Boot development
2. **python-ml** - Python ML/Data Science
3. **react-frontend** - React frontend development
4. **nodejs-fullstack** - Node.js full-stack
5. **mobile-react-native** - React Native mobile
6. **devops-platform** - DevOps/Platform engineering

## Component Scanner

The component scanner is the heart of the skill:

```bash
# List all available components
~/.agent-smith/skills/agent-smith-profile-builder/lib/component-scanner.sh list-all

# Find which profile has a skill
~/.agent-smith/skills/agent-smith-profile-builder/lib/component-scanner.sh find-profiles-with-skill "api-design-principles"

# Get agents in a category
~/.agent-smith/skills/agent-smith-profile-builder/lib/component-scanner.sh get-agents-in-category "backend-development"

# List all skills
~/.agent-smith/skills/agent-smith-profile-builder/lib/component-scanner.sh scan-skills
```

## Troubleshooting

### "Skill not found in any profile"
- This is normal - not all keyword patterns will match
- The skill gracefully skips missing components
- You can add more profiles to increase component availability

### "profile command not found"
- Build agent-smith first: `cd /Users/tgaines/dev/git/agent-smith && go build -o agent-smith .`
- Run from agent-smith directory or add to PATH

### "No matching skills found"
- Your installed profiles may not have components matching the template
- Try installing more profiles from the registry
- Use Custom Build workflow to specify different keywords

## Advanced Usage

### Prefer Active Profile as Source
When a component exists in multiple profiles, it uses the first alphabetically. To prefer your active profile:

```bash
# Activate your preferred source profile first
./agent-smith profile activate my-preferred-profile

# Then create new profile
# Modify component-scanner.sh to check active profile first
```

### Show Source Profiles
To see which profile each component comes from, check the SKILL.md recommendations section - you could enhance it to show source info.

### Manual Component Selection
If you want fine-grained control:
1. Run component scanner to list all available
2. Manually use `profile copy` commands
3. Build your profile exactly how you want

## Examples

### Create Java Backend Profile
```
User: "Create a Java backend profile named 'spring-api-dev'"

AI will:
- Load java-backend template
- Scan for matching skills (api-design, sql-optimization, etc.)
- Show recommendations
- Create profile with selected components
- Activate and link
```

### Update Existing Profile
```
User: "Update my java-api profile with new testing skills"

AI will:
- List your profiles
- You select java-api
- Scan for testing-related skills not already in profile
- Add selected new skills
```

## Files Modified

- `SKILL.md` - Core instructions (profile copy logic)
- `README.md` - Documentation examples
- `templates/java-backend.yaml` - Customization commands
- `CHANGELOG.md` - Detailed change log (new)
- `QUICK-START.md` - This file (new)

## Resources

- Full documentation: `skills/agent-smith-profile-builder/README.md`
- Detailed instructions: `skills/agent-smith-profile-builder/SKILL.md`
- Change history: `skills/agent-smith-profile-builder/CHANGELOG.md`
- Templates: `skills/agent-smith-profile-builder/templates/*.yaml`
- Scanner: `skills/agent-smith-profile-builder/lib/component-scanner.sh`
