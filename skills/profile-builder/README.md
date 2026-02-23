# Agent Smith Profile Builder Skill - Quick Start Guide

## Overview

The Agent Smith Profile Builder skill creates tailored agent-smith profiles by dynamically discovering and recommending skills, agents, and commands from your `~/.agent-smith/` directory.

## File Structure

```
skills/agent-smith-profile-builder/
├── SKILL.md                      # Main skill logic (639 lines)
├── lib/
│   └── component-scanner.sh      # Component discovery helper
└── templates/
    ├── java-backend.yaml         # Java/Spring Boot template
    ├── python-ml.yaml            # Python ML/Data Science template
    ├── react-frontend.yaml       # React frontend template
    ├── nodejs-fullstack.yaml     # Node.js full-stack template
    ├── mobile-react-native.yaml  # React Native mobile template
    └── devops-platform.yaml      # DevOps/Platform template
```

## Key Features

### 1. Dynamic Component Discovery
- **No hardcoded skill names** - Uses regex patterns to match skills
- **Adapts to your environment** - Works with any skills installed in ~/.agent-smith/
- **Keyword-based matching** - Templates define patterns, not specific skills

### 2. Three Workflows

**Quick Start (Template-Based)**
- Choose from 6 predefined templates
- Automatically discovers matching components
- Perfect for common tech stacks

**Custom Builder (Question-Based)**
- Answer questions about your needs
- Generates custom keyword patterns
- Fully tailored profiles

**Update Existing**
- Add new skills to existing profiles
- Re-scan based on patterns
- Keep profiles up-to-date

### 3. Intelligent Recommendations
- Prioritizes components (required > recommended > optional)
- Shows what will be added before creating
- Allows customization before finalizing

## How to Use

### Example 1: Create Java Backend Profile

```
User: "I want to create a Java backend profile"

AI will:
1. Ask: Quick Start, Custom, or Update?
2. If Quick Start → Show 6 templates
3. User selects "java-backend"
4. Ask for profile name (e.g., "spring-api-dev")
5. Scan ~/.agent-smith/ for matching skills:
   - api-design-principles (matches "api.*design")
   - architecture-patterns (matches "architecture.*pattern")
   - sql-optimization-patterns (matches "sql.*optimization")
   - etc.
6. Show recommendations with categories
7. Create profile with agent-smith commands
8. Generate README.md
9. Optionally activate and link
```

### Example 2: Custom Profile

```
User: "Create a custom profile for microservices development"

AI will:
1. Ask for profile name
2. Ask primary focus (Backend)
3. Ask language (Java)
4. Ask secondary capabilities (Microservices, Testing, Observability)
5. Generate keywords:
   - microservices, distributed, service.*mesh
   - api-design, architecture
   - testing.*pattern, e2e
   - prometheus, grafana, monitoring
6. Scan and match
7. Show recommendations
8. Create profile
```

## Template Format

Templates use keyword-based discovery:

```yaml
name: java-backend
display_name: "Java Backend Engineer (Spring Boot)"

skill_keywords:
  core:
    priority: required
    keywords:
      - api-design
      - api.*principle        # Regex: matches "api-design-principles"
      - architecture.*pattern  # Regex: matches "architecture-patterns"
    description: "Core backend skills"
  
  database:
    priority: required
    keywords:
      - sql.*optimization     # Matches "sql-optimization-patterns"
      - database.*migration   # Matches "database-migration"
    description: "Database skills"

agent_categories:
  primary:
    - backend-development     # Gets all agents in this category
    - database-design

config:
  auto_link: true
  create_readme: true
  max_skills: 25
```

## Component Scanner

The `component-scanner.sh` provides these functions:

```bash
# List all skills
./lib/component-scanner.sh scan-skills

# List all agent categories
./lib/component-scanner.sh scan-agents-categories

# Get agents in a category
./lib/component-scanner.sh get-agents-in-category backend-development

# Check if skill exists
./lib/component-scanner.sh skill-exists api-design-principles

# List all components with counts
./lib/component-scanner.sh list-all
```

## Testing

Verified functionality:

✓ Component scanner lists 147 skills, 107 agents, 40 commands
✓ Keyword matching works (e.g., "api.*design" → api-design-principles)
✓ Agent category scanning works (backend-development → 3 agents)
✓ Database skill matching works (sql.*optimization → sql-optimization-patterns)

## How It Works Internally

### Phase 1: Template Loading
1. Read template YAML file
2. Parse skill_keywords sections
3. Extract agent_categories and command_patterns

### Phase 2: Component Discovery
```bash
# For each keyword category in template:
for keyword in ${keywords[@]}; do
  # Scan all skills
  all_skills=$(scanner scan-skills)
  
  # Match with regex
  matched=$(echo "$all_skills" | grep -iE "$keyword")
  
  # Add to results
done
```

### Phase 3: Profile Creation
```bash
# Build agent-smith if needed
cd /path/to/agent-smith
go build -o agent-smith .

# Create profile
./agent-smith profile create <name>

# Add components
for skill in "${skills[@]}"; do
  ./agent-smith profile add skills <name> "$skill"
done

# Generate README
cat > ~/.agent-smith/profiles/<name>/README.md << EOF
...
EOF

# Optionally activate
./agent-smith profile activate <name>
./agent-smith link all
```

## Next Steps

### To use this skill:

1. **Install it to ~/.agent-smith/skills/**:
   ```bash
   cp -r skills/agent-smith-profile-builder ~/.agent-smith/skills/
   ```

2. **Test the scanner**:
   ```bash
   ~/.agent-smith/skills/agent-smith-profile-builder/lib/component-scanner.sh list-all
   ```

3. **Use in OpenCode/Claude**:
   ```
   "Use the agent-smith-profile-builder skill to create a Java backend profile"
   ```

### To test profile creation manually:

```bash
# Build agent-smith
cd /path/to/agent-smith
go build -o agent-smith .

# Create a test profile
./agent-smith profile create test-java-profile

# Add some skills
./agent-smith profile add skills test-java-profile api-design-principles
./agent-smith profile add skills test-java-profile sql-optimization-patterns

# Verify
./agent-smith profile status test-java-profile

# Clean up
./agent-smith profile delete test-java-profile --force
```

## Customization

### Add a new template:

1. Create `templates/my-template.yaml`
2. Define skill_keywords with regex patterns
3. Define agent_categories
4. Define command_patterns
5. Add to available templates list in SKILL.md

### Modify existing templates:

1. Edit `templates/<template-name>.yaml`
2. Add/remove keywords to match different skills
3. Adjust priorities (required/recommended/optional)
4. Update max_skills/min_skills in config

## Benefits

1. **Future-proof**: Works with any skills added to ~/.agent-smith/
2. **Flexible**: Easy to add new templates or modify existing ones
3. **Transparent**: Shows what will be added before creating
4. **Smart**: Ranks components by priority and relevance
5. **Complete**: Generates README, activates, and links automatically

## Current Component Counts

From your ~/.agent-smith/:
- **147 skills** available
- **107 agents** available
- **40 commands** available

The agent-smith-profile-builder can mix and match these to create custom profiles!
