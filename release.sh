#!/bin/bash

# Release script for Traefik Plugin Catalog submission
# This script helps create a proper release tag for plugin catalog submission

set -e

VERSION="v1.0.0"
MESSAGE="Initial release for Traefik Plugin Catalog"

echo "🚀 Preparing release for Traefik Plugin Catalog submission..."
echo "Version: $VERSION"
echo

# Check if working directory is clean
if [[ -n $(git status --porcelain) ]]; then
    echo "❌ Working directory is not clean. Please commit or stash your changes first."
    git status --short
    exit 1
fi

# Check if we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" != "main" ]]; then
    echo "⚠️  Warning: You're not on the main branch (current: $CURRENT_BRANCH)"
    read -p "Do you want to continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo "❌ Tag $VERSION already exists!"
    echo "Existing tags:"
    git tag -l "v*" | sort -V
    exit 1
fi

# Run tests before releasing
echo "🧪 Running tests..."
go test ./...
if [[ $? -ne 0 ]]; then
    echo "❌ Tests failed! Please fix the issues before releasing."
    exit 1
fi

# Build plugin to ensure it compiles
echo "🔨 Building plugin..."
go build .
if [[ $? -ne 0 ]]; then
    echo "❌ Build failed! Please fix the issues before releasing."
    exit 1
fi

echo "✅ All checks passed!"
echo

# Create and push tag
echo "📝 Creating release tag..."
git tag -a "$VERSION" -m "$MESSAGE"

echo "📤 Pushing tag to origin..."
git push origin "$VERSION"

echo
echo "🎉 Release $VERSION created successfully!"
echo
echo "Next steps:"
echo "1. Visit https://plugins.traefik.io/"
echo "2. Click 'Submit a Plugin'"
echo "3. Enter repository URL: https://github.com/G1356/domain_converter"
echo "4. Specify version: $VERSION"
echo "5. Submit for review"
echo
echo "📋 Plugin Information:"
echo "   Repository: https://github.com/G1356/domain_converter"
echo "   Module: github.com/G1356/domain_converter"
echo "   Version: $VERSION"
echo "   Type: Middleware"
echo
echo "⚠️  Monitor your GitHub repository for any issues created by the Plugin Catalog!"