# Contributing to GORM Cache Plugin

Thank you for your interest in contributing to GORM Cache Plugin! This document explains how you can contribute to the project.

## Development Environment

### Requirements

- Go 1.21 or higher
- Git
- Redis (optional, for Redis adapter tests)

### Setup

1. Fork the repository
2. Clone your fork:
```bash
git clone https://github.com/YOUR_USERNAME/gorm-cache.git
cd gorm-cache
```

3. Install dependencies:
```bash
go mod download
```

## Development Process

### Adding New Features

1. Create a new branch:
```bash
git checkout -b feature/amazing-feature
```

2. Make your changes
3. Add tests
4. Run tests:
```bash
go test -v ./...
```

5. Commit:
```bash
git commit -m "feat: add amazing feature"
```

6. Push:
```bash
git push origin feature/amazing-feature
```

7. Open a Pull Request

### Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/) format:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation change
- `test:` - Add or fix tests
- `refactor:` - Code refactoring
- `perf:` - Performance improvement
- `chore:` - Other changes

Examples:
```
feat: add support for custom cache key prefix
fix: resolve race condition in memory adapter
docs: update README with new examples
test: add tests for redis adapter
```

## Testing

### Unit Tests

```bash
go test -v ./...
```

### Coverage

```bash
go test -cover ./...
```

### Race Detector

```bash
go test -race ./...
```

## Code Standards

- Code formatted with `gofmt`
- No linter warnings from `golint`
- Meaningful variable and function names
- Descriptive comments
- Test coverage above 80%

### Code Format

```bash
# Format code
go fmt ./...

# Vet check
go vet ./...
```

## Adding New Adapter

To add a new cache adapter:

1. Implement the `Adapter` interface
2. Write tests for the adapter
3. Add example to README
4. Add usage example to `examples/` directory

Example structure:

```go
type MyAdapter struct {
    // fields
}

func NewMyAdapter(config MyAdapterConfig) *MyAdapter {
    // implementation
}

func (a *MyAdapter) Get(ctx context.Context, key string) ([]byte, error) {
    // implementation
}

// ... other methods
```

## Documentation

- Add GoDoc comments for public functions and types
- Keep README up to date
- Add examples for new features
- Note your changes in CHANGELOG (if exists)

## Pull Request Process

1. Ensure your fork is up to date
2. Open a PR describing your changes
3. Make sure all tests pass
4. Respond to feedback during review
5. Make necessary changes

### PR Checklist

- [ ] Tests added and passing
- [ ] Code formatted (`go fmt`)
- [ ] Documentation updated
- [ ] Examples added (if needed)
- [ ] CHANGELOG updated (for major changes)

## Questions and Support

For questions:
- Use GitHub Issues
- Provide detailed description
- Include code examples when possible
- Share error messages

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Thank You!

Thank you for contributing to GORM Cache Plugin! 🙏
