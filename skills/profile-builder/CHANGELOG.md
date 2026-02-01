# Profile Builder Skill - Changelog

## 2024-02-01 - Profile Copy Command Update

### Summary
Updated the profile-builder skill to correctly use agent-smith's `profile copy` command instead of the incorrect `profiles add` command. The skill now properly copies components from existing profiles in `~/.agent-smith/profiles/` rather than attempting to copy from a non-existent base directory.

### Changes Made

#### 1. SKILL.md Updates
- **Step A6 (Profile Creation)**: Updated to use `profile copy` command
  - Changed from: `./agent-smith profiles add skill`
  - Changed to: `./agent-smith profile copy skills "$source_profile" <profile-name> "$skill"`
  - Now uses component-scanner's `find-profiles-with-skill/agent/command` functions
  
- **Command Syntax Corrections**: Fixed all instances of `profiles` (plural) to `profile` (singular)
  - Lines 267, 390-392, 352, 358, 361, 415, 418
  - Commands: create, add, remove, copy, activate, deactivate, show

- **Documentation Enhancement**: Added explanation of why `profile copy` is used
  - Preserves component metadata from source profile
  - Copies lock file entries for provenance tracking
  - Creates independent copies
  - Notes that `profile add` copies from base dir (not applicable for profiles architecture)

#### 2. README.md Updates
- Updated example commands from `profiles` to `profile` (singular)
  - Lines 188, 192, 201: Profile creation and activation examples
  - Lines 232-242: Test profile creation walkthrough
  - All command syntax now matches actual agent-smith CLI

#### 3. Template YAML Updates
- **java-backend.yaml**: Updated readme_template customization examples
  - Changed: `agent-smith profiles add/remove` → `agent-smith profile add/remove`
  - Updated both skills and agents command examples

### Technical Details

#### Component Discovery Flow
The updated workflow now correctly:

1. **Scans all profiles** in `~/.agent-smith/profiles/` for available components
2. **Finds source profile** for each matched component using scanner functions:
   ```bash
   source_profile=$($SCANNER find-profiles-with-skill "$skill" | head -1)
   ```
3. **Copies from source profile** to new profile using:
   ```bash
   ./agent-smith profile copy skills "$source_profile" <new-profile> "$skill"
   ```

#### Component Scanner Functions Used
- `find-profiles-with-skill <skill-name>` - Returns profiles containing a skill
- `find-profiles-with-agent <agent-path>` - Returns profiles containing an agent
- `find-profiles-with-command <command-name>` - Returns profiles containing a command
- `scan-skills` - Lists all unique skills across all profiles
- `scan-agents` - Lists all unique agents across all profiles
- `scan-commands` - Lists all unique commands across all profiles

### Verification

#### Current System State
- **Skills available**: 162 (across all profiles)
- **Agents available**: 108 (across all profiles)
- **Commands available**: 41 (across all profiles)
- **Profiles**: anthropics-skills, vercel-labs-agent-skills, wshobson-agents

#### Tested Functions
✅ Component scanner lists all components correctly
✅ `find-profiles-with-skill` returns correct source profiles
✅ `get-agents-in-category` returns agents from categories
✅ All command syntax now uses `profile` (singular)
✅ Profile copy logic uses source profile detection

### Benefits

1. **Architecturally Correct**: Now works with actual profiles structure
2. **Preserves Metadata**: Uses `profile copy` which maintains lock files and provenance
3. **Multi-Profile Support**: Automatically discovers components across all installed profiles
4. **Error Handling**: Gracefully handles components not found in any profile
5. **Future-Proof**: Works with any profiles users install

### Breaking Changes
None - the skill instructions are self-contained and guide users through the correct workflow.

### Files Modified
- `skills/profile-builder/SKILL.md` - Core skill instructions
- `skills/profile-builder/README.md` - Documentation and examples
- `skills/profile-builder/templates/java-backend.yaml` - Template customization examples

### Files Unchanged
- `skills/profile-builder/lib/component-scanner.sh` - Already correctly scanned profiles
- `skills/profile-builder/templates/python-ml.yaml` - No command examples
- `skills/profile-builder/templates/react-frontend.yaml` - No command examples
- `skills/profile-builder/templates/nodejs-fullstack.yaml` - No command examples
- `skills/profile-builder/templates/mobile-react-native.yaml` - No command examples
- `skills/profile-builder/templates/devops-platform.yaml` - No command examples

### Testing Recommendations

When using the updated skill:
1. Ensure agent-smith binary is built: `cd /Users/tgaines/dev/git/agent-smith && go build -o agent-smith .`
2. Verify profiles exist: `ls ~/.agent-smith/profiles/`
3. Test component scanner: `skills/profile-builder/lib/component-scanner.sh list-all`
4. Follow SKILL.md Workflow A for template-based profile creation
5. Verify new profile contains copied components: `./agent-smith profile show <profile-name>`

### Implementation Notes

#### Source Profile Selection Strategy
When a component exists in multiple profiles, we use `head -1` to select the first match (alphabetically by profile name). This simple strategy works well because:
- Component content should be identical across profiles (same skill/agent/command)
- Lock files track the original source, so provenance is maintained
- Users can manually adjust after creation if needed

#### Future Enhancements (Optional)
- Could add preference for active profile as source
- Could show source profile name in recommendations (informational)
- Could allow user to select preferred source when multiple exist
- Could add validation that source and target profiles are different

### Command Reference

#### Old (Incorrect) Commands
```bash
./agent-smith profiles create <name>          # profiles (plural) doesn't exist
./agent-smith profiles add skill <name> <skill>  # profiles (plural) doesn't exist
./agent-smith profiles activate <name>        # profiles (plural) doesn't exist
```

#### New (Correct) Commands
```bash
./agent-smith profile create <name>                          # profile (singular)
./agent-smith profile copy skills <source> <target> <skill>  # copies from profile
./agent-smith profile activate <name>                        # profile (singular)
./agent-smith profile add skills <name> <skill>              # adds from base dir (not used)
```
