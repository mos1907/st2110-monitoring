# Contributing to ST 2110 Monitoring

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Report any unacceptable behavior

## How to Contribute

### Reporting Issues

- Use GitHub Issues to report bugs or request features
- Include: ST 2110 equipment details, error logs, expected behavior
- Search existing issues first to avoid duplicates

### Submitting Code

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes
4. Test thoroughly in a lab environment
5. Commit with clear messages: `git commit -m "Add feature X"`
6. Push and create a Pull Request

### Code Standards

- Follow Go best practices (gofmt, golint)
- Include comments for complex logic
- Update documentation for new features
- Add example configurations where relevant
- No breaking changes without major version bump

### Testing

- Test all code in lab environment first
- Include unit tests for new functionality
- Verify exporters work with real ST 2110 equipment
- Test in Docker and Kubernetes environments

### Documentation

- Update README.md for user-facing changes
- Add to docs/ for new features
- Include code examples
- Update API.md for metric changes

## Development Setup

1. Clone your fork
2. Install dependencies: `make deps`
3. Build exporters: `make build`
4. Start services: `make up`
5. Run tests: `make test`

## Review Process

- All PRs require review
- Maintainers will review within 7 days
- Address feedback promptly
- Squash commits before merging

## Questions?

- GitHub Discussions: For questions and discussions
- GitHub Issues: For bug reports and feature requests

Thank you for contributing!

