#!/bin/bash

# Component Scanner for Agent-Smith Profile Builder
# Scans ~/.agent-smith/profiles/ directory for available components

AGENT_SMITH_DIR="$HOME/.agent-smith"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to scan available skills
scan_skills() {
    local profiles_dir="$AGENT_SMITH_DIR/profiles"
    
    # Check if profiles directory exists
    if [[ ! -d "$profiles_dir" ]]; then
        return 0
    fi
    
    # Scan all profiles for skills and deduplicate
    # Skills are at: profiles/profile-name/skills/skill-name
    find "$profiles_dir" -mindepth 3 -maxdepth 3 -type d -path "*/skills/*" -exec basename {} \; | sort -u
}

# Function to scan available agents
scan_agents() {
    local profiles_dir="$AGENT_SMITH_DIR/profiles"
    
    # Check if profiles directory exists
    if [[ ! -d "$profiles_dir" ]]; then
        return 0
    fi
    
    # Scan all profiles for agents and deduplicate
    find "$profiles_dir" -type f -path "*/agents/*.md" | while read -r agent; do
        # Get relative path from agents directory within the profile
        # Extract everything after "agents/" and remove .md extension
        rel_path=$(echo "$agent" | sed 's|.*/agents/||' | sed 's/\.md$//')
        echo "$rel_path"
    done | sort -u
}

# Function to scan available agents by category
scan_agents_by_category() {
    local profiles_dir="$AGENT_SMITH_DIR/profiles"
    
    # Check if profiles directory exists
    if [[ ! -d "$profiles_dir" ]]; then
        return 0
    fi
    
    # Scan all profiles for agent categories (directories within agents/) and deduplicate
    # Agent categories are at: profiles/profile-name/agents/category-name
    find "$profiles_dir" -mindepth 3 -maxdepth 3 -type d -path "*/agents/*" -exec basename {} \; | sort -u
}

# Function to scan available commands
scan_commands() {
    local profiles_dir="$AGENT_SMITH_DIR/profiles"
    
    # Check if profiles directory exists
    if [[ ! -d "$profiles_dir" ]]; then
        return 0
    fi
    
    # Scan all profiles for commands and deduplicate
    # Commands are at: profiles/profile-name/commands/command-name
    find "$profiles_dir" -mindepth 3 -maxdepth 3 -type d -path "*/commands/*" -exec basename {} \; | sort -u
}

# Function to check if a skill exists
skill_exists() {
    local skill_name="$1"
    local profiles_dir="$AGENT_SMITH_DIR/profiles"
    
    # Check if profiles directory exists
    if [[ ! -d "$profiles_dir" ]]; then
        return 1
    fi
    
    # Check if skill exists in any profile
    # Skills are at: profiles/profile-name/skills/skill-name
    [[ -n $(find "$profiles_dir" -mindepth 3 -maxdepth 3 -type d -path "*/skills/$skill_name" | head -1) ]]
}

# Function to check if an agent exists
agent_exists() {
    local agent_path="$1"
    local profiles_dir="$AGENT_SMITH_DIR/profiles"
    
    # Check if profiles directory exists
    if [[ ! -d "$profiles_dir" ]]; then
        return 1
    fi
    
    # Check if agent exists in any profile
    [[ -n $(find "$profiles_dir" -type f -path "*/agents/${agent_path}.md" | head -1) ]]
}

# Function to check if a command exists
command_exists() {
    local command_name="$1"
    local profiles_dir="$AGENT_SMITH_DIR/profiles"
    
    # Check if profiles directory exists
    if [[ ! -d "$profiles_dir" ]]; then
        return 1
    fi
    
    # Check if command exists in any profile
    # Commands are at: profiles/profile-name/commands/command-name
    [[ -n $(find "$profiles_dir" -mindepth 3 -maxdepth 3 -type d -path "*/commands/$command_name" | head -1) ]]
}

# Function to get agents in a category
get_agents_in_category() {
    local category="$1"
    local profiles_dir="$AGENT_SMITH_DIR/profiles"
    
    # Check if profiles directory exists
    if [[ ! -d "$profiles_dir" ]]; then
        return 0
    fi
    
    # Find all agents in this category across all profiles and deduplicate
    find "$profiles_dir" -type f -path "*/agents/$category/*.md" -exec basename {} .md \; | sort -u
}

# Function to find matching skills by pattern
find_matching_skills() {
    local pattern="$1"
    scan_skills | grep -i "$pattern"
}

# Function to find matching agents by pattern
find_matching_agents() {
    local pattern="$1"
    scan_agents | grep -i "$pattern"
}

# Function to validate component availability
validate_components() {
    local component_type="$1"
    shift
    local components=("$@")
    
    local available=()
    local unavailable=()
    
    for comp in "${components[@]}"; do
        case "$component_type" in
            "skill")
                if skill_exists "$comp"; then
                    available+=("$comp")
                else
                    unavailable+=("$comp")
                fi
                ;;
            "agent")
                if agent_exists "$comp"; then
                    available+=("$comp")
                else
                    unavailable+=("$comp")
                fi
                ;;
            "command")
                if command_exists "$comp"; then
                    available+=("$comp")
                else
                    unavailable+=("$comp")
                fi
                ;;
        esac
    done
    
    # Output as JSON-like format for easy parsing
    echo "AVAILABLE:"
    printf '%s\n' "${available[@]}"
    echo "UNAVAILABLE:"
    printf '%s\n' "${unavailable[@]}"
}

# Function to list all components with counts
list_all_components() {
    local skill_count=$(scan_skills | wc -l | tr -d ' ')
    local agent_count=$(scan_agents | wc -l | tr -d ' ')
    local command_count=$(scan_commands | wc -l | tr -d ' ')
    
    echo -e "${GREEN}Available Components in ~/.agent-smith/profiles/${NC}"
    echo -e "${BLUE}Skills:${NC} $skill_count"
    echo -e "${BLUE}Agents:${NC} $agent_count"
    echo -e "${BLUE}Commands:${NC} $command_count"
}

# Main CLI interface
case "${1:-}" in
    "scan-skills")
        scan_skills
        ;;
    "scan-agents")
        scan_agents
        ;;
    "scan-agents-categories")
        scan_agents_by_category
        ;;
    "scan-commands")
        scan_commands
        ;;
    "skill-exists")
        skill_exists "$2" && echo "true" || echo "false"
        ;;
    "agent-exists")
        agent_exists "$2" && echo "true" || echo "false"
        ;;
    "command-exists")
        command_exists "$2" && echo "true" || echo "false"
        ;;
    "get-agents-in-category")
        get_agents_in_category "$2"
        ;;
    "find-skills")
        find_matching_skills "$2"
        ;;
    "find-agents")
        find_matching_agents "$2"
        ;;
    "validate")
        shift
        component_type="$1"
        shift
        validate_components "$component_type" "$@"
        ;;
    "list-all")
        list_all_components
        ;;
    *)
        echo "Usage: $0 {scan-skills|scan-agents|scan-agents-categories|scan-commands|skill-exists|agent-exists|command-exists|get-agents-in-category|find-skills|find-agents|validate|list-all}"
        exit 1
        ;;
esac
