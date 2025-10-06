# Contribution Guidelines

This document outlines the guidelines for contributing to the Apache Synapse Go project.

## Code of Conduct

Contributors are expected to adhere to the Apache Software Foundation's Code of Conduct. Please be respectful and inclusive in all interactions.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Set up the development environment (see [Development Setup](setup.md))
4. Create a branch for your changes

## Pull Request Process

1. Ensure your code follows the project's coding standards
2. Include appropriate tests for your changes
3. Update documentation if necessary
4. Submit a pull request with a clear description of the changes

## Coding Standards

- Follow Go best practices and idioms
- Use meaningful variable and function names
- Write clear comments for public APIs and complex logic
- Format code with `go fmt`
- Ensure code passes `go vet` and `go lint`

## Testing Requirements

- Add unit tests for new functionality
- Ensure all tests pass before submitting a pull request
- Test coverage should not decrease

## Documentation

- Update documentation for API changes
- Include code examples where appropriate
- Document design decisions for significant changes

## Commit Guidelines

- Use present tense, imperative style for commit messages
- Start with a capital letter
- No period at the end
- Keep the first line under 50 characters
- Provide a more detailed description if necessary

Example:
```
Add file inbound endpoint implementation

This commit adds the implementation of the file inbound endpoint,
including file watching, event handling, and message processing.
```

## License Headers

All source files must include the Apache License header. Use the existing files as a template.

## Questions?

If you have any questions about contributing, please open an issue on ASF Jira or reach out to the project maintainers.