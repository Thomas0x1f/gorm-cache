# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.1] - 2026-02-04

### Fixed
- COUNT query results now properly retrieved from cache instead of returning 0
- `calculateRowsAffected` now correctly handles primitive types (int64, uint64, float, etc.)
- Deserialization errors now gracefully fall back to database queries

### Added
- Comprehensive test coverage for COUNT query caching with MsgPack serializer
- Test cases for COUNT queries with and without conditions

### Technical Details
The root cause was that primitive types in `calculateRowsAffected` were returning 0 for RowsAffected, causing GORM to overwrite the correctly deserialized COUNT value with 0. This fix ensures that COUNT queries (which return a single primitive value) report RowsAffected=1, allowing GORM to preserve the cached count value.

## [v0.1.0] - 2026-01-09

### Added
- `Serializer` interface for pluggable serialization backends
- `JSONSerializer` implementation (default, backward compatible)
- `MsgPackSerializer` implementation for improved performance and smaller payload size
- `Serializer` field in `Config` to allow custom serialization strategies
- Support for MessagePack serialization via `github.com/vmihailenco/msgpack/v5`

### Changed
- Cache serialization now uses configurable `Serializer` instead of hardcoded JSON
- Default configuration uses `JSONSerializer` for backward compatibility

### Performance
- MsgPack serialization provides 2-5x faster encoding/decoding compared to JSON
- MsgPack typically reduces payload size by 20-50% compared to JSON
- Reduced Redis memory usage and network transfer time when using MsgPack

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

[v0.1.1]: https://github.com/Thomas0x1f/gorm-cache/releases/tag/v0.1.1
[v0.1.0]: https://github.com/Thomas0x1f/gorm-cache/releases/tag/v0.1.0
[v0.0.2]: https://github.com/Thomas0x1f/gorm-cache/releases/tag/v0.0.2
[v0.0.1]: https://github.com/Thomas0x1f/gorm-cache/releases/tag/v0.0.1
