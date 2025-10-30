# Example Traefik Configuration

This directory contains example configurations for using the Admin Agency ID Filter plugin with Traefik.

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
    admin-filter:
      plugin:
        header_converter:
          lookupServiceUrl: "http://admin-domain:8080"
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
        - admin-filter
        - security-headers
      service: api-service
      entryPoints:
        - web

    admin-router:
      rule: "Host(`admin.example.com`)"
      middlewares:
        - admin-filter
      service: admin-service
      entryPoints:
        - web

  # Services
  services:
    api-service:
      loadBalancer:
        servers:
          - url: "http://api-backend:3000"

    admin-service:
      loadBalancer:
        servers:
          - url: "http://admin-backend:8080"
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

  admin-domain:
    image: admin-domain-service:latest
    networks:
      - web
    environment:
      - DATABASE_URL=postgresql://user:pass@db:5432/admin
    
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
    admin-filter-dev:
      plugin:
        header_converter:
          lookupServiceUrl: "http://localhost:3001"
          defaultTtl: 60  # Shorter TTL for development
```

### Production
```yaml
http:
  middlewares:
    admin-filter-prod:
      plugin:
        header_converter:
          lookupServiceUrl: "https://admin-domain.internal.company.com"
          defaultTtl: 3600  # Longer TTL for production
```

## Usage Examples

### Basic Usage
```yaml
http:
  routers:
    my-app:
      rule: "Host(`myapp.com`)"
      middlewares:
        - admin-filter
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
        - admin-filter
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
        - admin-filter
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
    admin-filter-debug:
      plugin:
        header_converter:
          lookupServiceUrl: "http://admin-domain"
          defaultTtl: 60
          domainIdHeader: "x-domain-id"
```

### Health Check Route
```yaml
http:
  routers:
    health-check:
      rule: "Host(`myapp.com`) && Path(`/health`)"
      service: health-service
      # No admin-filter middleware for health checks
```