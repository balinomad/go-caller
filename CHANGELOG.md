# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [2.1.0] - 2026-06-29

### Added

- `NewEmpty()` constructor, returning an empty, invalid `Caller` for use as a `json.Unmarshal` destination. Without it, there was no way to obtain a concrete `Caller` from outside the package: `Caller` is a non-empty interface and its only implementation is unexported, so `json.Unmarshal(data, &c)` failed at runtime with `json: cannot unmarshal object into Go value of type caller.Caller`.

### Changed

- Minimum Go version raised from 1.21 to 1.23.
- `MarshalJSON` and `UnmarshalJSON` now wrap errors from the underlying `encoding/json` calls with additional context (`fmt.Errorf("...: %w", err)`).
- `NewFromPC`'s documentation now states its required input explicitly: a call-site program counter, such as the first return value of `runtime.Caller`. A raw program counter from `runtime.Callers` is a return address, not a call-site address, and passing one directly to `NewFromPC` can resolve to the wrong line or an unrelated function entirely.

### Fixed

- Line numbers above 65535 were silently truncated to 0 through an internal `uint16` field, making valid callers in large files appear to have no line information. The field is now a plain `int`; the `uint16` saved no actual memory in the first place, since struct alignment padded it to the same overall size regardless.
- Package documentation's example import path was missing the `/v2` module suffix and would not compile as written.
- Removed several data races in the test suite that surfaced once subtests were run in parallel (`t.Parallel()`); test behavior and coverage are unchanged.
- Corrected the README throughout: an example showing `New(0)` called with no wrapper function previously implied it captures its own call site, which is incorrect; the JSON `Unmarshal` example was non-functional (see Added, above); a performance claim referencing the removed `uint16` field has been removed. Also added the Go version requirement, a concurrency contract section, and an explanation of the package's purpose relative to the standard library.

## [2.0.0] - 2025-09-15

### Added

- `Equal(other Caller) bool` method on the `Caller` interface, for semantic comparison of callers.
- Package documentation moved into the main source file.

### Changed

- **Breaking:** `Valid()` now requires only a non-empty file path. Previously it required file, line, and function to all be present.
- **Breaking:** `New()` now returns `nil` only for negative skip values, not for `skip <= -2` as before.
- Optimized string operations in `Location()`, `ShortLocation()`, `Function()`, and `Package()`.
- Massively expanded test coverage, including benchmark tests, edge case and error condition testing, JSON marshaling/unmarshaling tests, and slog integration tests.

### Fixed

- Potential panics with nil receivers across all methods.
- Negative skip value validation in `New()`.
- Edge case handling in `functionNameIndex`.
- JSON marshaling now properly omits empty fields.
- `LogValue()` now returns an empty value for invalid callers.

## [1.0.0] - 2025-04-30

Initial release.

[2.1.0]: https://github.com/balinomad/go-caller/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/balinomad/go-caller/compare/v1.0.0...v2.0.0
[1.0.0]: https://github.com/balinomad/go-caller/releases/tag/v1.0.0
