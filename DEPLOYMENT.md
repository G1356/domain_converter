# Deployment Guide

This guide explains how to deploy the Admin Agency ID Filter Traefik plugin.

## Prerequisites

- Traefik v2.10+ or v3.0+
- Admin service running and accessible
- Go 1.20+ (for development)

## Deployment Methods

### Method 1: Traefik Pilot (Recommended for Production)

1. **Publish to Traefik Pilot**
   ```bash
   # Create a GitHub repository for your plugin
   git init
   git add .
   git commit -m "Initial commit"
   git remote add origin https://github.com/yourusername/header_converter.git
   git push -u origin main
   ```

2. **Configure Traefik Static Configuration**
   ```yaml
   # traefik.yml
   pilot:
     token: "your-pilot-token"
   
   experimental:
     plugins:
       header_converter:
         modulename: github.com/yourusername/header_converter
         version: v1.0.0
   ```

### Method 2: Local Development

1. **Clone/Copy the Plugin**
   ```bash
   # Place the plugin in Traefik's plugins directory
   mkdir -p /plugins-local/src/github.com/yourusername/header_converter
   cp -r . /plugins-local/src/github.com/yourusername/header_converter/
   ```

2. **Configure Traefik Static Configuration**
   ```yaml
   # traefik.yml
   experimental:
     localPlugins:
       header_converter:
         modulename: github.com/yourusername/header_converter
   ```

### Method 3: Docker Development

1. **Create a Dockerfile for Development**
   ```dockerfile
   FROM traefik:v3.0
   
   # Copy the plugin source
   COPY . /plugins-local/src/github.com/yourusername/header_converter/
   
   # Install Go (for plugin compilation)
   RUN apk add --no-cache go git
   
   # Set up plugin
   WORKDIR /plugins-local/src/github.com/yourusername/header_converter
   RUN go mod tidy
   ```

## Configuration Steps

### 1. Static Configuration

Choose one of the following based on your deployment method:

**For Traefik Pilot:**
```yaml
# traefik.yml
api:
  dashboard: true
  insecure: true  # Only for development

entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"

pilot:
  token: "${TRAEFIK_PILOT_TOKEN}"

experimental:
  plugins:
    header_converter:
      modulename: github.com/yourusername/header_converter
      version: v1.0.0

providers:
  file:
    filename: /etc/traefik/dynamic.yml
    watch: true
```

**For Local Development:**
```yaml
# traefik.yml
api:
  dashboard: true
  insecure: true

entryPoints:
  web:
    address: ":80"

experimental:
  localPlugins:
    header_converter:
      modulename: github.com/yourusername/header_converter

providers:
  file:
    filename: /etc/traefik/dynamic.yml
    watch: true
```

### 2. Dynamic Configuration

```yaml
# dynamic.yml
http:
  middlewares:
    admin-filter:
      plugin:
        header_converter:
          adminServiceUrl: "http://admin-domain:8080"
          defaultTtl: 300

  routers:
    api-router:
      rule: "Host(`api.example.com`)"
      middlewares:
        - admin-filter
      service: api-service

  services:
    api-service:
      loadBalancer:
        servers:
          - url: "http://backend:3000"
```

## Environment Variables

Set these environment variables for your deployment:

```bash
# Required
TRAEFIK_PILOT_TOKEN=your_pilot_token_here

# Optional - Override in dynamic config
ADMIN_SERVICE_URL=http://admin-domain:8080
DEFAULT_TTL=300
```

## Docker Compose Example

```yaml
version: '3.8'

services:
  traefik:
    image: traefik:v3.0
    container_name: traefik
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"
    environment:
      - TRAEFIK_PILOT_TOKEN=${TRAEFIK_PILOT_TOKEN}
    volumes:
      - ./traefik.yml:/etc/traefik/traefik.yml:ro
      - ./dynamic.yml:/etc/traefik/dynamic.yml:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
      # For local development, uncomment the next line:
      # - ./:/plugins-local/src/github.com/yourusername/header_converter:ro
    networks:
      - web

  admin-domain:
    image: your-admin-service:latest
    container_name: admin-domain
    restart: unless-stopped
    networks:
      - web
    environment:
      - DATABASE_URL=postgresql://user:pass@db:5432/admin

  backend:
    image: your-backend:latest
    container_name: backend
    restart: unless-stopped
    networks:
      - web
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.backend.rule=Host(`api.example.com`)"
      - "traefik.http.routers.backend.middlewares=admin-filter"

networks:
  web:
    external: true
```

## Kubernetes Deployment

### 1. Create ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: traefik-config
data:
  traefik.yml: |
    api:
      dashboard: true
    entryPoints:
      web:
        address: ":80"
    pilot:
      token: "your-pilot-token"
    experimental:
      plugins:
        header_converter:
          modulename: github.com/yourusername/header_converter
          version: v1.0.0
    providers:
      kubernetescrd: {}
```

### 2. Create Middleware CRD

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: admin-filter
spec:
  plugin:
    header_converter:
      adminServiceUrl: "http://admin-domain.admin.svc.cluster.local:8080"
      defaultTtl: 300
```

### 3. Apply to IngressRoute

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: api-route
spec:
  entryPoints:
    - web
  routes:
    - match: Host(`api.example.com`)
      kind: Rule
      middlewares:
        - name: admin-filter
      services:
        - name: api-service
          port: 80
```

## Verification

### 1. Check Plugin Loading

```bash
# Check Traefik logs for plugin loading
docker logs traefik 2>&1 | grep -i "header_converter"

# Expected output:
# time="..." level=info msg="Loading plugin header_converter..."
```

### 2. Test the Plugin

```bash
# Test with a valid domain
curl -H "Host: api.example.com" \
     -H "X-Forwarded-For: 192.168.1.1" \
     http://localhost/test

# Check for x-agency-id header in backend logs
# or use a debug endpoint to verify headers
```

### 3. Monitor Admin Service Calls

```bash
# Check admin service logs for lookup requests
# Expected: GET /api/admin-domain/domain-to-agency-id?domain=api.example.com
```

## Troubleshooting

### Common Issues

1. **Plugin Not Loading**
   ```bash
   # Check Traefik configuration
   docker exec traefik traefik version
   
   # Verify plugin configuration in Traefik logs
   docker logs traefik
   ```

2. **Admin Service Connection Issues**
   ```bash
   # Test admin service connectivity
   docker exec traefik wget -qO- http://admin-domain:8080/health
   
   # Check DNS resolution
   docker exec traefik nslookup admin-domain
   ```

3. **Cache Issues**
   ```bash
   # Plugin uses in-memory cache
   # Restart Traefik to clear cache if needed
   docker restart traefik
   ```

### Debug Mode

Enable debug logging in Traefik:

```yaml
# traefik.yml
log:
  level: DEBUG
  
accessLog:
  filePath: "/var/log/traefik/access.log"
```

### Health Checks

Add health check endpoints that bypass the plugin:

```yaml
# dynamic.yml
http:
  routers:
    health-router:
      rule: "Host(`api.example.com`) && Path(`/health`)"
      service: api-service
      # No middleware - bypasses admin filter
```

## Performance Considerations

1. **Cache TTL**: Set appropriate TTL values based on your admin service's update frequency
2. **Admin Service Timeout**: Default is 1 second, adjust based on network latency
3. **Memory Usage**: Cache is in-memory per Traefik instance
4. **High Availability**: Consider cache coherence across multiple Traefik instances

## Security Notes

1. **Admin Service Security**: Ensure admin service is only accessible from Traefik
2. **IP Validation**: Plugin validates client IPs from X-Forwarded-For header
3. **Error Handling**: Failed admin service calls allow requests to continue (fail-open)
4. **Cache Security**: Cached data is not encrypted in memory

## Rollback Plan

If issues occur:

1. **Disable Plugin**: Remove middleware from routers
2. **Fallback Configuration**: Have a backup dynamic.yml without the plugin
3. **Quick Rollback**: Use Traefik's file provider watch feature for instant updates

```bash
# Quick disable
cp dynamic-no-plugin.yml dynamic.yml
# Traefik will reload automatically
```