# Understanding Profiles

agent-smith organizes components using profiles:

- **📦 Repository Profiles**: Auto-created from `install all <repo>`, tied to source
- **👤 User Profiles**: Custom collections you create manually
- **⊙ Base Installation**: Default location (no profile active)

## Quick Profile Commands

```bash
# View all profiles
agent-smith profile list

# Activate a profile
agent-smith profile activate <profile-name>

# Check what's active
agent-smith status
```

## Profile Selection Workflow

When installing a skill, follow this workflow to determine where to install it:

### 1. Check Available Profiles

Run one of these commands:
```bash
agent-smith status
```
or
```bash
agent-smith profile list
```

### 2. Ask the User Using the Question Tool

Use the `question` tool with these options:
- **Header:** "Installation Profile"
- **Question:** "Which profile should this skill be installed to?"
- **Options:**
  1. "Active profile: `<profile-name>`" (only if a profile is active) - **Recommended**
  2. "Base installation (no profile)" - Available globally
  3. "Choose a different profile"

### 3. Handle the User's Choice

**Option 1 - Active profile (or Option 2 - Base installation):**
```bash
agent-smith install skill <owner/repo> <skill-name>
```

**Option 3 - Different profile:**
- Parse the output from `agent-smith profile list`
- Ask a follow-up question with all available profiles
- Install with:
```bash
agent-smith install skill <owner/repo> <skill-name> --profile <chosen-profile>
```
