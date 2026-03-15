#!/usr/bin/env bash
#
# install.sh - Agent Smith installer
# Downloads and installs the correct agent-smith binary for your platform
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/install.sh | bash
#   curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/install.sh | bash -s -- v1.1.0
#   curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/install.sh | bash -s -- --force

set -e

# Configuration
REPO="tjg184/agent-smith"
INSTALL_DIR="$HOME/.agent-smith/bin"
BINARY_NAME="agent-smith"
GITHUB_API="https://api.github.com/repos/${REPO}"
GITHUB_RELEASES="https://github.com/${REPO}/releases/download"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Temp directory for downloads
TMP_DIR=""

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
# Cleanup temp files on exit
#######################################
cleanup() {
    if [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ]; then
        rm -rf "$TMP_DIR"
    fi
}

trap cleanup EXIT

#######################################
# Show usage information
#######################################
show_help() {
    cat << EOF
Agent Smith Installer

Usage:
  curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/install.sh | bash
  curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/install.sh | bash -s -- [OPTIONS] [VERSION]

Arguments:
  VERSION           Specific version to install (e.g., v1.1.0)
                    If not specified, installs the latest version

Options:
  --force           Overwrite existing installation without prompting
  --help            Show this help message

Examples:
  # Install latest version
  curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/install.sh | bash

  # Install specific version
  curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/install.sh | bash -s -- v1.1.0

  # Force reinstall
  curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/install.sh | bash -s -- --force

EOF
}

#######################################
# Detect operating system
#######################################
detect_os() {
    local os
    os="$(uname -s)"
    
    case "$os" in
        Linux*)
            echo "linux"
            ;;
        Darwin*)
            echo "darwin"
            ;;
        *)
            print_error "Unsupported operating system: $os"
            print_error "Supported platforms: Linux, macOS"
            exit 1
            ;;
    esac
}

#######################################
# Detect architecture
#######################################
detect_arch() {
    local arch
    arch="$(uname -m)"
    
    case "$arch" in
        x86_64)
            echo "amd64"
            ;;
        amd64)
            echo "amd64"
            ;;
        arm64)
            echo "arm64"
            ;;
        aarch64)
            echo "arm64"
            ;;
        *)
            print_error "Unsupported architecture: $arch"
            print_error "Supported architectures: x86_64, arm64"
            exit 1
            ;;
    esac
}

#######################################
# Check if required tools are available
#######################################
check_requirements() {
    local missing=()
    
    if ! command -v curl >/dev/null 2>&1; then
        missing+=("curl")
    fi
    
    if ! command -v tar >/dev/null 2>&1; then
        missing+=("tar")
    fi
    
    # Check for sha256sum (Linux) or shasum (macOS)
    if ! command -v sha256sum >/dev/null 2>&1 && ! command -v shasum >/dev/null 2>&1; then
        missing+=("sha256sum or shasum")
    fi
    
    if [ ${#missing[@]} -gt 0 ]; then
        print_error "Missing required tools: ${missing[*]}"
        print_error "Please install them and try again"
        exit 1
    fi
}

#######################################
# Get latest version from GitHub API
#######################################
get_latest_version() {
    local version
    
    print_info "Fetching latest version..." >&2
    
    version=$(curl -sSL "${GITHUB_API}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        print_error "Failed to fetch latest version from GitHub"
        print_error "Please specify a version manually or check your internet connection"
        exit 1
    fi
    
    echo "$version"
}

#######################################
# Download and verify binary
#######################################
download_and_verify() {
    local version=$1
    local os=$2
    local arch=$3
    
    local archive_name="${BINARY_NAME}_${version#v}_${os}_${arch}.tar.gz"
    local download_url="${GITHUB_RELEASES}/${version}/${archive_name}"
    local checksums_url="${GITHUB_RELEASES}/${version}/checksums.txt"
    
    print_info "Downloading ${BINARY_NAME} ${version} for ${os}_${arch}..."
    
    # Download archive
    if ! curl -fsSL -o "${TMP_DIR}/${archive_name}" "$download_url"; then
        print_error "Failed to download ${archive_name}"
        print_error "URL: $download_url"
        print_error "Please check the version exists: https://github.com/${REPO}/releases"
        exit 1
    fi
    
    print_success "Downloaded ${archive_name}"
    
    # Download checksums
    print_info "Downloading checksums..."
    if ! curl -fsSL -o "${TMP_DIR}/checksums.txt" "$checksums_url"; then
        print_error "Failed to download checksums"
        exit 1
    fi
    
    # Verify checksum
    print_info "Verifying checksum..."
    
    cd "$TMP_DIR"
    
    local checksum_line
    checksum_line=$(grep "$archive_name" checksums.txt || true)
    
    if [ -z "$checksum_line" ]; then
        print_error "Checksum not found for ${archive_name}"
        exit 1
    fi
    
    # Use sha256sum on Linux, shasum on macOS
    if command -v sha256sum >/dev/null 2>&1; then
        echo "$checksum_line" | sha256sum -c --quiet || {
            print_error "Checksum verification failed"
            exit 1
        }
    else
        echo "$checksum_line" | shasum -a 256 -c --quiet || {
            print_error "Checksum verification failed"
            exit 1
        }
    fi
    
    print_success "Checksum verified"
    
    # Extract archive
    print_info "Extracting archive..."
    tar -xzf "$archive_name" || {
        print_error "Failed to extract archive"
        exit 1
    }
    
    if [ ! -f "${BINARY_NAME}" ]; then
        print_error "Binary not found in archive"
        exit 1
    fi
    
    print_success "Archive extracted"
}

#######################################
# Install binary
#######################################
install_binary() {
    local force=$1
    local binary_path="${INSTALL_DIR}/${BINARY_NAME}"
    
    # Check if already installed
    if [ -f "$binary_path" ]; then
        local current_version
        current_version=$("$binary_path" --version 2>/dev/null | awk '{print $NF}' || echo "unknown")
        
        if [ "$force" != "true" ]; then
            print_warning "${BINARY_NAME} is already installed at ${binary_path} (${current_version})"
            echo -n "Do you want to overwrite it? [y/N]: "
            read -r response
            
            case "$response" in
                [yY][eE][sS]|[yY])
                    print_info "Overwriting existing installation..."
                    ;;
                *)
                    print_info "Installation cancelled"
                    exit 0
                    ;;
            esac
        fi
    fi
    
    # Create install directory
    mkdir -p "$INSTALL_DIR"
    
    # Copy binary
    print_info "Installing to ${binary_path}..."
    cp "${TMP_DIR}/${BINARY_NAME}" "$binary_path"
    chmod +x "$binary_path"
    
    # Verify installation
    if ! "$binary_path" --version >/dev/null 2>&1; then
        print_error "Installation failed - binary is not executable"
        exit 1
    fi
    
    local installed_version
    installed_version=$("$binary_path" --version 2>/dev/null | awk '{print $NF}')
    
    print_success "${BINARY_NAME} ${installed_version} installed successfully!"
}

#######################################
# Print PATH setup instructions
#######################################
print_path_instructions() {
    local binary_path="${INSTALL_DIR}/${BINARY_NAME}"
    
    echo ""
    echo "Installation location: ${binary_path}"
    echo ""
    
    # Check if already in PATH
    if echo "$PATH" | grep -q "${INSTALL_DIR}"; then
        print_success "${INSTALL_DIR} is already in your PATH"
        echo ""
        echo "Verify installation: ${BINARY_NAME} --version"
    else
        print_warning "${INSTALL_DIR} is not in your PATH"
        echo ""
        echo "To use ${BINARY_NAME}, add this directory to your PATH:"
        echo ""
        echo "For bash (~/.bashrc) or zsh (~/.zshrc):"
        echo "  echo 'export PATH=\"\$HOME/.agent-smith/bin:\$PATH\"' >> ~/.zshrc"
        echo "  source ~/.zshrc"
        echo ""
        echo "For fish (~/.config/fish/config.fish):"
        echo "  fish_add_path ~/.agent-smith/bin"
        echo ""
        echo "Then verify: ${BINARY_NAME} --version"
    fi
    
    echo ""
    echo "Documentation: https://github.com/${REPO}#readme"
}

#######################################
# Parse command line arguments
#######################################
parse_args() {
    FORCE=false
    VERSION=""
    
    while [ $# -gt 0 ]; do
        case "$1" in
            --force)
                FORCE=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            -*)
                print_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
            *)
                VERSION="$1"
                shift
                ;;
        esac
    done
}

#######################################
# Main installation flow
#######################################
main() {
    # Parse arguments
    parse_args "$@"
    
    # Check requirements
    check_requirements
    
    # Detect platform
    OS=$(detect_os)
    ARCH=$(detect_arch)
    
    print_info "Detected platform: ${OS}_${ARCH}"
    
    # Get version
    if [ -z "$VERSION" ]; then
        VERSION=$(get_latest_version)
    fi
    
    print_info "Installing version: ${VERSION}"
    
    # Create temp directory
    TMP_DIR=$(mktemp -d)
    
    # Download and verify
    download_and_verify "$VERSION" "$OS" "$ARCH"
    
    # Install
    install_binary "$FORCE"
    
    # Print instructions
    print_path_instructions
}

# Run main function
main "$@"
