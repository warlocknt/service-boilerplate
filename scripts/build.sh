#!/bin/bash

# Build script for service-boilerplate
# Runs tests and builds the binary if all tests pass

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="service-boilerplate"
BUILD_DIR="./build"
TEST_TIMEOUT="120s"

# Disable CGO (required for cross-platform builds)
export CGO_ENABLED=0

# Functions
print_header() {
    echo -e "${CYAN}==> $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Clean build directory
clean() {
    print_header "Cleaning build directory..."
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"
    print_success "Clean complete"
}

# Download dependencies
deps() {
    print_header "Downloading dependencies..."
    go mod download
    go mod tidy
    print_success "Dependencies ready"
}

# Run tests
run_tests() {
    print_header "Running tests..."
    
    if ! go test -v -race -timeout "$TEST_TIMEOUT" ./...; then
        print_error "Tests failed!"
        exit 1
    fi
    
    print_success "All tests passed"
}

# Run fast tests (without race detector)
run_tests_fast() {
    print_header "Running tests (fast mode)..."
    
    if ! go test -v -timeout 60s ./...; then
        print_error "Tests failed!"
        exit 1
    fi
    
    print_success "All tests passed"
}

# Generate coverage report
coverage() {
    print_header "Generating coverage report..."
    
    go test -coverprofile=coverage.out -timeout "$TEST_TIMEOUT" ./...
    go tool cover -func=coverage.out
    
    print_success "Coverage report generated: coverage.out"
    
    # Generate HTML report
    go tool cover -html=coverage.out -o coverage.html
    print_success "HTML report generated: coverage.html"
}

# Build binary
build() {
    local output_name="${1:-$BINARY_NAME}"
    local goos="${2:-}"
    local goarch="${3:-}"
    
    print_header "Building binary: $output_name..."
    
    local output="$BUILD_DIR/$output_name"
    
    if [ -n "$goos" ]; then
        GOOS="$goos" GOARCH="$goarch" go build -ldflags="-s -w" -o "$output" ./cmd/service-boilerplate
    else
        go build -ldflags="-s -w" -o "$output" ./cmd/service-boilerplate
    fi
    
    print_success "Build complete: $output"
}

# Build for Windows (cross-compile from Linux/Mac)
build_win() {
    print_header "Building for Windows..."
    build "$BINARY_NAME-windows-amd64.exe" "windows" "amd64"
    print_success "Windows build complete"
}

# Build for Linux (native or cross-compile)
build_linux() {
    print_header "Building for Linux..."
    # Linux AMD64
    build "$BINARY_NAME-linux-amd64" "linux" "amd64"
    # Linux ARM64
    build "$BINARY_NAME-linux-arm64" "linux" "arm64"
    print_success "Linux build complete"
}

# Build for multiple platforms
build_all() {
    print_header "Building for multiple platforms..."
    
    # Linux AMD64
    build "$BINARY_NAME-linux-amd64" "linux" "amd64"
    
    # Linux ARM64
    build "$BINARY_NAME-linux-arm64" "linux" "arm64"
    
    # Windows AMD64
    build "$BINARY_NAME-windows-amd64.exe" "windows" "amd64"
    
    # Darwin AMD64 (macOS)
    build "$BINARY_NAME-darwin-amd64" "darwin" "amd64"
    
    # Darwin ARM64 (macOS M1)
    build "$BINARY_NAME-darwin-arm64" "darwin" "arm64"
    
    print_success "Multi-platform build complete"
}

# Run code checks
checks() {
    print_header "Running code checks..."
    
    # Format
    echo "Formatting code..."
    go fmt ./...
    
    # Vet
    echo "Running go vet..."
    go vet ./...
    
    print_success "Checks passed"
}

# Run CI pipeline
ci() {
    echo -e "${YELLOW}Running in CI mode...${NC}"
    deps
    checks
    run_tests
    clean
    build_all
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  CI Pipeline Completed Successfully!${NC}"
    echo -e "${GREEN}========================================${NC}"
}

# Show help
help() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  all         Run tests and build (default)"
    echo "  test        Run all tests with race detector"
    echo "  test-fast   Run tests without race detector"
    echo "  build       Build binary (requires passing tests)"
    echo "  build-only  Build without running tests"
    echo "  build-all   Build for multiple platforms (Linux, Windows, macOS)"
    echo "  build-win   Cross-compile for Windows from Linux/Mac"
    echo "  build-linux Build for Linux (amd64, arm64)"
    echo "  coverage    Generate coverage report"
    echo "  check       Run formatting and vet"
    echo "  clean       Clean build artifacts"
    echo "  deps        Download dependencies"
    echo "  ci          Full CI pipeline"
    echo "  help        Show this help message"
}

# Main
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  Service Boilerplate Build Script${NC}"
echo -e "${CYAN}========================================${NC}"

# Parse command
case "${1:-all}" in
    all)
        run_tests
        clean
        build
        echo -e "${GREEN}========================================${NC}"
        echo -e "${GREEN}  Build Completed Successfully!${NC}"
        echo -e "${GREEN}========================================${NC}"
        ;;
    test)
        run_tests
        ;;
    test-fast)
        run_tests_fast
        ;;
    build)
        run_tests
        clean
        build
        ;;
    build-only)
        clean
        build
        ;;
    build-all)
        run_tests
        clean
        build_all
        ;;
    build-win)
        clean
        build_win
        ;;
    build-linux)
        clean
        build_linux
        ;;
    coverage)
        coverage
        ;;
    check)
        checks
        ;;
    clean)
        clean
        ;;
    deps)
        deps
        ;;
    ci)
        ci
        ;;
    help)
        help
        ;;
    *)
        print_error "Unknown command: $1"
        help
        exit 1
        ;;
esac
