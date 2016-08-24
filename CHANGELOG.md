# Change Log

All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]
### Added
- HTTPClient, a wrapper for http.Client that adds request tracking ID and logs results.
- LogCmd, a wrapper for exec.Cmd that records command execution results together with request tracking ID.

### Changed
- HTTPServer adds request tracking ID to the request context if the request has "X-Cybozu-Request-ID" header.
- Install signal handler only for the global environment.

### Removed
- `Context` method of `Environment` is removed.  It was a design flaw.

## [1.0.1] - 2016-08-22
### Changed
- Update docs.
- Use [cybozu-go/netutil](https://github.com/cybozu-go/netutil).
- Conform to cybozu-go/log v1.1.0 spec.

[Unreleased]: https://github.com/cybozu-go/cmd/compare/v1.0.0...HEAD
[1.0.1]: https://github.com/cybozu-go/cmd/compare/v1.0.0...v1.0.1
