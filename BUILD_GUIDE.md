# Build & Release Guide

## Version Management

### Current Version
The current version is stored in the `VERSION` file in the project root.

```bash
cat VERSION
# Output: v1.0.0
```

### Updating Version

Use the `update_version.sh` script to bump the version:

```bash
# Patch release (bug fixes): v1.0.0 → v1.0.1
./update_version.sh patch

# Minor release (new features): v1.0.0 → v1.1.0
./update_version.sh minor

# Major release (breaking changes): v1.0.0 → v2.0.0
./update_version.sh major
```

The script will:
1. Read the current version from `VERSION` file
2. Increment the appropriate number (major, minor, or patch)
3. Update the `VERSION` file
4. Display the new version and next steps

### Version Format

We follow **Semantic Versioning** (semver):
- **MAJOR** version: Incompatible API changes
- **MINOR** version: New functionality (backwards compatible)
- **PATCH** version: Bug fixes (backwards compatible)

Example: `v1.2.3`
- Major: 1
- Minor: 2
- Patch: 3

## Building the Application

### Quick Build (Current Platform)

```bash
# Simple build for your current OS
go build -o server-dashboard ./cmd/main.go

# Build with version information
go build -ldflags "-X main.Version=$(cat VERSION)" -o server-dashboard ./cmd/main.go
```

### Multi-Platform Build

The `build.sh` script creates release bundles for all major platforms:

```bash
./build.sh
```

This will:
1. Read version from `VERSION` file
2. Build binaries for all platforms:
   - Linux (amd64, arm64)
   - Linux ARM (armv6, armv7) - **Raspberry Pi support**
   - macOS (amd64, arm64 - Intel & Apple Silicon)
   - Windows (amd64)
3. Create distribution bundles with:
   - Binary executable
   - Configuration files
   - Web assets (HTML, CSS, JS)
   - Documentation
   - Sample environment file
   - Platform-specific startup scripts
4. Package everything into archives:
   - `.tar.gz` for Linux/macOS
   - `.zip` for Windows

### Build Output

After running `./build.sh`, you'll find:

```
dist/
├── server-dashboard-v1.0.0-linux-amd64.tar.gz
├── server-dashboard-v1.0.0-linux-arm64.tar.gz
├── server-dashboard-v1.0.0-linux-armv6.tar.gz      # Raspberry Pi Zero/1
├── server-dashboard-v1.0.0-linux-armv7.tar.gz      # Raspberry Pi 2/3/4 (32-bit)
├── server-dashboard-v1.0.0-darwin-amd64.tar.gz
├── server-dashboard-v1.0.0-darwin-arm64.tar.gz
└── server-dashboard-v1.0.0-windows-amd64.zip
```

Each bundle contains:
```
server-dashboard-v1.0.0-linux-amd64/
├── server-dashboard-linux-amd64    # Binary
├── start.sh                        # Startup script
├── config/
│   └── config.yaml                 # Configuration
├── web/
│   ├── static/                     # CSS, JS, images
│   └── templates/                  # HTML templates
├── .env.example                    # Environment variables template
├── README.txt                      # Quick start guide
├── README.md                       # Full documentation
├── DISK_MONITORING.md              # Disk monitoring guide
└── VERSION                         # Version file
```

## Release Process

### 1. Update Version

```bash
# Example: releasing v1.1.0
./update_version.sh minor
```

### 2. Commit Version Change

```bash
git add VERSION
git commit -m "Bump version to v1.1.0"
```

### 3. Build Release Artifacts

```bash
./build.sh
```

### 4. Test the Build

Extract and test one of the distribution packages:

```bash
# Extract
cd dist
tar -xzf server-dashboard-v1.1.0-darwin-arm64.tar.gz
cd server-dashboard-v1.1.0-darwin-arm64

# Test
./start.sh
```

### 5. Create Git Tag

```bash
# Get version
VERSION=$(cat VERSION)

# Create annotated tag
git tag -a $VERSION -m "Release $VERSION"

# Verify tag
git tag -l -n1 $VERSION
```

### 6. Push to Repository

```bash
# Push commits
git push

# Push tags
git push --tags
```

### 7. Create GitHub Release (if using GitHub)

```bash
# Using GitHub CLI (gh)
gh release create $(cat VERSION) \
  dist/*.tar.gz dist/*.zip \
  --title "Release $(cat VERSION)" \
  --notes "Release notes here"
```

Or manually:
1. Go to GitHub repository → Releases → New Release
2. Select the tag you just pushed
3. Upload files from `dist/` directory
4. Add release notes
5. Publish release

## Build Flags & Variables

The build process injects version information at compile time:

```go
// Set via -ldflags during build
var (
    Version   = "dev"        // From VERSION file or git tag
    BuildTime = "unknown"    // UTC timestamp
    GitCommit = "unknown"    // Short git commit hash
)
```

These are displayed in:
- Application logs at startup
- Footer of web interface
- Can be queried via API (future feature)

## Platform-Specific Notes

### Linux (x86_64)
- Binaries are statically linked (no external dependencies)
- amd64 (x86_64) and arm64 (aarch64) supported
- Tested on Ubuntu 20.04+, Debian 11+, CentOS 8+

### Raspberry Pi
#### 64-bit OS (Recommended)
- Use `linux-arm64` build
- Supported: Raspberry Pi 3, 4, 5 running 64-bit Raspberry Pi OS
- Better performance and native 64-bit support

#### 32-bit OS
- **ARMv7** (`linux-armv7`): Raspberry Pi 2, 3, 4 with 32-bit OS
- **ARMv6** (`linux-armv6`): Raspberry Pi Zero, Zero W, Pi 1

#### Installation on Raspberry Pi
```bash
# Download the appropriate build
wget https://github.com/youruser/server-dashboard/releases/download/v1.0.0/server-dashboard-v1.0.0-linux-armv7.tar.gz

# Extract
tar -xzf server-dashboard-v1.0.0-linux-armv7.tar.gz
cd server-dashboard-v1.0.0-linux-armv7

# Make executable (if needed)
chmod +x server-dashboard-linux-armv7

# Configure
nano config/config.yaml

# Run
./start.sh
```

#### Performance Notes
- Raspberry Pi 3/4/5 with 64-bit OS: Best performance (use arm64)
- Raspberry Pi 3/4 with 32-bit OS: Good performance (use armv7)
- Raspberry Pi Zero/1: Limited resources (use armv6, may be slow)

### macOS
- Universal binaries for Intel and Apple Silicon
- Code-signed (optional, for distribution)
- Tested on macOS 11 (Big Sur) and newer

### Windows
- Binaries work on Windows 10/11 and Server 2016+
- No external DLLs required
- Can run as Windows Service (setup required)

## Troubleshooting Builds

### Build Fails

```bash
# Clean and retry
rm -rf build dist
go clean
go mod tidy
./build.sh
```

### Version Not Showing

```bash
# Make sure VERSION file exists
cat VERSION

# Rebuild with explicit version
go build -ldflags "-X main.Version=$(cat VERSION)" -o server-dashboard ./cmd/main.go
```

### Missing Dependencies

```bash
# Update dependencies
go mod download
go mod verify
```

## Development Builds

For development, skip the full build process:

```bash
# Quick compile
go build -o server-dashboard ./cmd/main.go

# Run directly (no binary)
go run cmd/main.go

# With race detection
go run -race cmd/main.go
```

## CI/CD Integration

Example GitHub Actions workflow (`.github/workflows/release.yml`):

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.18'
      
      - name: Build
        run: ./build.sh
      
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*
```

## Version Display

The version is automatically displayed in:

1. **Application Logs**
   ```
   2026/01/20 13:42:02 Server Dashboard v1.0.0
   ```

2. **Web Interface Footer**
   - Bottom right: Version number with icon
   - Sticky footer always visible

3. **Template Functions**
   ```html
   {{ appVersion }}    <!-- v1.0.0 -->
   {{ buildInfo }}     <!-- v1.0.0 (built 2026-01-20_13:42:02) -->
   ```

## Quick Reference

```bash
# Update version
./update_version.sh [major|minor|patch]

# Build all platforms
./build.sh

# Build current platform only
go build -o server-dashboard ./cmd/main.go

# View current version
cat VERSION

# Test binary
./server-dashboard
```
