<div align="center">

## 🤝 Contributing to DV.net Merchant Backend

</div>

---

## 📋 Table of Contents

- [Getting Started](#-getting-started)
- [Development Workflow](#-development-workflow)
- [Coding Standards](#-coding-standards)
- [Testing](#-testing)
- [Commit Messages](#-commit-messages)
- [Pull Request Process](#-pull-request-process)
- [Issue Reporting](#-issue-reporting)
- [Security](#-security)
- [Code Review](#-code-review)
- [Release Process](#-release-process)

---

## 🚀 Getting Started

### Prerequisites

- **Go 1.24.4+** — [Download](https://go.dev/dl/)
- **PostgreSQL** — database operations
- **Redis** — caching (optional for local dev)
- **Make** — build commands
- **Git** — version control

### Setup

```bash
# Fork and clone
git clone https://github.com/YOUR_USERNAME/dv-merchant.git
cd dv-merchant

# Add upstream
git remote add upstream https://github.com/dv-net/dv-merchant.git

# Install dependencies
go mod download

# Build and test
make build
make test
```

---

## 🔄 Development Workflow

### Branch Strategy

- 🌿 **`main`** — production-ready stable code
- 🔧 **`dev`** — active development branch
- 🌱 **`feature/*`** — new features
- 🐛 **`fix/*`** — bug fixes
- 📚 **`docs/*`** — documentation updates

### Workflow

```bash
# Update main
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name

# Make changes, then verify
make fmt
make lint
make test
```

---

## 📐 Coding Standards

### Style Guide

Follow [Effective Go](https://go.dev/doc/effective_go) and project conventions:

- **Formatting**: `gofumpt` (via `make fmt`)
- **Imports**: `goimports` for organization
- **Naming**: Go conventions
- **Errors**: explicit handling required
- **Documentation**: export all exported functions/types

### Linting

```bash
# Build custom plugins (first time)
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

### Rules

- 🚫 Never use `pgx.BeginTransaction` directly
- ✅ Initialize all struct fields in constructors
- ✅ Use `snake_case` for JSON/YAML fields
- ✅ Functions < 180 lines (handlers configurable)
- ✅ Cyclomatic complexity < 60

---

## 🧪 Testing

### Requirements

- All new features must include tests
- Bug fixes must include regression tests
- Use `testify` for assertions
- Test files: `*_test.go` in same package

### Running Tests

```bash
# All tests
make test

# Specific package
go test ./internal/service/package

# With coverage
go test -cover ./...
```

### Coverage

- Target **80%+** for new code
- Focus on business logic and edge cases

---

## 💬 Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat` — new feature
- `fix` — bug fix
- `docs` — documentation
- `refactor` — code refactoring
- `perf` — performance
- `test` — tests
- `chore` — maintenance
- `security` — security fixes

### Example

```
feat(exchange): add Binance withdrawal support

Add support for Binance exchange withdrawals with proper
error handling and retry logic.

Closes #123
```

---

## 🔀 Pull Request Process

### Before Submitting

```bash
# Update branch
git checkout main
git pull upstream main
git checkout your-branch
git rebase upstream/main

# Run checks
make fmt
make lint
make test
```

### Creating PR

1. Push branch: `git push origin your-branch`
2. Create PR on GitHub targeting `main` or `dev`
3. Include:
   - Clear title and description
   - Related issue numbers
   - What changed and why

### Checklist

- [ ] Code follows style guidelines
- [ ] Tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Documentation updated
- [ ] No merge conflicts
- [ ] Commit messages follow conventions

### Review Process

- Initial review: within 48 hours
- Follow-up: within 24 hours
- All CI checks must pass
- Keep branch updated with target branch

---

## 🐛 Issue Reporting

### Before Reporting

- Check existing issues for duplicates
- Verify issue exists in latest `main` or `dev` branch
- Ensure you're using the latest version

### Issue Template

Include:

- **OS and Version**: Your environment
- **Steps to Reproduce**: Clear, numbered steps
- **Expected Behavior**: What should happen
- **Actual Behavior**: What actually happens
- **Logs**: Relevant error logs
- **Screenshots**: If applicable

---

## 🔒 Security

### Security Issues

**DO NOT** create public issues for security vulnerabilities.

- Email: [support@dv.net](mailto:support@dv.net)
- Include detailed vulnerability information
- Allow time for fix before public disclosure

---

## 👀 Code Review

### Review Criteria

- Code quality and style
- Test coverage
- Documentation updates
- Security considerations
- Performance impact
- Backward compatibility

### Timeline

- Initial review: **48 hours**
- Follow-up reviews: **24 hours**
- Merge decision: **1 week** (for approved PRs)

---

## 🏷️ Release Process

### Versioning

Follows semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Tags

- **Stable**: `vX.X.X` — production releases
- **RC**: `vX.X.X-RC1` — release candidates

### Process

1. Development in `dev` branch
2. Testing and stabilization
3. Tag release candidate: `vX.X.X-RC1`
4. Tag stable release: `vX.X.X`
5. Merge to `main`

---

## 🛠️ Common Tasks

### Database Migrations

```bash
# Create migration
make db-create-migration migration_name

# Apply
make migrate up

# Rollback
make migrate down
```

### Code Generation

```bash
# SQL code
make gensql

# Swagger docs
make swag-gen

# Mocks
make genmocks
```

**⚠️ Never edit generated files directly.** Update source files.

### Running Server

```bash
# Build and run
make run start

# Direct run
go run ./cmd/app start
```

---

## 📚 Resources

- 📖 [Documentation](https://docs.dv.net)
- 🔌 [API Reference](https://docs.dv.net/en/operations/post-v1-external-wallet.html)
- 🧾 [Swagger](docs/swagger.yaml)
- 💬 [Support](https://dv.net/#support) • [Telegram](https://t.me/dv_net_support_bot)

---

<div align="center">

**Thank you for contributing!** 🙏

</div>

