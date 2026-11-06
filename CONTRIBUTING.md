# Contributing to qq

Contributions are welcome, especially upkeep, performance work, and new encodings. Issues labelled `good first issue` are a good place to start.

## Development setup

Requirements: Go (version from `go.mod`), `jq` (used by the shell tests), and Python 3 (release script tests only).

```sh
git clone https://github.com/JFryy/qq.git
cd qq
make build   # builds bin/qq
make test    # shell tests in tests/test.sh, then go test ./... -cover
```

## Before opening a pull request

CI runs the following, so run them locally first:

```sh
gofmt -l .        # must print nothing; fix with gofmt -w .
go vet ./...
make test
```

- Keep pull requests focused on one change.
- Add or update tests for any behaviour change.
- Update `README.md` if you add a flag, format, or user-visible behaviour.

## AI-assisted contributions

AI-assisted pull requests are permitted. You are responsible for the change: review and understand every line, test it locally, and be ready to explain and revise it during review.

Drive-by pull requests generated mostly by AI with minimal human involvement will be closed. This includes changes that were not tested, do not address an existing issue or clear problem, or whose author cannot answer questions about them without AI assistance.

## Project layout

- `cli/`: command-line interface and interactive mode.
- `codec/`: one package per format, registered in `codec/codec.go`.
- `tests/`: shell tests and one fixture per format (`tests/test.<ext>`).

## Adding a codec

1. Create `codec/<name>/` with a `Codec` type exposing `Unmarshal([]byte, any) error` and, if the format is writable, `Marshal(any) ([]byte, error)`.
2. In `codec/codec.go`:
   - add a constant to the `EncodingType` block,
   - add its name to `String()` at the same index (the order must match the `iota` order),
   - add an entry to the `Codecs` map with its file extensions,
   - add it to `IsBinaryFormat` if it is a binary format.
3. Add unit tests in `codec/<name>/<name>_test.go` covering round-trips, type handling, and malformed input.
4. Add a fixture `tests/test.<ext>`. `tests/test.sh` automatically converts every fixture to every other format.
5. List the format in `README.md`.

## Reporting bugs

Please use the bug report template and include the `qq --version` output, the exact command, a minimal input file, and the expected and actual output.
