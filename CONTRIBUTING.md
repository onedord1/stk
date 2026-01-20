# Contributing to SysTask

Thank you for your interest in contributing to SysTask! This document provides guidelines for contributing.

## Getting Started

### Prerequisites
- Go 1.21 or higher
- Git
- Linux/macOS development environment

### Setup
```bash
# Fork and clone
git clone https://github.com/YOUR_USERNAME/systask.git
cd systask

# Install dependencies
go mod download

# Build
go build -o systask ./main.go

# Run
./systask
```

## Development

### Project Structure
```
systask/
├── main.go                 # Entry point
├── internal/
│   ├── app/               # Main application logic
│   ├── config/            # Configuration and themes
│   ├── crypto/            # Encryption (vault)
│   ├── modules/           # Feature modules
│   │   ├── batch/         # Batch command execution
│   │   ├── disk/          # Disk management
│   │   ├── docker/        # Docker management
│   │   ├── health/        # System health metrics
│   │   ├── installer/     # Package installer
│   │   ├── logs/          # Log viewer
│   │   ├── processes/     # Process management
│   │   ├── services/      # Service management
│   │   ├── sftp/          # SFTP file manager
│   │   └── users/         # User management
│   ├── ssh/               # SSH client
│   └── ui/                # UI components
└── docs/                  # Documentation
```

### Code Style
- Follow standard Go conventions (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions
- Keep functions focused and small

### Testing
```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## Making Changes

### Branching
- `main` - Stable release branch
- `develop` - Development branch
- Feature branches: `feature/your-feature-name`
- Bug fixes: `fix/issue-description`

### Commit Messages
Follow conventional commits:
```
feat: add new module for kubernetes management
fix: resolve connection timeout in batch execution
docs: update README with new features
style: format code with gofmt
refactor: simplify SSH client connection logic
```

### Pull Requests
1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### PR Checklist
- [ ] Code follows project style guidelines
- [ ] Tests added/updated if applicable
- [ ] Documentation updated
- [ ] Changelog entry added
- [ ] No breaking changes (or documented if necessary)

## Adding a New Module

1. Create directory: `internal/modules/yourmodule/`
2. Implement the module interface:
```go
type Module interface {
    View() *tview.Flex
    Refresh() error
    SetApp(*tview.Application)
}
```
3. Register in `app.go`
4. Add keyboard shortcut
5. Update README and help

## Adding a New Theme

1. Edit `internal/config/themes.go`
2. Add theme to `themes` map:
```go
"yourtheme": {
    Name:       "yourtheme",
    Background: tcell.NewRGBColor(r, g, b),
    Foreground: tcell.NewRGBColor(r, g, b),
    // ... other colors
},
```
3. Update README with theme name

## Reporting Issues

### Bug Reports
Include:
- SysTask version
- Go version
- OS and terminal
- Steps to reproduce
- Expected vs actual behavior
- Error messages/logs

### Feature Requests
- Describe the feature
- Explain the use case
- Provide examples if possible

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- No harassment or discrimination

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
