# Repository Guidelines

## Project Structure & Module Organization
The CLI entrypoint is `main.go`, and subcommands sit in `cmd/` (for example `cmd/publish.go`). Shared logic belongs in `pkg/` packages such as `parsers`, `fileloader`, and `logger`; add new utilities here to keep commands thin. Reference docs live in `docs/`, helper scripts in `scripts/`, and runtime fixtures in `priv/`—update those assets whenever payload formats or examples shift.

## Build, Test, and Development Commands
Use Go 1.24+ locally. Key tasks:
- `make run arg="publish --help"` runs the CLI with custom arguments.
- `make test` invokes `gotestsum` inside the `cli` container and writes `junit-report.xml`.
- `make lint` pipes `revive` and `staticcheck` output into JSON files for CI surfacing.
- `make regen.golden` refreshes parser golden data after intentional output changes.
Run `make test.setup` once to build the Docker image before containerised tasks.

## Coding Style & Naming Conventions
Rely on `gofmt` defaults (tabs, short package names) and keep exported names purposeful. Honour the revive rules defined in `lint.toml`; any suppression needs reviewer sign-off. Commands should expose `cobra.Command` variables named after the verb (`publishCmd`, `compileCmd`) to match existing patterns. JSON payloads use lowerCamelCase tags—adjust fixtures in `priv/` whenever structures change.

## Testing Guidelines
Place `_test.go` files beside implementations and favour table-driven tests, mirroring suites under `pkg/parsers`. Use `testify/require` for clear failures. `go test ./...` offers fast local feedback, while `make test.cover` produces `coverage.lcov` plus a Markdown summary via `scripts/lcov-to-md.sh`. For integration flows, follow the README examples that combine `test-results publish` with `--ignore-missing` when pipelines generate optional reports.

## Commit & Pull Request Guidelines
Commits should be active-voice summaries with optional scopes and PR numbers, for example `feat(parser): add junit v2 adapter (#513)`. Keep related work within a single commit to simplify reviews. Pull requests need a short description, linked tickets, confirmation of the commands you ran, and artefact samples (screenshots, generated JSON) when behaviour changes. Call out breaking CLI changes explicitly so release notes stay accurate.

## Security & Dependency Checks
Security scanning depends on the shared Semaphore toolbox. Run `make check.static` for static analysis and `make check.deps` for dependency audits; both populate `/tmp/monorepo`, so clear it if the cache becomes stale. If you must accept a finding, document the rationale and planned follow-up directly in the pull request.
