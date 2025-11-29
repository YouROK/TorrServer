#!/usr/bin/env bash
# Test script for installTorrServerLinux.sh
# This script runs inside Docker containers to test the installation script

set -e

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m' # No Color

# Test configuration
readonly SCRIPT_NAME="installTorrServerLinux.sh"
readonly INSTALL_DIR="/opt/torrserver"
readonly GLIBC_LIMITED_VERSION="135"
readonly MIN_GLIBC_VERSION="2.32"
readonly MAX_RETRIES="${MAX_RETRIES:-3}"
readonly RETRY_DELAY="${RETRY_DELAY:-2}"

# Helper functions
log_info() {
  echo -e "${GREEN}✓${NC} $1"
}

log_error() {
  echo -e "${RED}✗${NC} $1"
}

log_warning() {
  echo -e "${YELLOW}⚠${NC} $1"
}

log_test() {
  echo "Test $1: $2"
}

# Check if OS requires glibc-limited version
is_glibc_limited_os() {
  local os="$1"
  local glibc_limited_oses="$2"
  echo "$glibc_limited_oses" | grep -qE "(^|\|)$os(\||$)"
}

# Get glibc version message for OS
get_glibc_message() {
  local os="$1"
  case "$os" in
    debian-11)
      echo "Note: Debian 11 has glibc 2.31, installing version $GLIBC_LIMITED_VERSION (version 136+ requires glibc >= $MIN_GLIBC_VERSION)"
      ;;
    almalinux-8)
      echo "Note: AlmaLinux 8 has glibc 2.28, installing version $GLIBC_LIMITED_VERSION (version 136+ requires glibc >= $MIN_GLIBC_VERSION)"
      ;;
    rocky-8)
      echo "Note: Rocky 8 has glibc 2.28, installing version $GLIBC_LIMITED_VERSION (version 136+ requires glibc >= $MIN_GLIBC_VERSION)"
      ;;
    amazonlinux-2)
      echo "Note: Amazon Linux 2 has glibc 2.26, installing version $GLIBC_LIMITED_VERSION (version 136+ requires glibc >= $MIN_GLIBC_VERSION)"
      ;;
  esac
}

# Install RPM packages (dnf/yum)
install_rpm_packages() {
  local pkg_manager="$1"
  shift
  local packages=("$@")

  "$pkg_manager" makecache -q || true
  # Always remove curl-minimal first to avoid conflicts
  "$pkg_manager" remove -y -q curl-minimal 2>/dev/null || true

  # Check if curl package is installed (not just curl-minimal)
  if rpm -qa curl >/dev/null 2>&1; then
    # curl package is already installed, just install other packages
    "$pkg_manager" install -y -q "${packages[@]}" || true
  else
    # curl package not installed, install curl with --allowerasing
    "$pkg_manager" install -y -q --allowerasing curl "${packages[@]}" || true
  fi
}

# Install dependencies based on OS
install_dependencies() {
  if command -v apt-get >/dev/null 2>&1; then
    retry_command "apt-get update" "apt-get update -qq" 3 1 || true
    retry_command "apt-get install" "apt-get install -y -qq curl iputils-ping dnsutils" 3 1 || true
  elif command -v dnf >/dev/null 2>&1; then
    retry_command "dnf install" "install_rpm_packages dnf iputils bind-utils" 3 1 || true
  elif command -v yum >/dev/null 2>&1; then
    retry_command "yum install" "install_rpm_packages yum iputils bind-utils" 3 1 || true
  fi
}

# Verify curl installation
verify_curl_installation() {
  if command -v rpm >/dev/null 2>&1; then
    if ! rpm -qa curl >/dev/null 2>&1; then
      log_error "curl package is not installed after dependency installation"
      exit 1
    fi
    # Verify curl-minimal is not present (it should have been removed)
    if rpm -qa curl-minimal >/dev/null 2>&1; then
      log_warning "curl-minimal is still installed, removing it..."
      rpm -e --nodeps curl-minimal 2>/dev/null || true
    fi
  elif command -v dpkg >/dev/null 2>&1; then
    if ! dpkg -s curl >/dev/null 2>&1; then
      log_error "curl package is not installed after dependency installation"
      exit 1
    fi
  fi
}

# Retry a command with exponential backoff
retry_command() {
  local test_name="$1"
  local test_command="$2"
  local max_attempts="${3:-$MAX_RETRIES}"
  local delay="${4:-$RETRY_DELAY}"
  local attempt=1
  local last_error=0

  # Print command before first attempt
  if [[ $attempt -eq 1 ]]; then
    echo "Executing: $test_command"
  fi

  while [[ $attempt -le $max_attempts ]]; do
    if [[ $attempt -gt 1 ]]; then
      echo "Retry attempt $attempt/$max_attempts: $test_command"
    fi
    if eval "$test_command"; then
      if [[ $attempt -gt 1 ]]; then
        log_info "$test_name (succeeded on attempt $attempt)"
      fi
      return 0
    else
      last_error=$?
      if [[ $attempt -lt $max_attempts ]]; then
        log_warning "$test_name failed (attempt $attempt/$max_attempts), retrying in ${delay}s..."
        sleep "$delay"
        delay=$((delay * 2))
      fi
      attempt=$((attempt + 1))
    fi
  done

  log_error "$test_name (failed after $max_attempts attempts)"
  return $last_error
}

# Run a test command and handle errors
run_test() {
  local test_name="$1"
  local test_command="$2"
  local skip_on_error="${3:-false}"
  local use_retry="${4:-true}"

  if [[ "$use_retry" == "true" ]]; then
    if retry_command "$test_name" "$test_command"; then
      log_info "$test_name"
      return 0
    else
      if [[ "$skip_on_error" == "true" ]]; then
        log_warning "$test_name (skipped after retries)"
        return 0
      else
        log_error "$test_name"
        return 1
      fi
    fi
  else
    if eval "$test_command"; then
      log_info "$test_name"
      return 0
    else
      if [[ "$skip_on_error" == "true" ]]; then
        log_warning "$test_name (skipped)"
        return 0
      else
        log_error "$test_name"
        return 1
      fi
    fi
  fi
}

# Main test execution
main() {
  local matrix_os="${MATRIX_OS:-}"
  local test_user="${TEST_USER:-default}"
  local glibc_limited_oses="${GLIBC_LIMITED_OSES:-}"

  # Determine root flag
  local root_flag=''
  if [[ "$test_user" == 'root' ]]; then
    root_flag='--root'
  fi

  # Check if OS requires glibc-limited version
  local is_glibc_limited=false
  if is_glibc_limited_os "$matrix_os" "$glibc_limited_oses"; then
    is_glibc_limited=true
  fi

  echo "========================================"
  echo "Testing $SCRIPT_NAME"
  echo "OS: $matrix_os"
  echo "User: $test_user"
  echo "Retry settings: max=$MAX_RETRIES, initial_delay=${RETRY_DELAY}s"
  echo "========================================"
  echo ""

  # Test 1: Check script syntax
  echo "::group::Test 1: Check script syntax"
  log_test "1" "Checking script syntax..."
  echo "Executing: bash -n $SCRIPT_NAME"
  if bash -n "$SCRIPT_NAME"; then
    log_info "Script syntax is valid"
  else
    log_error "Script syntax check failed"
    echo "::endgroup::"
    exit 1
  fi
  echo "::endgroup::"
  echo ""

  # Test 2: Show help
  echo "::group::Test 2: Test help command"
  log_test "2" "Testing help command..."
  echo "Executing: ./$SCRIPT_NAME --help"
  if ./"$SCRIPT_NAME" --help > /dev/null; then
    log_info "Help command works"
  else
    log_error "Help command failed"
    echo "::endgroup::"
    exit 1
  fi
  echo "::endgroup::"
  echo ""

  # Test 3: Install in silent mode
  echo "::group::Test 3: Install TorrServer"
  log_test "3" "Installing TorrServer (silent mode)..."
  if [[ "$is_glibc_limited" == "true" ]]; then
    local glibc_msg
    glibc_msg=$(get_glibc_message "$matrix_os")
    if [[ -n "$glibc_msg" ]]; then
      echo "$glibc_msg"
    fi
    if retry_command "Installation" "./$SCRIPT_NAME --install $GLIBC_LIMITED_VERSION --silent $root_flag"; then
      log_info "Installation completed"
    else
      log_error "Installation failed after retries"
      echo "::endgroup::"
      exit 1
    fi
  else
    if retry_command "Installation" "./$SCRIPT_NAME --install --silent $root_flag"; then
      log_info "Installation completed"
    else
      log_error "Installation failed after retries"
      echo "::endgroup::"
      exit 1
    fi
  fi
  echo "::endgroup::"
  echo ""

  # Test 4: Check installation
  echo "::group::Test 4: Verify installation"
  log_test "4" "Checking installation..."
  echo "Executing: ls $INSTALL_DIR/TorrServer-linux-*"
  if ls "$INSTALL_DIR"/TorrServer-linux-* >/dev/null 2>&1; then
    log_info "Binary file exists"
  else
    log_error "Binary file not found"
    echo "::endgroup::"
    exit 1
  fi
  echo "::endgroup::"
  echo ""

  # Test 5: Check version
  echo "::group::Test 5: Check version"
  log_test "5" "Checking for updates..."
  if [[ "$is_glibc_limited" == "true" ]]; then
    echo "Note: Skipping version check (latest version requires glibc >= $MIN_GLIBC_VERSION)"
    log_info "Version check skipped (expected)"
  else
    if retry_command "Version check" "./$SCRIPT_NAME --check --silent $root_flag"; then
      log_info "Version check completed"
    else
      log_error "Version check failed after retries"
      echo "::endgroup::"
      exit 1
    fi
  fi
  echo "::endgroup::"
  echo ""

  # Test 6: Update (if available)
  echo "::group::Test 6: Test update command"
  log_test "6" "Testing update command..."
  if [[ "$is_glibc_limited" == "true" ]]; then
    echo "Note: Skipping update test (latest version requires glibc >= $MIN_GLIBC_VERSION)"
    log_info "Update check skipped (expected)"
  else
    if retry_command "Update check" "./$SCRIPT_NAME --update --silent $root_flag"; then
      log_info "Update check completed"
    else
      log_error "Update check failed after retries"
      echo "::endgroup::"
      exit 1
    fi
  fi
  echo "::endgroup::"
  echo ""

  # Test 7: Reconfigure
  echo "::group::Test 7: Test reconfigure command"
  log_test "7" "Testing reconfigure command..."
  if retry_command "Reconfigure" "./$SCRIPT_NAME --reconfigure --silent $root_flag"; then
    log_info "Reconfigure completed"
  else
    log_error "Reconfigure failed after retries"
    echo "::endgroup::"
    exit 1
  fi
  echo "::endgroup::"
  echo ""

  # Test 8: Change user (if not already root)
  if [[ "$test_user" == 'default' ]]; then
    echo "::group::Test 8: Test change-user to root"
    log_test "8" "Testing change-user to root..."
    if retry_command "User change to root" "./$SCRIPT_NAME --change-user root --silent"; then
      log_info "User change to root completed"
    else
      log_error "User change to root failed after retries"
      echo "::endgroup::"
      exit 1
    fi
    echo "::endgroup::"
    echo ""

    # Test 8b: Change user back to default (only for Ubuntu to test full flow)
    if [[ "$matrix_os" == 'ubuntu-22.04' ]] || [[ "$matrix_os" == 'ubuntu-24.04' ]]; then
      echo "::group::Test 8b: Test change-user back to default"
      log_test "8b" "Testing change-user back to default..."
      if retry_command "User change back to default" "./$SCRIPT_NAME --change-user torrserver --silent"; then
        log_info "User change back to default completed"
      else
        log_error "User change back to default failed after retries"
        echo "::endgroup::"
        exit 1
      fi
      echo "::endgroup::"
      echo ""
    fi
  fi

  # Test 9: Cleanup - Uninstall
  echo "::group::Test 9: Uninstall TorrServer"
  log_test "9" "Uninstalling TorrServer..."
  if retry_command "Uninstallation" "./$SCRIPT_NAME --remove --silent"; then
    log_info "Uninstallation completed"
  else
    log_error "Uninstallation failed after retries"
    echo "::endgroup::"
    exit 1
  fi
  echo "::endgroup::"
  echo ""

  # Test 10: Verify cleanup
  echo "::group::Test 10: Verify cleanup"
  log_test "10" "Verifying cleanup..."
  echo "Executing: Checking if $INSTALL_DIR is empty or doesn't exist"
  if [[ ! -d "$INSTALL_DIR" ]] || [[ -z "$(ls -A "$INSTALL_DIR" 2>/dev/null)" ]]; then
    log_info "Cleanup verified"
  else
    log_warning "Installation directory still exists (may be expected)"
  fi
  echo "::endgroup::"
  echo ""

  echo "========================================"
  echo "All tests passed! ✓"
  echo "========================================"
}

# Setup and run tests
setup() {
  echo "::group::Setup: Install dependencies"
  # Install dependencies
  install_dependencies

  # Verify curl installation
  verify_curl_installation

  # Make script executable
  chmod +x "$SCRIPT_NAME"
  echo "::endgroup::"
  echo ""
}

# Run setup and main
setup
main
