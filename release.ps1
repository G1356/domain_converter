# Release script for Traefik Plugin Catalog submission
# PowerShell version for Windows

param(
    [string]$Version = "v1.0.0",
    [string]$Message = "Initial release for Traefik Plugin Catalog"
)

Write-Host "🚀 Preparing release for Traefik Plugin Catalog submission..." -ForegroundColor Green
Write-Host "Version: $Version"
Write-Host ""

# Check if working directory is clean
$gitStatus = git status --porcelain
if ($gitStatus) {
    Write-Host "❌ Working directory is not clean. Please commit or stash your changes first." -ForegroundColor Red
    git status --short
    exit 1
}

# Check if we're on main branch
$currentBranch = git branch --show-current
if ($currentBranch -ne "main") {
    Write-Host "⚠️  Warning: You're not on the main branch (current: $currentBranch)" -ForegroundColor Yellow
    $continue = Read-Host "Do you want to continue? (y/N)"
    if ($continue -notmatch "^[Yy]$") {
        exit 1
    }
}

# Check if tag already exists
try {
    git rev-parse $Version 2>$null
    Write-Host "❌ Tag $Version already exists!" -ForegroundColor Red
    Write-Host "Existing tags:"
    git tag -l "v*"
    exit 1
} catch {
    # Tag doesn't exist, which is what we want
}

# Run tests before releasing
Write-Host "🧪 Running tests..." -ForegroundColor Cyan
$testResult = go test ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Tests failed! Please fix the issues before releasing." -ForegroundColor Red
    exit 1
}

# Build plugin to ensure it compiles
Write-Host "🔨 Building plugin..." -ForegroundColor Cyan
$buildResult = go build .
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed! Please fix the issues before releasing." -ForegroundColor Red
    exit 1
}

Write-Host "✅ All checks passed!" -ForegroundColor Green
Write-Host ""

# Create and push tag
Write-Host "📝 Creating release tag..." -ForegroundColor Cyan
git tag -a $Version -m $Message

Write-Host "📤 Pushing tag to origin..." -ForegroundColor Cyan
git push origin $Version

Write-Host ""
Write-Host "🎉 Release $Version created successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Visit https://plugins.traefik.io/"
Write-Host "2. Click 'Submit a Plugin'"
Write-Host "3. Enter repository URL: https://github.com/G1356/domain_converter"
Write-Host "4. Specify version: $Version"
Write-Host "5. Submit for review"
Write-Host ""
Write-Host "📋 Plugin Information:" -ForegroundColor Cyan
Write-Host "   Repository: https://github.com/G1356/domain_converter"
Write-Host "   Module: github.com/G1356/domain_converter"
Write-Host "   Version: $Version"
Write-Host "   Type: Middleware"
Write-Host ""
Write-Host "⚠️  Monitor your GitHub repository for any issues created by the Plugin Catalog!" -ForegroundColor Yellow