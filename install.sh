#!/bin/bash

# ccswitch Installer
# A friendly CLI tool for managing multiple git worktrees

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_OWNER="ksred"
REPO_NAME="ccswitch"
BINARY_NAME="ccswitch"
INSTALL_DIR="/usr/local/bin"
USER_INSTALL_DIR="$HOME/.local/bin"

# Default values
VERSION="latest"
INSTALL_LOCATION="system"
FORCE=false
SKIP_SHELL_INTEGRATION=false

# Print colored output
print_info() {
    echo -e "${BLUE}ℹ${NC} $1" >&2
}

print_success() {
    echo -e "${GREEN}✓${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1" >&2
}

print_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

# Show help
show_help() {
    cat << EOF
ccswitch Installer

USAGE:
    install.sh [OPTIONS]

OPTIONS:
    -v, --version VERSION     Install specific version (default: latest)
    -l, --location LOCATION   Install location: system, user (default: system)
    -f, --force              Force reinstall even if already installed
    -s, --skip-shell         Skip shell integration setup
    -h, --help               Show this help message

EXAMPLES:
    install.sh                           # Install latest version to system
    install.sh --version v1.0.0         # Install specific version
    install.sh --location user          # Install to user directory
    install.sh --force                  # Force reinstall

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -l|--location)
                INSTALL_LOCATION="$2"
                shift 2
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -s|--skip-shell)
                SKIP_SHELL_INTEGRATION=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Detect operating system and architecture
detect_platform() {
    local os arch
    
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          print_error "Unsupported operating system: $(uname -s)"; exit 1 ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        armv7l)         arch="armv7" ;;
        *)              print_error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac
    
    echo "${os}-${arch}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check for curl or wget
    if ! command_exists curl && ! command_exists wget; then
        print_error "curl or wget is required but not installed"
        exit 1
    fi
    
    # Check for git
    if ! command_exists git; then
        print_error "git is required but not installed"
        exit 1
    fi
    
    # Check git version (need 2.20+ for worktree support)
    local git_version
    git_version=$(git --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    local git_major git_minor
    git_major=$(echo "$git_version" | cut -d. -f1)
    git_minor=$(echo "$git_version" | cut -d. -f2)
    
    if [ "$git_major" -lt 2 ] || ([ "$git_major" -eq 2 ] && [ "$git_minor" -lt 20 ]); then
        print_error "git version 2.20+ is required (found: $git_version)"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Get latest release version
get_latest_version() {
    local api_url="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
    local response
    
    if command_exists curl; then
        response=$(curl -s "$api_url")
    else
        response=$(wget -qO- "$api_url")
    fi
    
    # Check if we got a valid response
    if echo "$response" | grep -q '"tag_name"'; then
        echo "$response" | grep '"tag_name"' | sed 's/.*"tag_name": "\(.*\)".*/\1/'
    else
        echo "no-releases"
    fi
}

# Build from source
build_from_source() {
    local platform="$1"
    local temp_dir
    temp_dir=$(mktemp -d)
    
    print_info "Building ${BINARY_NAME} from source..."
    
    # Clone repository
    if ! git clone "https://github.com/${REPO_OWNER}/${REPO_NAME}.git" "$temp_dir/ccswitch"; then
        print_error "Failed to clone repository"
        exit 1
    fi
    
    cd "$temp_dir/ccswitch"
    
    # Check if Go is available
    if ! command_exists go; then
        print_error "Go is required to build from source but not installed"
        print_info "Please install Go from https://golang.org/dl/"
        exit 1
    fi
    
    # Build binary
    if ! go build -o "$BINARY_NAME" .; then
        print_error "Failed to build binary"
        exit 1
    fi
    
    # Make binary executable
    chmod +x "$BINARY_NAME"
    
    echo "$temp_dir/ccswitch/$BINARY_NAME"
}

# Download binary
download_binary() {
    local version="$1"
    local platform="$2"
    local download_url
    local archive_path
    local binary_path
    
    if [ "$version" = "latest" ]; then
        version=$(get_latest_version)
        if [ "$version" = "no-releases" ]; then
            print_warning "No releases found, building from source..."
            echo $(build_from_source "$platform")
            return
        fi
        print_info "Latest version: $version"
    fi
    
    # Construct download URL and archive path
    if [ "$platform" = "windows-amd64" ] || [ "$platform" = "windows-arm64" ]; then
        download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${version}/${BINARY_NAME}-${platform}.zip"
        archive_path="${BINARY_NAME}-${platform}.zip"
        binary_path="${BINARY_NAME}.exe"
    else
        download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${version}/${BINARY_NAME}-${platform}.tar.gz"
        archive_path="${BINARY_NAME}-${platform}.tar.gz"
        binary_path="${BINARY_NAME}"
    fi
    
    print_info "Downloading ${BINARY_NAME} ${version} for ${platform}..."
    
    # Create temporary directory
    local temp_dir
    temp_dir=$(mktemp -d)
    cd "$temp_dir"
    
    # Download archive
    if command_exists curl; then
        if ! curl -L -o "$archive_path" "$download_url"; then
            print_error "Failed to download archive from $download_url"
            print_info "Falling back to building from source..."
            echo $(build_from_source "$platform")
            return
        fi
    else
        if ! wget -O "$archive_path" "$download_url"; then
            print_error "Failed to download archive from $download_url"
            print_info "Falling back to building from source..."
            echo $(build_from_source "$platform")
            return
        fi
    fi
    
    # Extract archive
    if [ "$platform" = "windows-amd64" ] || [ "$platform" = "windows-arm64" ]; then
        if ! command_exists unzip; then
            print_error "unzip is required to extract the archive but not installed"
            print_info "Falling back to building from source..."
            echo $(build_from_source "$platform")
            return
        fi
        unzip -q "$archive_path"
    else
        tar -xzf "$archive_path"
    fi

    # Remove archive after extraction
    rm "$archive_path"

    # Check if binary was extracted
    if [ ! -f "$binary_path" ]; then
        print_error "Binary not found after extraction: $binary_path"
        print_info "Falling back to building from source..."
        echo $(build_from_source "$platform")
        return
    fi

    # Make binary executable
    chmod +x "$binary_path"

    echo "$temp_dir/$binary_path"
}

# Install binary
install_binary() {
    local binary_path="$1"
    local install_path
    local sudo_cmd=""
    
    # Determine install path
    if [ "$INSTALL_LOCATION" = "user" ]; then
        install_path="$USER_INSTALL_DIR/$BINARY_NAME"
        mkdir -p "$USER_INSTALL_DIR"
    else
        install_path="$INSTALL_DIR/$BINARY_NAME"
        # Check if we need sudo
        if [ ! -w "$INSTALL_DIR" ]; then
            if ! command_exists sudo; then
                print_error "sudo is required for system installation but not available"
                exit 1
            fi
            sudo_cmd="sudo"
        fi
    fi
    
    # Check if already installed
    if [ -f "$install_path" ] && [ "$FORCE" = false ]; then
        print_warning "${BINARY_NAME} is already installed at $install_path"
        print_info "Use --force to reinstall"
        return 0
    fi
    
    # Install binary
    print_info "Installing ${BINARY_NAME} to $install_path..."
    if [ -n "$sudo_cmd" ]; then
        $sudo_cmd cp "$binary_path" "$install_path"
        $sudo_cmd chmod +x "$install_path"
    else
        cp "$binary_path" "$install_path"
        chmod +x "$install_path"
    fi
    
    print_success "Binary installed successfully"
    
    # Clean up
    rm -rf "$(dirname "$binary_path")"
}

# Setup shell integration
setup_shell_integration() {
    if [ "$SKIP_SHELL_INTEGRATION" = true ]; then
        print_info "Skipping shell integration setup"
        return 0
    fi
    
    print_info "Setting up shell integration..."
    
    local shell_config
    local shell_name
    
    # Detect shell
    if [ -n "${ZSH_VERSION:-}" ] || [ "$SHELL" = "/bin/zsh" ] || [ "$SHELL" = "/usr/bin/zsh" ]; then
        shell_config="$HOME/.zshrc"
        shell_name="zsh"
    elif [ -n "${BASH_VERSION:-}" ] || [ "$SHELL" = "/bin/bash" ] || [ "$SHELL" = "/usr/bin/bash" ]; then
        shell_config="$HOME/.bashrc"
        shell_name="bash"
    else
        print_warning "Could not detect shell type. Manual setup required."
        print_info "Add this to your shell configuration file:"
        echo "  eval \"\$$($BINARY_NAME shell-init)\""
        return 0
    fi
    
    # Check if already configured
    if [ -f "$shell_config" ] && grep -q "eval \"\$$($BINARY_NAME shell-init)\"" "$shell_config" 2>/dev/null; then
        print_success "Shell integration already configured in $shell_config"
        return 0
    fi
    
    # Add shell integration
    echo "" >> "$shell_config"
    echo "# ccswitch shell integration" >> "$shell_config"
    echo "eval \"\$$($BINARY_NAME shell-init)\"" >> "$shell_config"
    
    print_success "Shell integration added to $shell_config"
    print_info "To activate now, run: source $shell_config"
}

# Verify installation
verify_installation() {
    print_info "Verifying installation..."
    
    local binary_path
    if [ "$INSTALL_LOCATION" = "user" ]; then
        binary_path="$USER_INSTALL_DIR/$BINARY_NAME"
    else
        binary_path="$INSTALL_DIR/$BINARY_NAME"
    fi
    
    if [ ! -f "$binary_path" ]; then
        print_error "Binary not found at $binary_path"
        return 1
    fi
    
    if ! "$binary_path" version >/dev/null 2>&1; then
        print_error "Binary is not executable or corrupted"
        return 1
    fi

    print_success "Installation verified successfully"

    # Show version
    local version_output
    version_output=$("$binary_path" version 2>/dev/null || echo "version unknown")
    print_info "Installed version: $version_output"
}

# Main installation function
main() {
    print_info "ccswitch Installer"
    echo
    
    # Parse arguments
    parse_args "$@"
    
    # Check prerequisites
    check_prerequisites
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    print_info "Detected platform: $platform"
    
    # Download and install binary
    local binary_path
    binary_path=$(download_binary "$VERSION" "$platform")

    # Check if we got a valid path
    if [ ! -f "$binary_path" ]; then
        print_error "Failed to obtain binary"
        exit 1
    fi
    
    install_binary "$binary_path"
    
    # Setup shell integration
    setup_shell_integration
    
    # Verify installation
    verify_installation
    
    echo
    print_success "Installation complete!"
    print_info "Run '$BINARY_NAME --help' to get started"
    
    if [ "$INSTALL_LOCATION" = "user" ] && [[ ":$PATH:" != *":$USER_INSTALL_DIR:"* ]]; then
        print_warning "Add $USER_INSTALL_DIR to your PATH to use $BINARY_NAME"
        print_info "Add this to your shell configuration:"
        echo "  export PATH=\"\$PATH:$USER_INSTALL_DIR\""
    fi
}

# Run main function only if script is executed directly
if [ -z "${BASH_SOURCE[0]:-}" ] || [ "${BASH_SOURCE[0]:-}" = "${0}" ]; then
    main "$@"
fi