#!/bin/bash
# Multi-platform Build Script for Server Dashboard
# Builds binaries for Linux, macOS, and Windows with bundled releases

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Read version from VERSION file
if [ -f "VERSION" ]; then
    VERSION=$(cat VERSION | tr -d '[:space:]')
else
    VERSION="v1.0.0"
    echo "$VERSION" > VERSION
fi

echo -e "${GREEN}Building Server Dashboard ${VERSION}${NC}"
echo "=============================================="

# Build output directory
BUILD_DIR="build"
DIST_DIR="dist"
APP_NAME="server-dashboard"

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf "$BUILD_DIR" "$DIST_DIR"
mkdir -p "$BUILD_DIR" "$DIST_DIR"

# Build information
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

echo ""
echo "Build Information:"
echo "  Version:    $VERSION"
echo "  Build Time: $BUILD_TIME"
echo "  Git Commit: $GIT_COMMIT"
echo ""

# Function to build for a specific platform
build_platform() {
    local GOOS=$1
    local GOARCH=$2
    local OUTPUT_NAME=$3
    local GOARM=$4  # Optional ARM version (6, 7)
    local BUNDLE_NAME="${APP_NAME}-${VERSION}-${GOOS}-${GOARCH}"
    
    # Add GOARM suffix to bundle name if specified
    if [ -n "$GOARM" ]; then
        BUNDLE_NAME="${BUNDLE_NAME}v${GOARM}"
        echo -e "${YELLOW}Building for ${GOOS}/${GOARCH} ARMv${GOARM}...${NC}"
    else
        echo -e "${YELLOW}Building for ${GOOS}/${GOARCH}...${NC}"
    fi
    
    # Build binary with optional GOARM
    if [ -n "$GOARM" ]; then
        GOOS=$GOOS GOARCH=$GOARCH GOARM=$GOARM go build -ldflags "$LDFLAGS" -o "${BUILD_DIR}/${OUTPUT_NAME}" main.go
    else
        GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "$LDFLAGS" -o "${BUILD_DIR}/${OUTPUT_NAME}" main.go
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Built ${OUTPUT_NAME}${NC}"
        
        # Create bundle directory
        BUNDLE_DIR="${BUILD_DIR}/${BUNDLE_NAME}"
        mkdir -p "$BUNDLE_DIR"
        
        # Copy binary
        cp "${BUILD_DIR}/${OUTPUT_NAME}" "$BUNDLE_DIR/"
        
        # Copy configuration and web assets
        cp -r config "$BUNDLE_DIR/"
        cp -r web "$BUNDLE_DIR/"
        cp README.md "$BUNDLE_DIR/" 2>/dev/null || true
        cp DISK_MONITORING.md "$BUNDLE_DIR/" 2>/dev/null || true
        cp VERSION "$BUNDLE_DIR/"
        
        # Create sample .env file
        cat > "$BUNDLE_DIR/.env.example" << 'EOF'
# Server Dashboard Environment Configuration
# Copy this file to .env and customize

# Server Configuration
SERVER_ADDRESS=0.0.0.0:8080
ENVIRONMENT=production

# Authentication
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=change-me-in-production

# TLS/HTTPS (optional)
TLS_ENABLED=false
TLS_CERT_FILE=/path/to/cert.pem
TLS_KEY_FILE=/path/to/key.pem

# Monitoring
USE_MOCK_DATA=false
SSH_ENABLED=true
SSH_USERNAME=monitor
SSH_PRIVATE_KEY_PATH=~/.ssh/id_rsa
SSH_PASSWORD=
EOF

        # Create startup script based on OS
        if [ "$GOOS" = "windows" ]; then
            cat > "$BUNDLE_DIR/start.bat" << 'EOF'
@echo off
echo Starting Server Dashboard...
server-dashboard.exe
pause
EOF
            cat > "$BUNDLE_DIR/README.txt" << EOF
Server Dashboard ${VERSION}
========================

Quick Start:
1. Edit config/config.yaml with your server details
2. (Optional) Copy .env.example to .env and customize
3. Run start.bat

For more information, see README.md

Documentation: https://github.com/yourusername/server-dashboard
EOF
        else
            cat > "$BUNDLE_DIR/start.sh" << EOF
#!/bin/bash
echo "Starting Server Dashboard ${VERSION}..."
cd "\$(dirname "\$0")"
./${OUTPUT_NAME}
EOF
            chmod +x "$BUNDLE_DIR/start.sh"
            
            cat > "$BUNDLE_DIR/README.txt" << EOF
Server Dashboard ${VERSION}
========================

Quick Start:
1. Edit config/config.yaml with your server details
2. (Optional) Copy .env.example to .env and customize
3. Run: ./start.sh

For more information, see README.md

Documentation: https://github.com/yourusername/server-dashboard
EOF
        fi
        
        # Create archive
        echo -e "${YELLOW}Creating bundle archive...${NC}"
        cd "$BUILD_DIR"
        
        if [ "$GOOS" = "windows" ]; then
            zip -rq "${BUNDLE_NAME}.zip" "$BUNDLE_NAME"
            mv "${BUNDLE_NAME}.zip" "../${DIST_DIR}/"
            echo -e "${GREEN}✓ Created ${BUNDLE_NAME}.zip${NC}"
        else
            tar -czf "${BUNDLE_NAME}.tar.gz" "$BUNDLE_NAME"
            mv "${BUNDLE_NAME}.tar.gz" "../${DIST_DIR}/"
            echo -e "${GREEN}✓ Created ${BUNDLE_NAME}.tar.gz${NC}"
        fi
        
        cd ..
    else
        echo -e "${RED}✗ Failed to build ${OUTPUT_NAME}${NC}"
        return 1
    fi
    
    echo ""
}

# Build for all platforms
echo -e "${GREEN}Starting multi-platform build...${NC}"
echo ""

# Linux builds
build_platform "linux" "amd64" "${APP_NAME}-linux-amd64"
build_platform "linux" "arm64" "${APP_NAME}-linux-arm64"

# Raspberry Pi builds (32-bit ARMv7 - Pi 2/3/4 with 32-bit OS)
build_platform "linux" "arm" "${APP_NAME}-linux-armv7" "7"

# Raspberry Pi builds (32-bit ARMv6 - Pi Zero/1)
build_platform "linux" "arm" "${APP_NAME}-linux-armv6" "6"

# macOS builds
build_platform "darwin" "amd64" "${APP_NAME}-darwin-amd64"
build_platform "darwin" "arm64" "${APP_NAME}-darwin-arm64"

# Windows builds
build_platform "windows" "amd64" "${APP_NAME}-windows-amd64.exe"

echo -e "${GREEN}=============================================="
echo "Build Complete!"
echo "=============================================="
echo ""
echo "Distribution packages created in ${DIST_DIR}/:"
ls -lh "$DIST_DIR"
echo ""
echo "Total size:"
du -sh "$DIST_DIR"
echo ""
echo -e "${GREEN}All builds successful!${NC}"
echo ""
echo "Platform Guide:"
echo "  Linux amd64:     Standard 64-bit Linux (servers, desktops)"
echo "  Linux arm64:     64-bit ARM (Raspberry Pi 3/4/5 with 64-bit OS)"
echo "  Linux armv7:     32-bit ARM (Raspberry Pi 2/3/4 with 32-bit OS)"
echo "  Linux armv6:     32-bit ARM (Raspberry Pi Zero/1)"
echo "  Darwin amd64:    macOS Intel"
echo "  Darwin arm64:    macOS Apple Silicon (M1/M2/M3)"
echo "  Windows amd64:   Windows 64-bit"
echo ""
echo "To install on target system:"
echo "  1. Extract the appropriate archive for your OS"
echo "  2. Edit config/config.yaml"
echo "  3. Run ./start.sh (Linux/macOS) or start.bat (Windows)"
