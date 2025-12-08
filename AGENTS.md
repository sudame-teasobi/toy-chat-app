# Repository Guidelines

## Project Structure & Module Organization
- Go module root is `github.com/sudame/toy-chat-app`; Go sources live under `app/`.
- Entry points: `app/cmd/read-server` and `app/cmd/write-server` each expose a simple HTTP server.
- Container build definitions sit in `dockerfiles/` (one Dockerfile per server).
- Kubernetes/Helm assets live in `helm/` with per-service values under `helm/values/local/`; `skaffold.yaml` wires images and Helm releases for local dev.

## Build, Test, and Development Commands
- Build binaries locally: `go build ./app/cmd/read-server` and `go build ./app/cmd/write-server`.
- Quick run without containers: `go run ./app/cmd/read-server` (ports 8080) or `go run ./app/cmd/write-server`.
- Build images: `docker build -f dockerfiles/Dockerfile.read-server -t read-server .` (swap filename/tag for write-server).
- Kubernetes dev loop: `kubectl create ns toy-chat-app` once, then `skaffold dev --namespace toy-chat-app` to build, deploy, and port-forward (read-server on 8081, write-server on 8080 locally).
- Tests (add as you go): `go test ./...` for package-wide execution.

## Coding Style & Naming Conventions
- Follow standard Go style: `gofmt` (tabs, goimports-style imports, exported symbols need comments).
- Package names should stay short, lower-case, and singular; binaries live under `cmd/<service-name>`.
- Favor small handlers and explicit error handling; log actionable messages (`fmt.Println` suffices for now, but prefer structured logs if added).
- Keep HTTP handlers pure where possible; avoid globals beyond configuration constants.

## Testing Guidelines
- Place tests alongside code as `*_test.go`; prefer table-driven cases for handlers/utilities.
- Mock external dependencies (e.g., Kafka or HTTP clients) via interfaces; avoid hitting live clusters in unit tests.
- Aim for coverage on request handlers and any serialization/parsing logic; run `go test ./...` before pushing.

## Commit & Pull Request Guidelines
- Use concise, type-prefixed commit messages (current history favors `feat: ...`; use `chore:`, `fix:`, etc. as appropriate).
- PRs should include: scope/intent summary, testing performed (commands and results), and any deployment notes (affected Helm values or Skaffold settings).
- Link issues when available; add screenshots only if UI changes are introduced.
- Keep diffs small and focused; mention any follow-ups explicitly.

## Kubernetes & Deployment Notes
- Services expect the `toy-chat-app` namespace; keep values in `helm/values/local/` in sync with image tags when iterating.
- Use `helmfile.yaml.gotmpl` only if you need templated releases; otherwise, prefer Skaffold’s Helm integration for local changes.
