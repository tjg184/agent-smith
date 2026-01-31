---
name: profile-builder
description: Build agent-smith profiles tailored to your development needs. Dynamically discovers and recommends relevant skills, agents, and commands based on your language, framework, and focus areas. Supports both quick-start templates and custom profile creation.
---

# Profile Builder

Intelligently build agent-smith profiles by analyzing available components in `~/.agent-smith/` and recommending the best fit for your development needs.

## When to Use This Skill

- Creating a new agent-smith profile for a specific tech stack
- Setting up a Java/Spring Boot development environment
- Building profiles for Python ML, React, Node.js, mobile, or DevOps
- Updating existing profiles with newly installed skills
- Customizing profiles for specific project needs

## How It Works

This skill uses **keyword-based pattern matching** to dynamically discover relevant components. Instead of hardcoding skill names (which can change), it:

1. Reads template configurations with keyword patterns
2. Scans `~/.agent-smith/skills/` for matching names
3. Ranks by priority (required > recommended > optional)
4. Presents recommendations for user approval

This ensures the profile builder works with any skills installed in your system.

## Available Workflows

### 1. Quick Start (Template-Based)
- Select from 6 predefined templates
- Automatically discovers matching components
- Fast setup for common tech stacks

### 2. Custom Builder (Question-Based)
- Answer questions about your needs
- Generates custom keyword patterns
- Fully tailored to your requirements

### 3. Update Existing Profile
- Add newly installed skills to existing profiles
- Re-scan based on template patterns
- Manually add specific components

## Available Templates

1. **java-backend**: Java + Spring Boot (API, SQL, microservices)
2. **python-ml**: Python ML/Data Science (MLOps, LLM, pipelines)
3. **react-frontend**: React frontend (UI, TypeScript, testing)
4. **nodejs-fullstack**: Node.js full-stack (API + frontend)
5. **mobile-react-native**: React Native mobile development
6. **devops-platform**: DevOps/Platform (K8s, IaC, observability)

## Instructions

When the user wants to create or update a profile, follow these steps:

### STEP 1: Determine Workflow

Ask the user to choose:
```
What would you like to do?

1. Quick Start - Use a predefined template (recommended)
2. Custom Build - Answer questions to build a tailored profile
3. Update Existing - Add components to an existing profile
```

---

### WORKFLOW A: Quick Start

#### A1. Show Available Templates

List templates with descriptions:
```
Available Templates:

1. java-backend
   Java Backend Engineer (Spring Boot)
   → API development, SQL databases, microservices, testing

2. python-ml
   Python ML/Data Science Engineer
   → ML pipelines, LLM apps, MLOps, data engineering

3. react-frontend
   React Frontend Engineer
   → Modern React, TypeScript, UI design, accessibility

4. nodejs-fullstack
   Node.js Full-Stack Engineer
   → Backend APIs + React frontend, full-stack development

5. mobile-react-native
   Mobile Engineer (React Native)
   → Cross-platform mobile apps with React Native

6. devops-platform
   DevOps/Platform Engineer
   → Kubernetes, IaC, CI/CD, monitoring, SRE

Which template would you like to use? (1-6)
```

#### A2. Get Profile Name

Ask for profile name and validate:
```
What would you like to name this profile?

Examples: "java-api-dev", "spring-microservices", "ml-research"

Note: Use only letters, numbers, and dashes (agent-smith requirement)
```

Validate with regex: `^[a-zA-Z0-9-]+$`

#### A3. Load Template and Scan Components

1. **Read the selected template** from `skills/profile-builder/templates/<template>.yaml`

2. **Scan for matching skills** using the component scanner:

```bash
# Get base directory for the skill
SKILL_DIR="$(cd "$(dirname "$0")" && pwd)"
SCANNER="$SKILL_DIR/lib/component-scanner.sh"

# Scan all available skills
available_skills=$($SCANNER scan-skills)

# Match skills based on template keywords
matched_skills=()

# Example for java-backend template:
# Match skills containing: api-design, api.*principle, architecture.*pattern, etc.
for skill in $available_skills; do
  # Check against each keyword pattern from template
  # Use grep -iE for case-insensitive regex matching
done
```

3. **Scan for matching agents**:

```bash
# Get agents from specified categories
# Example: backend-development, database-design, api-testing-observability

for category in "${agent_categories[@]}"; do
  agents=$($SCANNER get-agents-in-category "$category")
  matched_agents+=($agents)
done
```

4. **Scan for matching commands**:

```bash
# Match command patterns (e.g., *java*, *spring*, *maven*)
all_commands=$($SCANNER scan-commands)

for pattern in "${command_patterns[@]}"; do
  # Use glob matching
  for cmd in $all_commands; do
    if [[ $cmd == $pattern ]]; then
      matched_commands+=("$cmd")
    fi
  done
done
```

#### A4. Show Recommendations

Present findings to user:

```
Profile: <profile-name>
Template: <template-display-name>

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SKILLS TO INCLUDE (12 found)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Core Backend (required):
  ✓ api-design-principles
  ✓ architecture-patterns
  ✓ error-handling-patterns
  ✓ auth-implementation-patterns

Database (required):
  ✓ sql-optimization-patterns
  ✓ database-migration
  ✓ postgresql-table-design

Testing (recommended):
  ○ e2e-testing-patterns
  ○ debugging-strategies

Microservices (recommended):
  ○ microservices-patterns
  ○ distributed-tracing

Observability (optional):
  ○ prometheus-configuration

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
AGENTS TO INCLUDE (5 found)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  • backend-development/api-architect
  • backend-development/tdd-orchestrator
  • database-design/sql-pro
  • api-testing-observability/api-documenter
  • backend-api-security/security-auditor

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
COMMANDS (0 found)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  (No matching commands found)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

What would you like to do?
1. Continue with these components
2. Customize (add/remove components)
3. Cancel
```

#### A5. Handle Customization (if requested)

If user chooses to customize:

```
Customization Options:

1. Remove skills (uncheck items)
2. Add more skills manually
3. Remove agents
4. Add more agents manually
5. Done customizing

Which option? (1-5)
```

Allow iterative customization until user is satisfied.

#### A6. Create Profile

Execute agent-smith commands:

```bash
# Navigate to agent-smith project directory
cd /Users/tgaines/dev/git/agent-smith

# Build agent-smith if not already built
if [[ ! -f "./agent-smith" ]]; then
  go build -o agent-smith .
fi

# Create the profile
./agent-smith profiles create <profile-name>

# Add all matched skills
for skill in "${matched_skills[@]}"; do
  echo "Adding skill: $skill"
  ./agent-smith profiles add skill <profile-name> "$skill"
done

# Add all matched agents
for agent in "${matched_agents[@]}"; do
  echo "Adding agent: $agent"
  ./agent-smith profiles add agent <profile-name> "$agent"
done

# Add all matched commands
for command in "${matched_commands[@]}"; do
  echo "Adding command: $command"
  ./agent-smith profiles add command <profile-name> "$command"
done
```

#### A7. Generate README

Create README.md in the profile directory:

```bash
PROFILE_DIR="$HOME/.agent-smith/profiles/<profile-name>"

# Read readme_template from template YAML
# Replace {COMPONENT_LIST} with actual component list

cat > "$PROFILE_DIR/README.md" << 'EOF'
# <Profile Name>

<Description from template>

## Included Components

### Skills (X)
- **skill-name**: Description
...

### Agents (Y)
- **category/agent-name**: Description
...

### Commands (Z)
- **command-name**: Description
...

## Getting Started

1. This profile is ready to use
2. To activate: `agent-smith profiles activate <profile-name>`
3. To link components: `agent-smith link all`

## Customization

Add more components:
  agent-smith profiles add skill <profile-name> <skill-name>

Remove components:
  agent-smith profiles remove skill <profile-name> <skill-name>

EOF
```

#### A8. Ask About Activation

```
✓ Profile '<profile-name>' created successfully!

Summary:
  • X skills added
  • Y agents added
  • Z commands added
  • README generated

Would you like to activate this profile now? (y/n)

If yes:
  - Profile will be activated
  - All components will be linked
  - Ready to use immediately
```

If user says yes:

```bash
cd /Users/tgaines/dev/git/agent-smith

./agent-smith profiles activate <profile-name>
./agent-smith link all
./agent-smith profiles show <profile-name>
```

#### A9. Show Final Summary

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ PROFILE READY!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Profile: <profile-name>
Status: Active and linked
Location: ~/.agent-smith/profiles/<profile-name>

Components:
  • X skills
  • Y agents  
  • Z commands

To view details:
  cat ~/.agent-smith/profiles/<profile-name>/README.md

To deactivate:
  agent-smith profiles deactivate

To switch profiles:
  agent-smith profiles activate <other-profile>
```

---

### WORKFLOW B: Custom Builder

#### B1. Get Profile Name

Same as A2 above.

#### B2. Ask Primary Focus

```
What is your primary development focus?

1. Backend Development
2. Frontend Development
3. Full-Stack Development
4. Mobile Development
5. Data Science / Machine Learning
6. DevOps / Platform Engineering

Select (1-6):
```

#### B3. Ask Language/Framework (Context-Dependent)

Based on focus, ask appropriate follow-up:

**For Backend:**
```
Which backend language/framework?
1. Java / Spring Boot
2. Python
3. Node.js / Express
4. Go
5. .NET / C#
6. Ruby / Rails
7. Other

Select (1-7):
```

**For Frontend:**
```
Which frontend framework?
1. React
2. Vue.js
3. Angular
4. Svelte
5. Other

Select (1-5):
```

**For Mobile:**
```
Which mobile platform?
1. React Native
2. iOS (Swift/SwiftUI)
3. Android (Kotlin)
4. Flutter
5. Other

Select (1-5):
```

**For ML/Data Science:**
```
Primary language for ML?
1. Python (recommended)
2. R
3. Julia
4. Other

Select (1-4):
```

**For DevOps:**
Skip to next question.

#### B4. Ask Secondary Capabilities (Multi-Select)

```
Select additional areas of interest (multi-select):

1. API Design & Development
2. Database Design & Optimization
3. Testing (Unit, Integration, E2E)
4. Security & Authentication
5. Observability & Monitoring
6. CI/CD & DevOps
7. Microservices Architecture
8. Event-Driven Architecture
9. Performance Optimization
10. None / Skip

Enter your choices (comma-separated, e.g., "1, 2, 3"):
```

#### B5. Generate Custom Keywords

Based on answers, build keyword patterns:

```
Example mapping:

Backend + Java:
  - api-design, api.*principle, architecture.*pattern
  - java, spring
  - error.*handling, auth.*implementation

API Design:
  - openapi, rest, graphql

Database:
  - sql.*optimization, database.*design, postgresql

Testing:
  - testing.*pattern, e2e, debugging

Microservices:
  - microservices, distributed, service.*mesh

Security:
  - security, auth.*pattern, encryption

Observability:
  - prometheus, grafana, monitoring, tracing
```

#### B6. Scan and Match

Use same scanning logic as Workflow A, step A3.

#### B7. Show Recommendations

Same as Workflow A, step A4.

#### B8-B9. Create, Activate, Summarize

Same as Workflow A, steps A6-A9.

---

### WORKFLOW C: Update Existing Profile

#### C1. List Existing Profiles

```bash
cd /Users/tgaines/dev/git/agent-smith

# List profiles
profiles=$(ls ~/.agent-smith/profiles/)

echo "Existing profiles:"
echo "$profiles" | nl
```

#### C2. Select Profile to Update

```
Which profile would you like to update?

Enter profile name:
```

#### C3. Choose Update Strategy

```
How would you like to update '<profile-name>'?

1. Scan for new matching skills (based on profile template)
2. Add specific skills/agents manually
3. Remove components
4. Cancel

Select (1-4):
```

#### C4. Execute Update

**Option 1: Scan for new skills**
- Detect which template the profile was created from (check README)
- Re-run keyword matching
- Show only new skills not already in profile
- Add selected skills

**Option 2: Add manually**
- List all available skills not in profile
- User selects which to add
- Execute add commands

**Option 3: Remove**
- List current profile components
- User selects which to remove
- Execute remove commands

---

## Template Format Reference

Templates use this YAML structure:

```yaml
name: template-id
display_name: "Human Readable Name"
description: "Description"

skill_keywords:
  category_name:
    priority: required|recommended|optional
    keywords:
      - regex-pattern-1
      - regex-pattern-2
    description: "Category description"

agent_categories:
  primary:
    - category-name-1
  secondary:
    - category-name-2

command_patterns:
  - glob-pattern-1
  - glob-pattern-2

config:
  auto_link: true|false
  create_readme: true|false
  max_skills: 25
  min_skills: 10

readme_template: |
  Markdown content with {COMPONENT_LIST} placeholder
```

## Error Handling

- **Profile name validation fails**: Show error, ask again
- **Profile already exists**: Offer to update or choose different name
- **No matching skills found**: Suggest installing more skills from agent-smith registry
- **agent-smith binary not found**: Guide user to build with `go build`
- **Component doesn't exist**: Skip gracefully, continue with available components
- **Template file not found**: List available templates, ask again

## Important Notes

1. **Always use the component scanner** - Don't hardcode skill names
2. **Show what will be added** before creating - Give user control
3. **Generate README** - Helps users understand their profile
4. **Validate profile names** - Must match agent-smith requirements
5. **Handle missing components gracefully** - Not all keywords will match

## Example Session

```
User: "I want to create a Java backend profile"