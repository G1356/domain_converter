# Traefik Plugin Catalog Submission Guide

This guide explains how to properly submit your Traefik plugin to the Traefik Plugin Catalog.

## Prerequisites Checklist ✅

- [x] **Git Repository**: Plugin is in a public GitHub repository
- [x] **Go Module**: Properly configured `go.mod` with correct module name
- [x] **Plugin Metadata**: `.traefik.yml` file with plugin information
- [x] **No External Dependencies**: Plugin only uses Go standard library (or properly vendored)
- [ ] **Git Tag**: Version tag following semantic versioning (v1.0.0, v1.0.1, etc.)
- [ ] **Plugin Catalog Submission**: Submit to Traefik Plugin Catalog

## Your Plugin Setup

### Current Configuration
- **Repository**: `https://github.com/G1356/domain_converter`
- **Module Name**: `github.com/G1356/domain_converter`
- **Plugin Type**: Middleware
- **Dependencies**: None (uses only Go standard library)

### Plugin Metadata (`.traefik.yml`)
```yaml
displayName: Domain Converter
type: middleware
import: github.com/G1356/domain_converter
summary: 'Traefik plugin that filters requests based on domain lookup and client IP validation'
testData:
  lookupServiceUrl: http://lookup-service
  defaultTtl: 60
  domainIdHeader: x-domain-id
  urlPath: /api/domain-lookup
```

## Steps to Submit to Plugin Catalog

### 1. Create a Release Tag
The Plugin Catalog requires a proper semantic version tag:

```bash
# Ensure your code is committed and pushed
git add .
git commit -m "Prepare plugin for catalog submission"
git push

# Create a version tag (Plugin Catalog requirement)
git tag -a v1.0.0 -m "Initial release for Traefik Plugin Catalog"
git push origin v1.0.0
```

### 2. Verify Plugin Structure
Your plugin follows the correct structure:
```
domain_converter/
├── .traefik.yml          # Plugin metadata (✅ exists)
├── go.mod                # Go module definition (✅ correct)
├── main.go               # Plugin implementation (✅ exists)
├── main_test.go          # Tests (✅ exists)
└── README.md             # Documentation (✅ exists)
```

### 3. Submit to Plugin Catalog
Visit the [Traefik Plugin Catalog](https://plugins.traefik.io/) and:

1. **Click "Submit a Plugin"**
2. **Enter your repository URL**: `https://github.com/G1356/domain_converter`
3. **Specify the version**: `v1.0.0` (or your latest tag)
4. **Submit for review**

### 4. Monitor for Issues
The Plugin Catalog will:
- Validate your plugin structure
- Check that it compiles
- Test the plugin functionality
- Create GitHub issues if there are problems

## Plugin Catalog Requirements Met

### ✅ **Go Module Proxy Compatible**
- Your plugin uses semantic versioning with git tags
- Module name matches your GitHub repository
- No external dependencies to vendor

### ✅ **Proper Plugin Structure**
- `.traefik.yml` file with correct metadata
- Plugin implements required interfaces
- Tests are included

### ✅ **Documentation**
- README.md with configuration examples
- Clear usage instructions
- Configuration options documented

## Common Plugin Catalog Issues & Solutions

### Issue: "Module not found"
**Solution**: Ensure your git tag is pushed and module name matches repository
```bash
git tag -a v1.0.1 -m "Fix module issues"
git push origin v1.0.1
```

### Issue: "Plugin doesn't compile"
**Solution**: Test compilation locally
```bash
go build .
go test ./...
```

### Issue: "Missing .traefik.yml"
**Solution**: Already handled - you have the file

### Issue: "Dependencies not vendored"
**Solution**: Not applicable - you use only standard library

## Testing Your Plugin Locally

### Test Plugin Compilation
```bash
cd /path/to/domain_converter
go build .
go test ./...
```

### Test with Traefik Locally
Create a test configuration:

```yaml
# docker-compose.yml
version: '3.7'
services:
  traefik:
    image: traefik:v3.0
    command:
      - --experimental.plugins.domain_converter.modulename=github.com/G1356/domain_converter
      - --experimental.plugins.domain_converter.version=v1.0.0
    # ... rest of your Traefik config
```

## Version Management

### Creating New Versions
For updates after initial submission:

```bash
# Make your changes
git add .
git commit -m "Update plugin functionality"
git push

# Create new version tag
git tag -a v1.0.1 -m "Bug fixes and improvements"
git push origin v1.0.1

# Update in Plugin Catalog (if needed)
```

### Version Naming Convention
- **Major**: `v2.0.0` - Breaking changes
- **Minor**: `v1.1.0` - New features, backward compatible
- **Patch**: `v1.0.1` - Bug fixes, backward compatible

## Post-Submission

### Monitor Plugin Status
1. Check your GitHub repository for any issues created by the Plugin Catalog
2. Monitor the Plugin Catalog for your plugin approval
3. Update documentation if requested

### Maintain Your Plugin
- Respond to Plugin Catalog issues quickly
- Keep your plugin updated for new Traefik versions
- Monitor for security issues
- Provide support to users

## Current Status

Your plugin is ready for submission! Next steps:
1. Create and push a `v1.0.0` tag
2. Submit to the Traefik Plugin Catalog
3. Monitor for any feedback or issues

## Support Resources

- [Traefik Plugin Catalog](https://plugins.traefik.io/)
- [Plugin Development Guide](https://doc.traefik.io/traefik-pilot/plugins/plugin-dev/)
- [Traefik Plugin Template](https://github.com/traefik/plugindemo)
- [GitHub Repository](https://github.com/G1356/domain_converter)