# Development Setup

This guide will help you set up your development environment for contributing to the Apache Synapse Go project.

## Prerequisites

- **Go 1.20+** - The project requires Go version 1.20 or later
- **Make** - Used for building and managing the project
- **Git** - For version control

## Setting Up Your Environment

### 1. Install Go

Follow the [official Go installation instructions](https://golang.org/doc/install) for your platform.

Verify your installation:
```
go version
```

### 2. Fork and Clone the Repository

1. Fork the repository on GitHub by clicking the "Fork" button
2. Clone your fork locally:
```
git clone https://github.com/YOUR_USERNAME/synapse-go.git
cd synapse-go
```

3. Add the upstream repository as a remote:
```
git remote add upstream https://github.com/apache/synapse-go.git
```

### 3. Install Development Tools

Install recommended Go development tools:

```
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 4. Build the Project

Build the project using Make:

```
make build
```

This will:
- Fetch dependencies
- Compile the project
- Place the binary in the `bin/` directory

### 5. Run Tests

Run the test suite:

```
make test
```

## Development Workflow

### Creating a Branch

Create a new branch for your work:

```
git checkout -b feature/your-feature-name
```

### Making Changes

1. Make your code changes
2. Format your code:
```
go fmt ./...
```

3. Run linting:
```
golangci-lint run
```

4. Run tests to ensure your changes don't break anything:
```
make test
```

### Submitting Changes

1. Commit your changes:
```
git commit -m "Add feature XYZ"
```

2. Push your branch to your fork:
```
git push origin feature/your-feature-name
```

3. Create a pull request on GitHub

## Running the Development Server

To run the Synapse server locally:

```
make build
bin/synapse
```

## Documentation Development

To work on documentation:

1. Install MkDocs and the Material theme:
```
pip install mkdocs mkdocs-material
```

2. Run the documentation server locally:
```
mkdocs serve
```

3. View the documentation at http://localhost:8000

## Troubleshooting

### Common Build Issues

**Problem**: Dependencies failing to download
**Solution**: Run `go mod tidy` to clean up dependencies

**Problem**: Test failures
**Solution**: Check that your Go version meets requirements and that you haven't broken existing functionality

### Getting Help

If you encounter issues setting up your development environment, please:
1. Check existing GitHub issues
2. Ask for help in the project's discussion forum
3. Open a new issue with details about your problem