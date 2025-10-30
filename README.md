# Admin Agency ID Filter Traefik Plugin

This Traefik plugin converts the functionality from the original WebAssembly proxy code to filter HTTP requests based on domain lookup and client IP validation.

## Features

- **Domain-to-Agency-ID Lookup**: Queries an admin service to get agency information for incoming domains
- **Client IP Validation**: Validates that the client IP is allowed for the specific domain
- **Caching**: In-memory caching of lookup results with configurable TTL
- **Redirect Support**: Handles redirect responses (HTTP 201) from the admin service
- **Request Header Injection**: Adds `x-agency-id` header to validated requests

## Configuration

### Static Configuration

```yaml
pilot:
  token: "your-pilot-token"

experimental:
  plugins:
    domain_converter:
      modulename: github.com/yourusername/domain_converter
      version: v1.0.0
```

### Dynamic Configuration

```yaml
http:
  middlewares:
    admin-filter:
      plugin:
        domain_converter:
          lookupServiceUrl: "http://domain-lookup"
          defaultTtl: 60
          domainIdHeader: "x-domain-id"
          urlPath: /api/domain-lookup

  routers:
    my-router:
      rule: "Host(`example.com`)"
      middlewares:
        - admin-filter
      service: my-service
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `lookupServiceUrl` | string | `http://domain-lookup` | Base URL of the admin lookup service |
| `defaultTtl` | int | `60` | Default cache TTL in seconds when no Cache-Control header is present |
| `domainIdHeader` | string | `x-domain-id` | DomainHeader to pass id |
| `urlPath` | string | `/api/domain-lookup` | Url path that will be called |

## API Contract

The plugin expects the admin service to respond to GET requests at urlPath set at config with:

### Success Response (HTTP 200)
- **Body**: `{uuid}|{ip1,ip2,ip3}` - Agency UUID followed by comma-separated allowed IPs
- **Headers**: Optional `Cache-Control: max-age={seconds}` for caching

### Redirect Response (HTTP 201)
- **Body**: Redirect URL
- **Headers**: Optional `Cache-Control: max-age={seconds}` for caching

### Not Found (HTTP 404)
- Domain not found, cached for `defaultTtl` seconds

## Behavior

1. **Cache Check**: First checks if domain info is cached and not expired
2. **IP Validation**: Extracts client IP from `X-Forwarded-For` header (first IP) or `RemoteAddr`
3. **Domain Lookup**: If not cached, queries the admin service
4. **Response Handling**:
   - **200**: Validates IP and sets `x-agency-id` header if allowed
   - **201**: Redirects to the provided URL
   - **404**: Returns 401 Unauthorized and caches negative result
   - **Other**: Returns 500 Internal Server Error

## Client IP Extraction

The plugin extracts client IP in the following order:
1. First IP from `X-Forwarded-For` header (trimmed of whitespace)
2. Fallback to `RemoteAddr` (without port)

## Caching

- Uses in-memory cache with read-write mutex for thread safety
- Cache keys are the hostname/domain
- TTL is determined by `Cache-Control: max-age` header or `defaultTtl` config
- Expired entries are automatically removed on access
- Supports caching of negative results (404 responses)

## Error Handling

- Network errors during admin service lookup allow the request to continue
- Invalid responses return appropriate HTTP error codes
- Malformed domain info is handled gracefully

## Installation

### Using as a Traefik Plugin

1. Add the plugin to your Traefik configuration
2. Configure the middleware in your dynamic configuration
3. Apply the middleware to your routers

### Using as a Standalone CLI

Download the latest binary from the [releases page](https://github.com/yourusername/domain_converter/releases).

#### Linux/macOS
```bash
# Download and extract
wget https://github.com/yourusername/domain_converter/releases/latest/download/domain_converter_linux_amd64.tar.gz
tar -xzf domain_converter_linux_amd64.tar.gz
chmod +x domain_converter

# Run the service
./domain_converter --port 8080 --lookup-url http://your-domain-service
```

#### Windows
```powershell
# Download from releases page or use PowerShell
Invoke-WebRequest -Uri "https://github.com/yourusername/domain_converter/releases/latest/download/domain_converter_windows_amd64.zip" -OutFile "domain_converter.zip"
Expand-Archive -Path "domain_converter.zip" -DestinationPath "."

# Run the service
.\domain_converter.exe --port 8080 --lookup-url http://your-domain-service
```

#### CLI Options
```bash
./domain_converter --help
  -header string
        Header name for domain ID (default "x-domain-id")
  -lookup-url string
        URL for domain lookup service (default "http://domain")
  -urlPath string
        URL Path for domain lookup service (default "/api/domain-lookup")
  -port string
        Port to run the server on (default "8080")
  -ttl int
        Default TTL for cache entries (default 60)
  -version
        Show version information
```

## Releases

This project uses automated releases with GitHub Actions. Binary releases are available for:

- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)  
- **Windows**: amd64

### Creating a Release

1. Update the `CHANGELOG.md` with your changes
2. Create and push a tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. GitHub Actions will automatically build and create the release

### Manual Build

For local development and testing:

```bash
# Build for current platform
make build-cli

# Build for all platforms
make build-all

# Create release archives
make release-build

# View version info
./domain_converter --version
```

## Development

To build and test locally:

```bash
go mod tidy
go build
go test
```

## Migration from WebAssembly

This plugin replicates the functionality of the original Rust WebAssembly proxy with the following key differences:

- Uses Go's standard HTTP client instead of WebAssembly HTTP dispatch
- Implements thread-safe caching using sync.RWMutex
- Integrates with Traefik's middleware chain
- Provides configuration through Traefik's standard config system