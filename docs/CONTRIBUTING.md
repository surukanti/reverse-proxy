# Contributing to Reverse Proxy

Thank you for your interest in contributing to the reverse proxy project! This document provides guidelines and information for contributors.

## Development Setup

### Prerequisites
- Go 1.21 or later
- Docker and Docker Compose
- Make (optional, for using Makefile)

### Getting Started

1. Clone the repository:
```bash
git clone https://github.com/surukanti/reverse-proxy.git
cd reverse-proxy
```

2. Install dependencies:
```bash
go mod download
```

3. Build the project:
```bash
make build
# or
go build -o bin/proxy ./cmd/proxy
```

4. Run tests:
```bash
make test
# or
go test ./...
```

5. Run full CI checks locally:
```bash
make fmt          # Format code
make vet          # Run go vet
make lint         # Run linter
make coverage     # Generate coverage report
```

## Development Workflow

### 1. Choose an Issue
- Check existing [issues](../../issues) for something to work on
- Create a new issue if you have a feature request or bug report
- Comment on the issue to indicate you're working on it

### 2. Create a Branch
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number-description
```

### 3. Make Changes
- Follow the existing code style and conventions
- Add tests for new functionality
- Update documentation as needed
- Ensure all tests pass

### 4. Commit Changes
```bash
git add .
git commit -m "feat: add new feature description"
```

Use conventional commit format:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Test additions/changes
- `chore:` - Maintenance tasks

### 5. Create Pull Request
- Push your branch to your fork
- Create a pull request with a clear description
- Reference any related issues
- Ensure CI checks pass

## Code Guidelines

### Go Code Style
- Follow standard Go formatting (`go fmt`)
- Use `gofmt -s` for additional simplifications
- Run `go vet` to check for common mistakes
- Use `golint` or `golangci-lint` for additional checks

### Naming Conventions
- Use descriptive names for variables, functions, and types
- Follow Go naming conventions (exported vs unexported)
- Use consistent naming patterns throughout the codebase

### Package Organization
- Keep packages focused on single responsibilities
- Avoid circular dependencies
- Place test files alongside source files
- Use internal packages appropriately

### Error Handling
- Return errors rather than panicking
- Provide context in error messages
- Use error wrapping when appropriate
- Handle errors at appropriate levels

### Testing
- Write unit tests for all public functions
- Use table-driven tests where appropriate
- Test error conditions and edge cases
- Maintain good test coverage (>80%)

### Documentation
- Add comments to exported functions/types
- Keep package comments up to date
- Update README and docs for significant changes
- Use examples to demonstrate functionality

## Adding New Features

### 1. Configuration
- Add new configuration options to `internal/config/config.go`
- Update the config structs and validation
- Add example configurations in `examples/configs/`

### 2. Core Logic
- Place new features in appropriate `internal/` packages
- Follow existing patterns and interfaces
- Ensure thread safety where needed
- Add comprehensive tests

### 3. Middleware
- Add new middleware to `internal/middleware/`
- Implement the `Handler` interface
- Add configuration support
- Include in the middleware chain

### 4. Examples
- Create example configurations for new features
- Add Docker setups if applicable
- Update documentation and README

## Pull Request Process

### Before Submitting
- [ ] Tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] Linting passes (`make lint`)
- [ ] Documentation is updated
- [ ] Commit messages follow conventional format
- [ ] CI checks pass (GitHub Actions)

### CI/CD Requirements
All pull requests must pass the automated CI pipeline:

- **Go CI Workflow**: Builds and tests code on Ubuntu with Go 1.24.2
- **Test Coverage**: Maintain or improve test coverage (>80%)
- **Code Quality**: Passes `go vet` and linting checks
- **Formatting**: Code follows Go standards (`go fmt`)

The CI pipeline runs automatically on:
- Every push to `main` branch
- Every pull request targeting `main`
- Manual triggers for releases

### Review Process
1. **Automated CI Checks**: GitHub Actions runs comprehensive checks
   - Go build and test suite
   - Code formatting and linting
   - Security scanning (CodeQL)
   - Dependency checks
2. Code review by maintainers
3. Address review feedback
4. Merge when approved and CI passes

### Review Guidelines
- Be constructive and respectful
- Focus on code quality and maintainability
- Suggest improvements rather than demands
- Acknowledge good practices

## Issue Reporting

When reporting bugs or requesting features:

### Bug Reports
- Use the bug report template
- Include reproduction steps
- Provide version information
- Include relevant logs/configs

### Feature Requests
- Clearly describe the feature
- Explain the use case
- Consider implementation complexity
- Reference similar features in other projects

## Community

- Be respectful and inclusive
- Help other contributors
- Participate in discussions
- Share knowledge and best practices

Thank you for contributing to make this project better! 🚀
