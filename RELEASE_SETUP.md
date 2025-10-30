# Release Setup Guide

This guide explains how to set up automated binary releases for your Go project using GitHub Actions.

## What Was Set Up

### 1. CLI Application Structure
- Created `cmd/domain_converter/main.go` - A CLI wrapper for your domain converter
- Updated `Makefile` with cross-platform build targets
- Added version support via build-time ldflags

### 2. GitHub Actions Workflow
- **File**: `.github/workflows/release.yml`
- **Triggers**: When you push a tag matching `v*.*.*` (e.g., `v1.0.0`)
- **Builds for**:
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)  
  - Windows (amd64)

### 3. Release Process
The workflow automatically:
- Builds binaries for all target platforms
- Creates compressed archives (`.tar.gz` for Unix, `.zip` for Windows)
- Creates a GitHub release with all binaries attached
- Uses changelog entries if `CHANGELOG.md` exists

## How to Create a Release

### Step 1: Prepare Your Release
1. Update `CHANGELOG.md` with your changes
2. Test your code: `make build-cli && ./domain_converter --version`
3. Commit and push your changes

### Step 2: Create and Push a Tag
```bash
# Create a new tag (replace with your version)
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag to trigger the release workflow
git push origin v1.0.0
```

### Step 3: Monitor the Release
1. Go to your GitHub repository
2. Click on "Actions" tab
3. Watch the "Release" workflow complete
4. Check the "Releases" section for your new release

## Manual Building

### Build for Current Platform
```bash
make build-cli
```

### Build for All Platforms
```bash
make build-all
```

### Create Release Archives
```bash
make release-build
```

## File Structure Created

```
├── .github/workflows/
│   ├── ci.yml          # Existing CI workflow
│   └── release.yml     # New release workflow
├── cmd/
│   └── domain_converter/
│       └── main.go     # CLI application entry point
├── CHANGELOG.md        # Release notes template
├── Makefile           # Updated with build targets
└── .gitignore         # Updated to ignore build artifacts
```

## Usage Examples

Once released, users can download and use your binary:

### Linux/macOS
```bash
# Download latest release
curl -LO https://github.com/yourusername/domain_converter/releases/latest/download/domain_converter_linux_amd64.tar.gz

# Extract and use
tar -xzf domain_converter_linux_amd64.tar.gz
./domain_converter --help
```

### Windows
```powershell
# Download from GitHub releases page or use:
Invoke-WebRequest -Uri "https://github.com/yourusername/domain_converter/releases/latest/download/domain_converter_windows_amd64.zip" -OutFile "domain_converter.zip"
Expand-Archive -Path "domain_converter.zip" -DestinationPath "."
.\domain_converter.exe --help
```

## Customization Options

### Change Build Targets
Edit the `PLATFORMS` variable in `Makefile`:
```makefile
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64
```

### Modify Release Workflow
Edit `.github/workflows/release.yml`:
- Add/remove build platforms in the `matrix.goos` and `matrix.goarch` sections
- Customize the binary naming convention
- Add additional build steps or tests

### Version Information
The binary includes version info set during build:
- Automatic version from git tags when building via GitHub Actions
- Manual version: `make build-cli VERSION=v1.2.3`

## Troubleshooting

### Release Not Created
- Ensure your tag follows the pattern `v*.*.*` (e.g., `v1.0.0`)
- Check the Actions tab for workflow errors
- Verify you have write permissions to the repository

### Build Failures
- Check that your code compiles: `go build ./cmd/domain_converter`
- Ensure all dependencies are properly vendored: `go mod tidy`
- Test cross-compilation locally: `GOOS=linux GOARCH=amd64 go build ./cmd/domain_converter`

### Large Binary Size
- The binaries are built with `-ldflags="-s -w"` to strip debug info
- Consider using UPX compression for even smaller binaries
- Review dependencies for unnecessary imports

## Next Steps

1. **Update Repository URL**: Replace `yourusername` in `go.mod` and documentation with your actual GitHub username
2. **Test the Workflow**: Create a test tag to verify everything works
3. **Configure Branch Protection**: Ensure releases only happen from protected branches
4. **Add Security Scanning**: Consider adding security checks to your workflow
5. **Documentation**: Update your main README.md with installation instructions

## Security Considerations

- The workflow uses `GITHUB_TOKEN` which is automatically provided
- No external secrets are required for basic releases
- Consider signing your releases for additional security
- Review the permissions granted to the workflow in the `permissions` section