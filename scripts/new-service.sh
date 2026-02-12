#!/bin/bash

# New Service Generator
# Based on service-boilerplate

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# Functions
print_header() {
    echo -e "${CYAN}==> $1${NC}"
}

print_success() {
    echo -e "${GREEN}[OK] $1${NC}"
}

print_error() {
    echo -e "${RED}[ERROR] $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

show_help() {
    echo ""
    echo "New Service Generator"
    echo "===================="
    echo ""
    echo "Usage: ./new-service.sh <project-name> [output-directory]"
    echo ""
    echo "Examples:"
    echo "  ./new-service.sh my-api-service"
    echo "  ./new-service.sh payment-service ~/projects"
    echo "  ./new-service.sh notification-service .."
    echo ""
    echo "This will:"
    echo "  1. Clone service-boilerplate from GitHub"
    echo "  2. Rename Go module"
    echo "  3. Update configuration files"
    echo "  4. Initialize new git repository"
    echo "  5. Clean up build artifacts"
    echo ""
}

# Check arguments
if [ $# -eq 0 ] || [ "$1" == "help" ] || [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
    show_help
    exit 0
fi

PROJECT_NAME=$1
OUTPUT_DIR=${2:-.}

echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  New Service Generator${NC}"
echo -e "${CYAN}  Based on service-boilerplate${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

print_header "Creating new project: $PROJECT_NAME"

# Check if directory exists
if [ -d "$OUTPUT_DIR/$PROJECT_NAME" ]; then
    print_error "Directory $OUTPUT_DIR/$PROJECT_NAME already exists"
    exit 1
fi

# Create directory
mkdir -p "$OUTPUT_DIR/$PROJECT_NAME"
print_success "Directory created"

# Clone boilerplate
print_header "Cloning boilerplate..."
git clone --depth 1 https://github.com/warlocknt/service-boilerplate.git "$OUTPUT_DIR/$PROJECT_NAME/temp"
if [ $? -ne 0 ]; then
    print_error "Failed to clone boilerplate"
    rm -rf "$OUTPUT_DIR/$PROJECT_NAME"
    exit 1
fi

# Move files from temp
cd "$OUTPUT_DIR/$PROJECT_NAME/temp"
mv * ../ 2>/dev/null || true
mv .* ../ 2>/dev/null || true
cd ..
rm -rf temp

# Remove .git and initialize new
print_header "Initializing new git repository..."
rm -rf .git
git init
git add .
git commit -m "Initial commit based on service-boilerplate"
print_success "Git repository initialized"

# Update go.mod module name
print_header "Updating module name..."
if command -v go &> /dev/null; then
    go mod edit -module "$PROJECT_NAME"
else
    print_warning "Go not found, updating go.mod manually..."
    sed -i "s/module service-boilerplate/module $PROJECT_NAME/g" go.mod
fi
print_success "Module name updated to: $PROJECT_NAME"

# Update config
print_header "Updating configuration..."
sed -i "s/service-boilerplate/$PROJECT_NAME/g" configs/config.yaml
print_success "Configuration updated"

# Update README
print_header "Updating README..."
sed -i "s/service-boilerplate/$PROJECT_NAME/g" README.md
print_success "README updated"

# Clean up
print_header "Cleaning up..."
rm -f service.exe
rm -rf build
rm -rf logs
rm -f coverage.out
rm -f coverage.html
print_success "Cleanup complete"

# Create new commit with changes
git add .
git commit -m "chore: rename module to $PROJECT_NAME"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Project created successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Location: $OUTPUT_DIR/$PROJECT_NAME"
echo ""
echo "Next steps:"
echo "  cd $PROJECT_NAME"
echo "  make"
echo ""
echo "Or:"
echo "  go mod tidy"
echo "  go build -o $PROJECT_NAME ./cmd/service-boilerplate"
echo ""
