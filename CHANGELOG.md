# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- N/A

### Changed
- N/A

### Deprecated
- N/A

### Removed
- N/A

### Fixed
- N/A

### Security
- N/A

## [v1.0.0] - 2025-10-30

### Added
- Initial release of domain converter Traefik plugin
- Domain-to-Agency-ID lookup functionality
- Client IP validation against allowed IP lists
- In-memory caching with configurable TTL
- Support for redirect responses (HTTP 201)
- Request header injection (`x-agency-id`)
- Comprehensive error handling and logging
- CLI application wrapper for standalone usage
- Cross-platform binary builds via GitHub Actions
- Automated release pipeline
- Complete documentation and examples

### Configuration Options
- `lookupServiceUrl`: URL for domain lookup service
- `defaultTtl`: Default cache TTL in seconds
- `domainIdHeader`: Header name for domain ID
- `urlPath`: URL path for domain lookup API

[Unreleased]: https://github.com/G1356/domain_converter/compare/v1.0.0...HEAD
[v1.0.0]: https://github.com/G1356/domain_converter/releases/tag/v1.0.0