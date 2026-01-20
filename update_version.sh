#!/bin/bash
# Version Update Script for Server Dashboard
# Usage: ./update_version.sh [major|minor|patch]

VERSION_FILE="VERSION"

if [ ! -f "$VERSION_FILE" ]; then
    echo "v1.0.0" > "$VERSION_FILE"
    echo "Created VERSION file with v1.0.0"
    exit 0
fi

# Read current version
CURRENT_VERSION=$(cat "$VERSION_FILE" | tr -d '[:space:]')
echo "Current version: $CURRENT_VERSION"

# Remove 'v' prefix if present
VERSION_NUM="${CURRENT_VERSION#v}"

# Split version into parts
IFS='.' read -r -a VERSION_PARTS <<< "$VERSION_NUM"
MAJOR="${VERSION_PARTS[0]}"
MINOR="${VERSION_PARTS[1]}"
PATCH="${VERSION_PARTS[2]}"

# Determine what to update
UPDATE_TYPE="${1:-patch}"

case "$UPDATE_TYPE" in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        echo "Bumping MAJOR version"
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        echo "Bumping MINOR version"
        ;;
    patch)
        PATCH=$((PATCH + 1))
        echo "Bumping PATCH version"
        ;;
    *)
        echo "Usage: $0 [major|minor|patch]"
        echo "  major - Breaking changes (x.0.0)"
        echo "  minor - New features (0.x.0)"
        echo "  patch - Bug fixes (0.0.x)"
        exit 1
        ;;
esac

# Create new version
NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"
echo "$NEW_VERSION" > "$VERSION_FILE"

echo "Updated version: $CURRENT_VERSION → $NEW_VERSION"
echo ""
echo "Next steps:"
echo "  1. Review changes: git diff"
echo "  2. Commit: git add VERSION && git commit -m 'Bump version to $NEW_VERSION'"
echo "  3. Tag: git tag -a $NEW_VERSION -m 'Release $NEW_VERSION'"
echo "  4. Build: ./build.sh"
echo "  5. Push: git push && git push --tags"
