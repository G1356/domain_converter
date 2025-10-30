# Example Traefik Configuration

This directory contains example configurations for using the Doman Converter Filter plugin with Traefik.

## Static Configuration (traefik.yml)

```yaml
# API and dashboard configuration
api:
  dashboard: true
  insecure: true

# Entry points
entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"

# Plugin configuration
experimental:
  plugins:
    header_converter:
      modulename: github.com/yourusername/header_converter
      version: v1.0.0

# Providers
providers:
  file:
    filename: /etc/traefik/dynamic.yml
    watch: true

# Logging
log:
  level: INFO

accessLog: {}
```

## Dynamic Configuration (dynamic.yml)

```yaml
# HTTP Configuration
http:
  # Middlewares
  middlewares:
    domain-lookup:
      plugin:
        header_converter:
          lookupServiceUrl: "http://domain-lookup:8080"
          defaultTtl: 300
          domainIdHeader: "x-domain-id"

    # Additional middlewares can be chained
    security-headers:
      headers:
        customRequestHeaders:
          X-Forwarded-Proto: "https"
        customResponseHeaders:
          X-Frame-Options: "DENY"
          X-Content-Type-Options: "nosniff"

  # Routers
  routers:
    api-router:
      rule: "Host(`api.example.com`)"
      middlewares:
        - domain-lookup
        - security-headers
      service: api-service
      entryPoints:
        - web

    domain-router:
      rule: "Host(`domain.example.com`)"
      middlewares:
        - domain-lookup
      service: domain-service
      entryPoints:
        - web

  # Services
  services:
    api-service:
      loadBalancer:
        servers:
          - url: "http://api-backend:3000"

    domain-service:
      loadBalancer:
        servers:
          - url: "http://domain-backend:8080"
```

## Docker Compose Example

```yaml
version: '3.8'

services:
  traefik:
    image: traefik:v3.0
    command:
      - --configfile=/etc/traefik/traefik.yml
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"  # Dashboard
    volumes:
      - ./traefik.yml:/etc/traefik/traefik.yml
      - ./dynamic.yml:/etc/traefik/dynamic.yml
      - /var/run/docker.sock:/var/run/docker.sock:ro
    networks:
      - web

  domain-lookup:
    image: domain-lookup-service:latest
    networks:
      - web
    environment:
      - DATABASE_URL=postgresql://user:pass@db:5432/dbname
    
  api-backend:
    image: api-service:latest
    networks:
      - web
    environment:
      - DATABASE_URL=postgresql://user:pass@db:5432/api

networks:
  web:
    external: true
```

## Environment-Specific Configurations

### Development
```yaml
http:
  middlewares:
    domain-lookup-dev:
      plugin:
        header_converter:
          lookupServiceUrl: "http://localhost:3001"
          defaultTtl: 60  # Shorter TTL for development
          urlPath: /api/domain-lookup
```

### Production
```yaml
http:
  middlewares:
    domain-lookup-prod:
      plugin:
        header_converter:
          lookupServiceUrl: "https://domain-lookup.internal.company.com"
          defaultTtl: 3600  # Longer TTL for production
          urlPath: /api/domain-lookup
```

## Usage Examples

### Basic Usage
```yaml
http:
  routers:
    my-app:
      rule: "Host(`myapp.com`)"
      middlewares:
        - domain-lookup
      service: my-app-service
```

### With Rate Limiting
```yaml
http:
  middlewares:
    rate-limit:
      rateLimit:
        burst: 100
        average: 50

  routers:
    protected-api:
      rule: "Host(`api.myapp.com`)"
      middlewares:
        - domain-lookup
        - rate-limit
      service: api-service
```

### With CORS
```yaml
http:
  middlewares:
    cors:
      headers:
        accessControlAllowMethods:
          - GET
          - POST
          - PUT
          - DELETE
        accessControlAllowOriginList:
          - "https://myapp.com"
        accessControlAllowHeaders:
          - "Content-Type"
          - "Authorization"
          - "x-agency-id"

  routers:
    api-with-cors:
      rule: "Host(`api.myapp.com`)"
      middlewares:
        - domain-lookup
        - cors
      service: api-service
```

## Troubleshooting

### Enable Debug Logging
```yaml
log:
  level: DEBUG

# Add to middleware configuration for detailed logging
http:
  middlewares:
    domain-lookup-debug:
      plugin:
        header_converter:
          lookupServiceUrl: "http://domain-lookup"
          defaultTtl: 60
          domainIdHeader: "x-domain-id"
          urlPath: /api/domain-lookup
```

### Health Check Route
```yaml
http:
  routers:
    health-check:
      rule: "Host(`myapp.com`) && Path(`/health`)"
      service: health-service
      # No domain-lookup middleware for health checks
```