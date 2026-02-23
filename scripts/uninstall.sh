#!/usr/bin/env bash
#
# uninstall.sh - Agent Smith uninstaller
# Removes the agent-smith binary and optionally removes all data
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/uninstall.sh | bash
#   curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/uninstall.sh | bash -s -- --purge

set -e

# Configuration
INSTALL_DIR="$HOME/.agent-smith/bin"
DATA_DIR="$HOME/.agent-smith"
BINARY_NAME="agent-smith"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

#######################################
# Print colored message
#######################################
print_info() {
    echo -e "${BLUE}$1${NC}"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

#######################################
# Show usage information
#######################################
show_help() {
    cat << EOF
Agent Smith Uninstaller

Usage:
  curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/uninstall.sh | bash
  curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/uninstall.sh | bash -s -- [OPTIONS]

Options:
  --purge           Remove all data including skills, agents, and profiles
                    without prompting
  --help            Show this help message

Examples:
  # Remove binary only (keeps data)
  curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/uninstall.sh | bash

  # Remove everything
  curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/uninstall.sh | bash -s -- --purge

EOF
}

#######################################
# Remove binary
#######################################
remove_binary() {
    local binary_path="${INSTALL_DIR}/${BINARY_NAME}"
    
    if [ ! -f "$binary_path" ]; then
        print_warning "${BINARY_NAME} is not installed at ${binary_path}"
        return 1
    fi
    
    print_info "Removing ${BINARY_NAME} binary..."
    rm -f "$binary_path"
    
    # Remove bin directory if empty
    if [ -d "$INSTALL_DIR" ] && [ -z "$(ls -A "$INSTALL_DIR")" ]; then
        rmdir "$INSTALL_DIR"
    fi
    
    print_success "${BINARY_NAME} binary removed"
    return 0
}

#######################################
# Prompt for data removal
#######################################
prompt_data_removal() {
    local purge=$1
    
    # Check if data directory exists
    if [ ! -d "$DATA_DIR" ]; then
        return 1
    fi
    
    # If --purge flag, skip prompt
    if [ "$purge" = "true" ]; then
        return 0
    fi
    
    echo ""
    print_warning "Do you want to remove all agent-smith data?"
    echo "This includes installed skills, agents, commands, and profiles."
    echo "Location: ${DATA_DIR}"
    echo -n "[y/N]: "
    read -r response
    
    case "$response" in
        [yY][eE][sS]|[yY])
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

#######################################
# Remove data directory
#######################################
remove_data() {
    if [ ! -d "$DATA_DIR" ]; then
        return
    fi
    
    print_info "Removing data directory..."
    rm -rf "$DATA_DIR"
    print_success "Data directory removed"
}

#######################################
# Print PATH cleanup instructions
#######################################
print_path_instructions() {
    echo ""
    print_info "Don't forget to remove ${INSTALL_DIR} from your PATH"
    echo ""
    echo "Edit your shell config file and remove the line:"
    echo "  export PATH=\"\$HOME/.agent-smith/bin:\$PATH\""
    echo ""
    echo "Files to check:"
    echo "  - ~/.bashrc"
    echo "  - ~/.zshrc"
    echo "  - ~/.config/fish/config.fish"
    echo ""
}

#######################################
# Print success summary
#######################################
print_summary() {
    local binary_removed=$1
    local data_removed=$2
    
    echo ""
    print_success "${BINARY_NAME} uninstalled successfully"
    echo ""
    
    if [ "$binary_removed" = "true" ] || [ "$data_removed" = "true" ]; then
        echo "Removed:"
        [ "$binary_removed" = "true" ] && echo "  - Binary: ${INSTALL_DIR}/${BINARY_NAME}"
        [ "$data_removed" = "true" ] && echo "  - Data: ${DATA_DIR} (skills, agents, profiles)"
        echo ""
    fi
    
    if [ "$data_removed" != "true" ] && [ -d "$DATA_DIR" ]; then
        echo "Kept:"
        echo "  - Data: ${DATA_DIR} (skills, agents, profiles)"
        echo ""
        echo "To remove all data, run with --purge flag"
        echo ""
    fi
}

#######################################
# Parse command line arguments
#######################################
parse_args() {
    PURGE=false
    
    while [ $# -gt 0 ]; do
        case "$1" in
            --purge)
                PURGE=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
}

#######################################
# Main uninstall flow
#######################################
main() {
    # Parse arguments
    parse_args "$@"
    
    # Track what was removed
    BINARY_REMOVED=false
    DATA_REMOVED=false
    
    # Remove binary
    if remove_binary; then
        BINARY_REMOVED=true
    fi
    
    # Handle data removal
    if prompt_data_removal "$PURGE"; then
        remove_data
        DATA_REMOVED=true
    fi
    
    # Print PATH instructions (only if binary was removed and data wasn't)
    if [ "$BINARY_REMOVED" = "true" ] && [ "$DATA_REMOVED" != "true" ]; then
        print_path_instructions
    fi
    
    # Print summary
    print_summary "$BINARY_REMOVED" "$DATA_REMOVED"
}

# Run main function
main "$@"
