# Changelog

All notable changes to this project will be documented in this file.

---

## [0.3.5] - 2026-07-10

### Added
- **Key Order Preservation (`--preserve-key-order` / `-k`)**:
  - Custom JSON token parsing (`dec.Token()`) tracks and registers map key sequences in a global thread-safe registry (`sync.RWMutex`).
  - Custom YAML decoding registers AST mappings and normalizes pointers to guarantee key order preservation.
  - CSV/TSV codecs maintain column sequence matching the original headers on both import and export.
  - XML codecs traverse standard token streams to preserve order of sibling tags.
  - Added an unconditional global memory registry cleanup routine (`ClearKeyOrder()`) at query startup to prevent memory leaks during interactive REPL sessions and module integrations.
- **Auto-Coercion Disabling (`--no-auto-convert`)**:
  - Disables automatic string conversion during parsing (e.g. converting `"T"`, `"F"`, `"1"`, or `"0"` to booleans/numbers).
  - Supported across CSV, TSV, XML, INI, Gron, and Line formats.
  - Bypasses string trimming and parse routines, offering a significant parsing performance boost.
- **Interactive Mode Format Decoupling (`--interactive` / `-I`)**:
  - Always initializes the interactive TUI state buffer using JSON to prevent parsing errors when query prototyping with non-JSON target output formats.
  - Correctly serializes and formats the final query result to the target output format (e.g. YAML, XML, TOML) on graceful TUI exit.
- **Cross-Platform Build Detection**:
  - Added OS detection in `Makefile` to automatically build with the `.exe` extension on Windows hosts, ensuring native shell execution.
- **Architecture Documentation**:
  - Authored a comprehensive `ARCHITECTURE.md` file detailing application objectives, design choices, data flow pipelines, compatibility matrices, security audits, and testing structures.


### Testing & Code Quality
- **Integration Test Suite**:
  - Created `cli/integration_test.go` covering full integration pipelines.
  - Added E2E tests for JQ expression support (`.[]`, `length`, `first`, `.[0]`) on CSV inputs.
  - Added E2E tests for key preservation across all formats (JSON, JSONL, CSV, TSV, YAML, XML).
  - Added coercion tests for CSV, XML, INI, and Gron formats under `--no-auto-convert`.
  - Added edge case tests for Unicode characters and header-only empty files.
- **Clean Linter & Security Scan**:
  - Resolved all 24 warnings in the codebase.
  - Verified `golangci-lint run ./... --no-config` passes with **0 issues**.
  - Verified `gosec ./...` security scanner returns **0 issues**.
