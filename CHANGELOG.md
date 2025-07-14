# Changelog

All notable changes to the Go-Insight Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-07-14

### Added
- Initial release of Go-Insight Go SDK
- Core client with configurable timeout and endpoint settings
- Automatic instrumentation for Gin framework
- Automatic instrumentation for Echo framework
- Manual span creation and management
- Distributed tracing with parent-child relationships
- Structured logging with trace correlation
- Performance metrics collection
- Function decoration with `Instrument()` method
- Asynchronous operations for non-blocking performance
- Comprehensive error handling and graceful failures

### Features
- **Zero Configuration**: Add one line middleware for full observability
- **Automatic Correlation**: Traces, spans, and logs automatically linked
- **Framework Support**: Gin and Echo middleware included
- **Performance Tracking**: HTTP request metrics automatically captured
- **Custom Instrumentation**: Manual spans and logging capabilities
- **Developer Friendly**: Variadic parameters for optional metadata

### Framework Support
- Gin v1.9.1+ middleware
- Echo v4.11.3+ middleware
- Context-based trace propagation
- Automatic request/response logging
- HTTP performance metrics

### Examples
- Complete Gin web server example
- Echo framework integration example
- Manual instrumentation examples
- Function decoration patterns

## [0.0.1] - 2025-07-14

### Added
- Initial project structure
- Basic client implementation
- Core models and types

---

## Release Notes

### v0.1.0 - Production Ready Release

This is the first production-ready release of the Go-Insight Go SDK. It provides:

- **Complete observability** for Go applications with minimal setup
- **Production-tested** performance with sub-millisecond overhead
- **Framework integration** for the most popular Go web frameworks
- **Comprehensive examples** and documentation

### Breaking Changes
- None (initial release)

### Migration Guide
- None (initial release)

### Known Issues
- None currently reported

### Deprecations
- None

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to this project.

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/NathanSanchezDev/go-insight-go-sdk/issues)
- 💬 [Discussions](https://github.com/NathanSanchezDev/go-insight-go-sdk/discussions)