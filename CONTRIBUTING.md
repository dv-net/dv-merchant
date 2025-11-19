<div align="center">

## 🤝 Contributing to DV.net Merchant Backend

*Guidelines for contributing to the project*

</div>

---

## 📋 Table of Contents

- [🚀 Getting Started](#-getting-started) — Setup development environment
- [🔄 Development Workflow](#-development-workflow) — Branch strategy and workflow
- [📐 Coding Standards](#-coding-standards) — Code style and conventions
- [🧪 Testing](#-testing) — Testing requirements and guidelines
- [💬 Commit Messages](#-commit-messages) — Commit message format
- [🔀 Pull Request Process](#-pull-request-process) — PR submission and review
- [🐛 Issue Reporting](#-issue-reporting) — How to report bugs
- [🔒 Security](#-security) — Security vulnerability reporting
- [👀 Code Review](#-code-review) — Review process and criteria
- [🏷️ Release Process](#-release-process) — Versioning and releases

---

## 🚀 Getting Started

### Prerequisites

- **Go 1.24.4+** — [Download](https://go.dev/dl/)
- **PostgreSQL** — Database operations
- **Redis** — Caching (optional for local dev)
- **Make** — Build commands
- **Git** — Version control

### Setup

```bash
# 1. Fork and clone
git clone https://github.com/YOUR_USERNAME/dv-merchant.git
cd dv-merchant

# 2. Add upstream remote
git remote add upstream https://github.com/dv-net/dv-merchant.git

# 3. Build and verify
make build
make test
```

> 💡 **Tip**: Run `go mod download` if dependencies are missing

---

## 🔄 Development Workflow

### Branch Strategy

- 🌿 **`main`** — Production-ready stable code
- 🔧 **`dev`** — Active development branch
- 🌱 **`feature/*`** — New features (target: `dev`)
- 🐛 **`fix/*`** — Bug fixes (target: `dev`)

### Workflow

```bash
# 1. Update main branch
git checkout main
git pull upstream main

# 2. Create feature branch
git checkout -b feature/your-feature-name

# 3. Make changes, then verify
make fmt
make lint
make test
```

> ⚠️ **Important**: Always create PRs from feature branches, never from `main` or `dev`

---

## 📐 Coding Standards

### Style Guide

Follow [Effective Go](https://go.dev/doc/effective_go) and project conventions:

- **Formatting** — `gofumpt` (via `make fmt`)
- **Imports** — `goimports` for organization
- **Naming** — Go naming conventions
- **Errors** — Explicit handling required
- **Documentation** — Document all exported functions/types

### Linting

```bash
# Build custom plugins (first time only)
make build_plugins

# Run linter
make lint
```

### Architecture

```
cmd/                CLI entrypoints
internal/delivery  HTTP handlers, middleware
internal/service   Business logic
internal/storage   Repositories
pkg/               Shared libraries
sql/               Migrations, codegen
```

### Key Rules

- 🚫 **Transactions** — Never use `pgx.BeginTransaction` directly
- ✅ **Structs** — Initialize all struct fields in constructors
- ✅ **Naming** — Use `snake_case` for JSON/YAML fields
- ✅ **Size** — Functions < 180 lines (handlers configurable)
- ✅ **Complexity** — Cyclomatic complexity < 60

---

## 🧪 Testing

### Requirements

- ✅ **New Features** — Must include tests
- ✅ **Bug Fixes** — Must include regression tests
- ✅ **Framework** — Use `testify` for assertions
- ✅ **Naming** — Test files: `*_test.go` in same package

### Running Tests

```bash
# Run all tests
make test

# Run specific package
go test ./internal/service/package

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...
```

### Coverage

> 🎯 **Target**: **80%+** coverage for new code
> 
> Focus on testing business logic and edge cases

---

## 💬 Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Commit Types

- `feat` — New feature
- `fix` — Bug fix
- `docs` — Documentation changes
- `refactor` — Code refactoring
- `perf` — Performance improvements
- `test` — Adding or updating tests
- `chore` — Maintenance tasks
- `security` — Security fixes

### Example

```bash
feat(exchange): add Binance withdrawal support

Add support for Binance exchange withdrawals with proper
error handling and retry logic.

Closes #123
```

---

## 🔀 Pull Request Process

### Before Submitting

```bash
# 1. Update your branch
git checkout main
git pull upstream main
git checkout your-branch
git rebase upstream/main

# 2. Run all checks
make fmt
make lint
make test
```

### Creating PR

**Step 1**: Push your branch
```bash
git push origin your-branch
```

**Step 2**: Create PR on GitHub
- Target: `main` or `dev` branch
- Title: Clear and descriptive
- Description: Include what changed and why
- Issues: Link related issue numbers

**Step 3**: Verify requirements

- ✅ **Code style** — Follows project guidelines
- ✅ **Tests** — `make test` passes
- ✅ **Linting** — `make lint` passes
- ✅ **Documentation** — Updated if needed
- ✅ **Conflicts** — No merge conflicts
- ✅ **Commits** — Follow conventions

### Review Process

- **Initial Review** — Within 48 hours
- **Follow-up** — Within 24 hours
- **CI Checks** — Must all pass
- **Branch Status** — Keep updated with target

> 💡 **Tip**: Address review comments promptly and keep your branch rebased

---

## 🐛 Issue Reporting

### Before Reporting

- 🔍 **Duplicates** — Check existing issues
- 🌿 **Branch** — Verify in latest `main` or `dev`
- 📦 **Version** — Ensure using latest version

### Issue Template

When creating an issue, include:

- **OS and Version** — Your environment details
- **Steps to Reproduce** — Clear, numbered steps
- **Expected Behavior** — What should happen
- **Actual Behavior** — What actually happens
- **Logs** — Relevant error logs
- **Screenshots** — If applicable

> 📝 **Note**: The more details you provide, the faster we can help

---

## 🔒 Security

### Security Issues

> ⚠️ **IMPORTANT**: **DO NOT** create public issues for security vulnerabilities.

- 📧 **Email** — [support@dv.net](mailto:support@dv.net)
- 📋 **Details** — Include detailed vulnerability information
- ⏱️ **Disclosure** — Allow time for fix before public disclosure

> 🔐 Security issues are handled privately to protect users

---

## 👀 Code Review

### Review Criteria

- ✅ **Code Quality** — Style and best practices
- ✅ **Test Coverage** — Adequate test coverage
- ✅ **Documentation** — Updated documentation
- ✅ **Security** — Security considerations
- ⚡ **Performance** — Performance impact
- 🔄 **Compatibility** — Backward compatibility

### Timeline

- **Initial Review** — **48 hours**
- **Follow-up Reviews** — **24 hours**
- **Merge Decision** — **1 week** (for approved PRs)

---

## 🏷️ Release Process

### Release Tags

- **Stable** — `vX.X.X` (production releases)
- **RC** — `vX.X.X-RC1` (release candidates)

### Process

```
1. Development in `dev` branch
2. Testing and stabilization
3. Tag release candidate: vX.X.X-RC1
4. Merge to `main`
5. Tag stable release: vX.X.X
```

---

## 🛠️ Common Tasks

### Database Migrations

```bash
# Create new migration
make db-create-migration migration_name

# Apply migrations
make migrate up

# Rollback migrations
make migrate down
```

### Code Generation

```bash
# Generate SQL code
make gensql

# Generate Swagger documentation
make swag-gen

# Generate mocks
make genmocks
```

> ⚠️ **Warning**: Never edit generated files directly. Always update source files.

### Running Server

```bash
# Build and run
make run start

# Or run directly
go run ./cmd/app start
```

---

## 📚 Resources

- 📖 **Documentation** — [docs.dv.net](https://docs.dv.net)
- 🔌 **API Reference** — [API Docs](https://docs.dv.net/en/operations/post-v1-external-wallet.html)
- 🧾 **Swagger** — [swagger.yaml](docs/swagger.yaml)
- 💬 **Support** — [dv.net/support](https://dv.net/#support)
- 📱 **Telegram** — [@dv_net_support_bot](https://t.me/dv_net_support_bot)

---

<div align="center">

**Thank you for contributing to DV.net Merchant Backend!** 🙏

*Your contributions make this project better for everyone.*

</div>

