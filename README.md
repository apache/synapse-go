# Synapse

This is an attempt to re-write the synapse code is Golang.

---

## Prerequisites

- **Go 1.20+** (or a similar, recent version)
- **Make** (commonly available on Linux/macOS; on Windows, you can install via [Chocolatey](https://chocolatey.org/) or use a compatible environment like [Git Bash](https://gitforwindows.org/) or WSL)

---

## Getting Started

1. **Clone the repository**:
   ```
   git clone https://github.com/apache/synapse-go.git
   ```

2. **Check your Go version (optional)**:
    ```
    go version
    ```

Ensure it meets the minimum requirement.

## Comprehensive Developer Documentation

The project includes comprehensive developer documentation that covers the architecture, components, and contribution guidelines.

You can find:
- **Project Overview**: Introduction to Synapse Go, its features, and documentation organization
- **Getting Started Guide**: Complete setup instructions including prerequisites, building, packaging, and running the application
- **Architecture and implementation details**: Current state of the implementation and roadmap

### Accessing the Documentation

1. **Install MkDocs and the Material theme with extensions**:
   ```
   pip install mkdocs mkdocs-material mkdocs-material-extensions
   ```

2. **Run the documentation server locally**:
   ```
   mkdocs serve
   ```

3. **View the documentation** at http://localhost:8000 in your web browser

### Documentation Structure

The documentation is organized into several sections:

- **Architecture**: Details about hexagonal architecture, application lifecycle, and context flow
- **Core Components**: Documentation for configuration, logging, context usage, and various inbound endpoints
- **Contributing**: Guidelines for contribution and development setup

For more details, refer to the `docs/` directory in the project.

## Building & Packaging

1. **Install Dependencies**

The Makefile automatically fetches Go module dependencies (via go mod tidy) when you run make for the first time.

2. **Build**

To compile the Synapse binary for your local machine

```
make build
```

This fetches dependencies (if not already done).

Compiles the Go application and places the binary in the bin/ directory.

3. **Package**

To create a zip file (synapse.zip) containing the compiled binary and the required folder structure run:

```
make package
```

This will:

- Create a temporary synapse/ directory (in the project root) with:
bin/ containing the compiled synapse binary
artifacts/APIs
artifacts/Endpoints
- Zip everything into synapse.zip at the root of the project.
- Clean up the temporary folders and the bin/ directory.

4. **All-in-One (Default)**

Simply running **make** (or **make all**) will execute the following steps in order:

- deps — Installs and tidies Go dependencies.
- build — Builds the synapse binary in bin/.
- package — Creates the synapse.zip with the required folder structure.

```
make
```

5. **Clean**

If you want to remove all build artifacts and start fresh, run:

```
make clean
```

This deletes the bin/ folder and any synapse/ directories created during the packaging step.

**Customizing the Build**

If you need to cross-compile for multiple OS/architectures, you can add additional targets to the Makefile. For example:

```
build-linux:
    GOOS=linux GOARCH=amd64 go build -ldflags=$(LDFLAGS) -o bin/$(PROJECT_NAME) $(MAIN_PACKAGE)
```

Then run:

```
make build-linux
```

…and package as usual with:

```
make package
```

(Adjust paths and names as needed.)

## Running the server

After you unzip synapse.zip, you will see:

```
synapse/
├── bin/
│   └── synapse       # Compiled binary
└── artifacts/
    ├── APIs/
    └── Endpoints/
```

Unzip the archive:

```
unzip synapse.zip
```

Run the binary:

```
cd synapse/bin
./synapse
```

(On Windows, it would be .\synapse.exe if compiled for Windows.)

**Contributing**

- Fork the repository

- Create your feature branch (git checkout -b feature/my-feature)

- Commit your changes (git commit -am 'Add some feature')

- Push to the branch (git push origin feature/my-feature)

- Create a new Pull Request

**License**

Apache 2