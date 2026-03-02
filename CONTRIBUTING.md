# Contributing to agent-pulse

First off — thank you for taking the time to contribute! Whether it's a bug fix, a new provider, a typo, or an idea you want to discuss, every contribution matters and is genuinely appreciated.

This project is built in the open and we want to keep it that way. These guidelines exist to make the process smooth for everyone.

## Getting started

```bash
git clone https://github.com/SantiagoBobrik/agent-pulse.git
cd agent-pulse
go build -o agent-pulse .
```

Requires Go 1.25+.

### Running tests

```bash
go test ./...
```

Always run tests before submitting a PR. If you're adding new behavior, add tests for it.

## Branching model

Create your branch from `main` using the following naming convention:

| Type | Branch name | Example |
|------|-------------|---------|
| New feature | `feature/short-description` | `feature/codex-provider` |
| Bug fix | `fix/short-description` | `fix/dispatcher-timeout` |
| Documentation | `docs/short-description` | `docs/config-reference` |
| Refactor | `refactor/short-description` | `refactor/event-handler` |

## Commit messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add codex cli provider
fix: resolve race condition in dispatcher
docs: update config reference
refactor: extract event validation logic
chore: update dependencies
test: add dispatcher timeout tests
```

Keep them short, lowercase, imperative. If a commit closes an issue, reference it in the body:

```
fix: resolve client timeout on slow networks

Closes #42
```

## Code style

Keep it clean. Keep it absurdly simple.

Go already has strong conventions — follow them. Run `go fmt`, run `go vet`, write code that reads like it was obvious from the start. If something feels over-engineered, it probably is. Prefer the straightforward approach over the clever one.

## Reporting bugs

Open an [issue](https://github.com/SantiagoBobrik/agent-pulse/issues/new?template=bug_report.md) with:

- What you expected to happen
- What actually happened
- Steps to reproduce
- Go version and OS
- Relevant logs (`agent-pulse server logs`)

## Suggesting features

Have an idea? Open an [issue](https://github.com/SantiagoBobrik/agent-pulse/issues/new?template=feature_request.md) describing the use case and why it would be useful. We're happy to discuss before you start coding — this saves everyone time and keeps things aligned.

## Pull requests

1. Fork the repo and create your branch from `main` (see [branching model](#branching-model))
2. Make your changes
3. Add or update tests if applicable
4. Run `go test ./...` and make sure everything passes
5. Open a PR with a clear description of what and why

Keep PRs focused. One concern per PR. Small PRs get reviewed faster and merged sooner.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
