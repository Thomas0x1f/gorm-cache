# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.0.2] - 2026-01-09

### Added
- Empty query result detection to prevent caching zero-row results
- Comprehensive tests for empty result handling and cache behavior

### Changed
- Refactored empty result check to use `RowsAffected` instead of reflection
- Updated examples import paths to `github.com/Thomas0x1f/gorm-cache`

### Removed
- Reflection-based `isEmptyResult` helper function (replaced with simpler approach)

### Performance
- Improved performance by eliminating reflection overhead in empty result checks

## [v0.0.1] - 2026-01-09

### Added
- Error-based query skipping mechanism using custom `ErrCacheHit` type
- Proper `RowsAffected` calculation for cached query results
- `calculateRowsAffected` helper function with reflection support for slices and structs
- Early error detection in `queryCallback` to prevent overwriting existing errors
- SQL building before cache key generation to ensure proper cache key creation

### Changed
- Refactored `queryCallback` to use error-based control flow instead of `SkipHooks`
- Improved `afterQueryCallback` to detect and clear internal cache hit errors
- Optimized cache hit detection logic by removing redundant Settings checks

### Fixed
- Cache hits now properly skip database queries by leveraging GORM's error checking
- Existing errors are no longer overwritten by cache operations
- Query metadata (RowsAffected) is correctly preserved for cached results

[v0.0.2]: https://github.com/Thomas0x1f/gorm-cache/releases/tag/v0.0.2
[v0.0.1]: https://github.com/Thomas0x1f/gorm-cache/releases/tag/v0.0.1
