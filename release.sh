#!/bin/bash
# GitHub Release Script for Server Dashboard
# Uploads distribution packages to GitHub Releases

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Read version from VERSION file
if [ -f "VERSION" ]; then
    VERSION=$(cat VERSION | tr -d '[:space:]')
else
    echo -e "${RED}ERROR: VERSION file not found${NC}"
    exit 1
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Server Dashboard Release Script${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}Version: ${VERSION}${NC}"
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}ERROR: GitHub CLI (gh) is not installed${NC}"
    echo "Install it with: brew install gh"
    echo "Or visit: https://cli.github.com/"
    exit 1
fi

# Check if user is authenticated
if ! gh auth status &> /dev/null; then
    echo -e "${YELLOW}You are not authenticated with GitHub CLI${NC}"
    echo "Running: gh auth login"
    gh auth login
fi

# Check if dist directory exists
DIST_DIR="dist"
if [ ! -d "$DIST_DIR" ]; then
    echo -e "${RED}ERROR: ${DIST_DIR}/ directory not found${NC}"
    echo "Run ./build.sh first to create distribution packages"
    exit 1
fi

# Count distribution files
DIST_COUNT=$(ls -1 "$DIST_DIR" 2>/dev/null | wc -l | tr -d ' ')
if [ "$DIST_COUNT" -eq 0 ]; then
    echo -e "${RED}ERROR: No distribution packages found in ${DIST_DIR}/${NC}"
    echo "Run ./build.sh first to create distribution packages"
    exit 1
fi

echo -e "${YELLOW}Distribution packages found:${NC}"
ls -lh "$DIST_DIR"
echo ""

# Confirm release
echo -e "${YELLOW}This will create a new GitHub release: ${VERSION}${NC}"
echo -e "${YELLOW}And upload ${DIST_COUNT} distribution packages${NC}"
read -p "Continue? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Release cancelled"
    exit 0
fi

# Create release notes
RELEASE_NOTES=$(cat << EOF
# Server Dashboard ${VERSION}

## Distribution Packages

Download the appropriate package for your platform:

- **Linux amd64**: Standard 64-bit Linux servers/desktops
- **Linux arm64**: 64-bit ARM (Raspberry Pi 3/4/5 with 64-bit OS)
- **Linux armv7**: 32-bit ARM (Raspberry Pi 2/3/4 with 32-bit OS)
- **Linux armv6**: 32-bit ARM (Raspberry Pi Zero/1)
- **macOS Intel**: Intel-based Macs
- **macOS Apple Silicon**: M1/M2/M3 Macs
- **Windows**: 64-bit Windows systems

## Quick Start

1. Download and extract the appropriate package
2. Edit \`config/config.yaml\` with your server details
3. (Optional) Copy \`.env.example\` to \`.env\` and customize
4. Run \`./start.sh\` (Linux/macOS) or \`start.bat\` (Windows)

## What's New

See [CHANGELOG.md](https://github.com/jlamb85/dashboard/blob/main/CHANGELOG.md) for full release notes.

## Installation

### Linux/macOS
\`\`\`bash
tar -xzf server-dashboard-${VERSION}-linux-amd64.tar.gz
cd server-dashboard-${VERSION}-linux-amd64
./start.sh
\`\`\`

### Windows
1. Extract the ZIP file
2. Double-click \`start.bat\`

## Documentation

- [README](https://github.com/jlamb85/dashboard/blob/main/README.md)
- [Getting Started Guide](https://github.com/jlamb85/dashboard/blob/main/GETTING-STARTED.md)
- [Configuration Guide](https://github.com/jlamb85/dashboard/blob/main/PRODUCTION.md)

---

Built with Go and love ❤️
EOF
)

echo ""
echo -e "${YELLOW}Creating GitHub release...${NC}"

# Create the release
if gh release create "$VERSION" \
    --title "Server Dashboard $VERSION" \
    --notes "$RELEASE_NOTES" \
    "$DIST_DIR"/* ; then
    
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Release created successfully!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${GREEN}Release: https://github.com/jlamb85/dashboard/releases/tag/${VERSION}${NC}"
    echo ""
    echo "Distribution packages uploaded:"
    ls -1 "$DIST_DIR"
    echo ""
    echo -e "${BLUE}Total size:${NC}"
    du -sh "$DIST_DIR"
    echo ""
    
else
    echo ""
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}Release creation failed!${NC}"
    echo -e "${RED}========================================${NC}"
    echo ""
    echo "Possible issues:"
    echo "  - Release $VERSION already exists (use a different version)"
    echo "  - Not authenticated with GitHub"
    echo "  - Network connectivity issues"
    echo ""
    echo "To delete an existing release:"
    echo "  gh release delete $VERSION"
    echo ""
    exit 1
fi
