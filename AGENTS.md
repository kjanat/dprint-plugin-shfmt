# Repository Guidelines

## Project Structure & Module Organization

This repository implements a dprint Wasm plugin for `shfmt` in Go/TinyGo.

- Root package: plugin entrypoint and handlers (`main.go`, `handler_*.go`).
- `dprint/`: reusable runtime bridge, config resolver, and shared plugin types.
- `integration/`: end-to-end tests (`integration_test.go`) and fixtures in `integration/testdata/cases/<case>/`.
- `schema.json`: published plugin schema (generated).
- Generated files: `*_generated.go` and `schema.json` (regenerate; do not hand-edit).
- Build output: `plugin.wasm`; release artifacts: `dist/`.

## Build, Test, and Development Commands

Use `mise` to keep tool versions consistent.

- `mise install`: install pinned tools (Go, TinyGo, golangci-lint, dprint, goreleaser).
- `mise run generate`: run `go generate ./...` for boilerplate, config resolver, and schema outputs.
- `mise run lint-fix`: run `golangci-lint run --fix ./...` for auto-fix and formatting.
- `mise run fmt-dprint`: format Markdown/JSON/TOML/YAML files via dprint.
- `mise run lint`: run `golangci-lint` (includes `gofumpt`, `gci`, and enabled linters).
- `mise run test`: run unit tests (`go test ./...`).
- `mise run test-integration`: run integration tests with `-tags=integration`.
- `mise run build-wasm`: produce `plugin.wasm` with TinyGo.
- `mise run release-check`: validate `.goreleaser.yaml`.
- `mise run release-snapshot`: build local release artifacts without publishing.

## Coding Style & Naming Conventions

- Follow standard Go style; rely on formatters, not manual alignment.
- Keep files focused by responsibility (for example `handler_format.go`, `handler_config.go`).
- Test files must use `*_test.go`; prefer table-driven cases where practical.
- Never manually edit generated files; update generators/specs, then run `mise run generate`.

## Testing Guidelines

- Unit tests use Go’s `testing` package and should cover handler/config/runtime behavior.
- Integration cases are fixture-based: each case directory contains `config.json`, `input.sh`, and `expected.stdout`.
- No strict coverage threshold is enforced, but behavior changes should include tests at the appropriate level.

## Dependency Automation

- Dependencies are kept current by the **Renovate** GitHub App (installed, scoped to this repo). It opens PRs for `go.mod` modules, `[tools]` pins in `.config/mise.toml` (TinyGo, dprint, Go, go-licenses, golangci-lint, goreleaser), and pinned GitHub Actions. Config: `renovate.json`.
- Updates are batched: Renovate runs Mondays before 06:00 (Europe/Amsterdam) and groups its PRs into `go modules` (non-major `go.mod` updates), `github actions` (including digest re-pins), and `dev tooling` (everything under `[tools]` in `.config/mise.toml`). Major `go.mod` updates and the exceptions below still get their own PR, and security advisories ignore the schedule.
- Renovate PRs are validated by `.github/workflows/ci.yml` (test, integration, lint, build-wasm, release-check). Review and merge manually; no automerge is configured.
- High-risk bumps needing careful manual review even when CI is green: `mvdan.cc/sh/v3` (core formatter; past array-subscript regression) and `aqua:tinygo-org/tinygo` (Wasm compiler; `-X` ldflag injection quirk). Both are kept out of the groups and held for 5 days after release.
- The `go` pin (in `.config/mise.toml` and the `go.mod` directive) is capped below 1.27 by a `packageRules` entry: TinyGo 0.41.x refuses any GOROOT newer than Go 1.26, which fails `build-wasm` and `test-integration`. Lift the cap in the same change that pins TinyGo 0.42.0 or later.
- `gomodTidy` runs after every `go.mod` update, so Renovate branches keep `go.sum` tidy on their own.
- Match rules use `matchDepNames`, not `matchPackageNames`: the mise manager rewrites package names (`go` becomes `golang/go` on the `github-tags` datasource), so package-name rules silently miss. Verify a rule change with `npx --package renovate -- renovate --platform=local --dry-run=extract` before trusting it.
- Renovate labels its PRs `dependencies`, and security fixes additionally `security`. Repository labels are declared in `.github/labels.yml` and synced by `.github/workflows/labels.yml` (`kjanat/github-labelmanager`) on pushes to `master` that touch that file, or via `workflow_dispatch`; PRs run the same sync in dry-run mode. Edit the file, not the GitHub UI. Labels are created and updated from the file, but only deleted when listed under `delete:`.
- A weekly remote watchdog (`renovate-health-watchdog`, Mondays 09:00 Europe/Amsterdam) reports Renovate health and open dependency PRs into the GitHub issue titled `Dependency automation health (Renovate watchdog)`. It is report-only and never merges or edits.

## Commit & Pull Request Guidelines

- Match existing commit style: short, imperative summaries (for example, `Split runtime internals into dedicated modules`).
- Keep commits scoped to one logical change.
- Before committing, confirm documentation files (for example `README.md` and `AGENTS.md`) reflect the latest project state.
- PRs should clearly describe what changed and why.
- Link issue(s) when applicable.
- Include test evidence (commands run, such as `mise run test` and `mise run test-integration`).
- Call out schema updates or release-impacting changes.

## Release Procedure

- Official releases are created by GitHub Actions when a tag matching `v*` is pushed (see `.github/workflows/release.yml`).
- Recommended flow:
  1. Run checks locally (`mise run lint`, `mise run test`, and optionally `mise run test-integration`).
  2. Create a version tag (for example, `git tag -a v0.0.1 -m "v0.0.1"`).
  3. Push the tag (`git push origin v0.0.1`).
- Do not rely on local `mise run release` for normal releases; CI provides `GITHUB_TOKEN` and publishes the release automatically.

## Documentation Language & Sandbox Constraints

- Write project documentation in English (README, guides, and in-repo reference docs).
- Run test commands via `mise run` wrappers (for example, `mise run test` and `mise run test-integration`) instead of invoking `go test` directly, because direct execution may hit sandbox permission constraints.
- If a required command is blocked by sandbox or network restrictions, request temporary approval and rerun with escalation.
