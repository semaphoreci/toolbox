# Test Results CLI – Internal Notes

## Purpose
- CLI used in Semaphore pipelines to parse raw test artifacts, normalize them into a common JSON schema, and push both detailed and summary reports to artifact storage.
- Supports multiple JUnit dialects and Go linters, providing consistent IDs and metadata enrichment for downstream dashboards.

## Top-Level Layout
- `cmd/` – Cobra commands (`publish`, `compile`, `combine`, `gen-pipeline-report`, `command-metrics`, `resource-metrics`) that expose user-facing verbs and orchestrate work.
- `pkg/cli/` – shared execution helpers: flag handling, file discovery, parsing, JSON marshaling, gzip compression, artifact uploads, and stats reporting.
- `pkg/parser/` – core data model (`TestResults`, suites, cases), deterministic ID generation, JSON helpers, and XML utilities.
- `pkg/parsers/` – concrete parser implementations (multiple JUnit variants, Go revive/staticcheck, etc.) plus fixtures/goldens.
- `pkg/logger/` – thin logrus wrapper that enforces the `* message` format and centralizes log level control.
- `pkg/fileloader/` and `pkg/compression/` – memoized file readers and gzip utilities.
- `scripts/`, `Dockerfile.dev`, `docker-compose.yml` – development helpers; `docs/` contains design notes (e.g., ID generation); `priv/` holds sample payloads used in tests.

## Command Flow
1. `main.go` boots `cmd.Execute()`, initializing Cobra/viper and persistent flags (`--parser`, `--name`, `--verbose`, etc.).
2. Each subcommand calls `cli.SetLogLevel` to honor verbosity/trace flags and then composes operations:
   - Load candidate files (`cli.LoadFiles`) and resolve parser selection (`cli.FindParserForFile` or `--parser` override).
   - Parse to `parser.Result`, decorate with metadata (`cli.DecorateResults`) and regenerate IDs if names/prefixes change.
   - Marshal to JSON, optionally gzip (`cli.WriteToFile`, `cli.WriteToFilePath`), and publish via `cli.PushArtifacts`.
   - `publish` merges reports, trims stdout/stderr, uploads job/workflow summaries, and pushes raw XML files when requested.
   - `compile` stops after producing a local JSON artifact.
   - `gen-pipeline-report` aggregates previously published job reports for the pipeline ID in `SEMAPHORE_PIPELINE_ID`.
   - `combine` merges existing JSON summaries; metrics commands (`command-metrics`, `resource-metrics`) emit telemetry payloads.

## Parser & ID Architecture
- `parser.Parser` interface defines `Parse`, `IsApplicable`, name/description metadata, and supported extensions.
- Concrete implementations under `pkg/parsers/` normalize framework-specific quirks:
  - `junit_generic` handles baseline XML; specialized files shim mocha, rspec, golang, phpunit, exunit.
  - Linters (revive, staticcheck) parse JSON streams into suites/cases.
- ID strategy (documented in `docs/id-generation.md`):
  - UUIDv3 seeded by existing IDs, names, framework labels, and parent hierarchy to stay deterministic across runs.
  - Suites namespace child IDs by parent result; failed tests include state to differentiate retries.
- Helpers trim output length, detect flaky states, and inject Semaphore metadata (job, workflow, repo identifiers) from environment variables.

## Artifact & Storage Integration
- Upload layer targets Semaphore artifact storage via `cli.PushArtifacts`, grouping by `job` or `workflow`.
- CLI writes a merged `test-results/junit.json` plus per-input `junit-<index>.xml` (unless `--no-raw`).
- Summaries (`summary.json`) include aggregated counts and upload statistics; stats accumulate via `cli.ArtifactStats`.
- Temporary files land under the OS temp dir; cleanup uses `defer os.Remove` and `os.RemoveAll`.

## Configuration & Environment
- Viper auto-loads `$HOME/.test-results.yaml` when present; command flags override config.
- Key env vars: `SEMAPHORE_PIPELINE_ID`, `SEMAPHORE_WORKFLOW_ID`, `SEMAPHORE_JOB_ID`, `SEMAPHORE_AGENT_MACHINE_TYPE`, etc.; parsers read them to enrich metadata.
- Flags to know: `--parser`, `--ignore-missing`, `--no-compress`, `--suite-prefix`, `--omit-output-for-passed`, `--trim-output-to`, `--no-trim-output`, `--name`.

## Development Workflow
- Requires Go 1.24+. Use `make test.setup` once to build the `cli` Docker image and fetch module deps.
- Everyday commands:
  - `make run arg="publish --help"` – smoke-run CLI without building binary.
  - `make test` – executes `gotestsum` inside Docker and exports `junit-report.xml`.
  - `make lint` – runs revive and staticcheck, producing `revive.json` and `staticcheck.json`.
  - `make regen.golden` – refresh parser golden fixtures after intentional format changes.
  - `make test.cover` – generates `coverage.lcov` and Markdown summary via `scripts/lcov-to-md.sh`.
- Local iteration without Docker: `go test ./...` and `go run main.go ...` respect the same flags.

## Testing & Fixtures
- Parser tests live beside implementations using table-driven cases and the shared helpers in `pkg/parsers/test_helpers.go`.
- Golden files live in `priv/` and `pkg/parsers/testdata` (embedded string literals); keep them in sync when tweaking parsers.
- Metrics commands have log-based fixtures under `priv/command-metrics` and `priv/resource-metrics`.
- Use `make regen.golden` plus `git diff` to validate intentional output deltas.

## Utilities & Cross-Cutting Concerns
- Logging: `pkg/logger` wraps logrus; `cli.SetLogLevel` maps `--verbose`, `--trace` to logrus levels.
- Compression: `pkg/compression` toggles gzip; CLI defaults to compressed output for artifacts.
- File caching: `pkg/fileloader.Load` memoizes `bytes.Reader` instances to avoid rereading the same path during batch operations.
- Metrics: `cmd/command-metrics` and `cmd/resource-metrics` parse Semaphore job logs to emit structured telemetry; they rely on shared parsing in `pkg/cli`.

## Tips for Future Changes
- When adding a parser, implement the `parser.Parser` interface, register it in `pkg/parsers/parsers.go`, and supply fixtures + golden tests.
- Any change to JSON schema or summary aggregation should update sample payloads in `priv/` and documented behaviour in `README.md`.
- Artifact uploads are sensitive to path conventions (`test-results/junit.json`, `test-results/<pipeline>.json`); double-check destinations when modifying.
- For CLI UX tweaks, adjust Cobra flags and keep `README.md` examples aligned; leverage `cmd/root.go` for persistent options.
