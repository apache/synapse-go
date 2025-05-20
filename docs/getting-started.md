# Getting Started with Synapse Go

This guide will help you get up and running with the Synapse Go project on your local machine. We'll cover the prerequisites, installation, build process, and basic usage.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.23.8+** (The project uses Go 1.23.8 with toolchain 1.24.1)
- **Make** (for building and packaging)
- **Git** (for cloning the repository)
- **Unzip** (for extracting the packaged application)

You can check your Go version with:

```bash
go version
```

If you need to install or update Go, visit the [official Go download page](https://golang.org/dl/).

## Cloning the Repository

1. Clone the repository using Git:

```bash
git clone https://github.com/apache/synapse-go.git
```

2. Navigate to the project directory:

```bash
cd synapse-go
```

## Building the Project

The project includes a comprehensive Makefile with several useful targets:

### Basic Build Commands

- **Build the project** (creates the binary in the `bin/` directory):

```bash
make build
```

- **Build with debug information** (useful for development and troubleshooting):

```bash
make build-debug
```

- **Run tests**:

```bash
make test
```

### Packaging the Application

Create a distributable package containing the compiled binary and the required directory structure:

```bash
make package
```

This command:
1. Builds the application
2. Runs tests to ensure everything works
3. Creates a temporary directory structure
4. Packages everything into `synapse.zip`

### All-in-One Command

You can run all steps (dependency resolution, build, and package) with a single command:

```bash
make
```

or

```bash
make all
```

### Cleaning Build Artifacts

To clean up all build artifacts:

```bash
make clean
```

## Running Synapse Go

After building and packaging, follow these steps to run the application:

1. **Extract the package**:

```bash
unzip synapse.zip
```

This creates a `synapse` directory with the following structure:

```
synapse/
├── bin/
│   └── synapse       # Compiled binary
├── conf/             # Configuration directory
└── artifacts/
    ├── APIs/         # API definitions
    ├── Endpoints/    # Endpoint definitions
    ├── Sequences/    # Sequence definitions
    └── Inbounds/     # Inbound definitions
```

2. **Configure Synapse**:

Place your Synapse configuration files in their respective directories:
   - Configuration files go into `synapse/conf/`
   - API definitions go into `synapse/artifacts/APIs/`
   - Endpoint definitions go into `synapse/artifacts/Endpoints/`
   - Sequence definitions go into `synapse/artifacts/Sequences/`
   - Inbound definitions go into `synapse/artifacts/Inbounds/`

Ensure you have a `LoggerConfig.toml` file in the `synapse/conf/` directory to configure logging.

3. **Run the application**:

```bash
cd synapse/bin
./synapse
```

## Configuration Files

### Logger Configuration

The logger configuration file (`LoggerConfig.toml`) should be placed in the `synapse/conf/` directory. Here's a basic example:

```toml
# Sample LoggerConfig.toml
[logger]
level.default = "warn"

[logger.level.packages]
mediation = "error"
deployers = "error"
router = "info"

[logger.handler]
format = "json"
outputPath = "stdout"
```

### Deployment Configuration

The main configuration file (`deployment.toml`) should be placed in the `synapse/conf/` directory. This file defines the core behavior of your Synapse HTTP server configurations.

Here's a basic example:

```toml
[server]
hostname = "localhost"
offset  = 10
```

## Troubleshooting

If you encounter any issues:

1. Ensure your Go version meets the requirements (1.23.8+)
2. Make sure all configuration files are correctly placed in the `synapse/conf/` directory
3. Check that the binary has execute permissions:
   ```bash
   chmod +x synapse/bin/synapse
   ```
4. Check for error messages in the console output

## Next Steps

- Review the [Architecture Documentation](architecture/hexagonal.md) to understand how Synapse works
- Learn about [Core Components](components/configuration.md) to customize and extend functionality
- See the [Contributing Guidelines](contributing/guidelines.md) if you'd like to contribute to the project

## Getting Help

If you need assistance:
- Check the [issue tracker](https://github.com/apache/synapse-go/issues) for known problems
- Join the community discussions on the mailing list
- Consult the rest of this documentation for detailed information on specific components