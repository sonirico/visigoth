```
$$\    $$\ $$\           $$\                      $$\     $$\
$$ |   $$ |\__|          \__|                     $$ |    $$ |
$$ |   $$ |$$\  $$$$$$$\ $$\  $$$$$$\   $$$$$$\ $$$$$$\   $$$$$$$\
\$$\  $$  |$$ |$$  _____|$$ |$$  __$$\ $$  __$$\\_$$  _|  $$  __$$\
 \$$\$$  / $$ |\$$$$$$\  $$ |$$ /  $$ |$$ /  $$ | $$ |    $$ |  $$ |
  \$$$  /  $$ | \____$$\ $$ |$$ |  $$ |$$ |  $$ | $$ |$$\ $$ |  $$ |
   \$  /   $$ |$$$$$$$  |$$ |\$$$$$$$ |\$$$$$$  | \$$$$  |$$ |  $$ |
    \_/    \__|\_______/ \__| \____$$ | \______/   \____/ \__|  \__|
                             $$\   $$ |
                             \$$$$$$  |
                              \______/
```

## Table of Contents

- [Table of Contents](#table-of-contents)
  - [About ](#about-)
- [Getting Started ](#getting-started-)
  - [Prerequisites](#prerequisites)
  - [Installing](#installing)
  - [Quick Start](#quick-start)
- [Usage ](#usage-)
  - [Server Modes](#server-modes)
  - [Basic Operations](#basic-operations)
    - [Starting the Server](#starting-the-server)
    - [Indexing Documents](#indexing-documents)
    - [Search Queries](#search-queries)
  - [Development](#development)
    - [Available Make Targets](#available-make-targets)
    - [Project Structure](#project-structure)
  - [Configuration](#configuration)
- [Architecture](#architecture)
  - [Text Analysis Pipeline](#text-analysis-pipeline)
  - [Indexing](#indexing)
  - [Search Engine](#search-engine)
  - [Transport Layer](#transport-layer)
- [Contributing](#contributing)
  - [Development Setup](#development-setup)
- [License](#license)
- [Acknowledgments](#acknowledgments)

### About <a name = "about"></a>

**Visigoth** is a full-text search engine implemented in Go, designed for learning purposes and high-performance text indexing and searching. It features:

- **Fast indexing** of documents with customizable text analyzers
- **Boolean search queries** with support for complex expressions
- **Multiple transport protocols** (HTTP and TCP)
- **Configurable text analysis pipeline** with tokenizers, filters, and stemmers
- **Memory-efficient indexing** with AVL trees and sharded maps
- **Spanish language support** with Snowball stemming

## Getting Started <a name = "getting_started"></a>

### Prerequisites

- **Go 1.16+** (tested on Go 1.16)
- **Make** for build automation
- **Docker** (optional, for containerized deployment)
- **jq** (for JSON processing in build scripts)

Tested on macOS and Linux.

### Installing

1. Clone the repository:
   ```bash
   git clone https://github.com/sonirico/visigoth.git
   cd visigoth
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the project:
   ```bash
   make build
   ```

### Quick Start

1. **Start the server** (runs on ports 7373 TCP and 7374 HTTP):
   ```bash
   make run-server
   ```

2. **Test with the client**:
   ```bash
   make run-client
   ```

## Usage <a name = "usage"></a>

### Server Modes

The search engine supports two transport protocols:

- **HTTP Server** - Port 7374 (RESTful API)
- **TCP Server** - Port 7373 (Binary protocol)

### Basic Operations

#### Starting the Server
```bash
# Development mode
make run-server

# Production build
make build
./build/linux-amd64/server
```

#### Indexing Documents

Documents can be indexed via HTTP API or TCP protocol. The engine supports JSON document format with customizable field mapping.

#### Search Queries

The engine supports a custom query language (VQL - Visigoth Query Language) for complex boolean searches:

- Simple term search: `hello`
- Boolean operators: `hello AND world`
- Field-specific search: `title:golang`
- Phrase search: `"search engine"`

### Development

#### Available Make Targets

```bash
make help          # Show all available commands
make build         # Build for multiple platforms
make build-dev     # Build for current platform only
make test          # Run tests with race detection
make coverage      # Run tests with coverage report
make lint          # Run linter
make fmt           # Format code
make clean         # Clean build artifacts
make docker        # Build Docker image
make setup         # Install git pre-commit hooks
```

#### Project Structure

- `cmd/` - Main applications (server and client)
- `internal/` - Private application code
  - `index/` - Indexing algorithms and data structures
  - `search/` - Search implementations
  - `server/` - HTTP and TCP server implementations
- `pkg/` - Public library code
  - `analyze/` - Text analysis pipeline
  - `vql/` - Query language parser
  - `entities/` - Core data structures

### Configuration

The server can be configured through environment variables or command-line flags. See the server documentation for detailed configuration options.

## Architecture

Visigoth follows a modular architecture:

### Text Analysis Pipeline
- **Tokenization**: Splits text into tokens
- **Filtering**: Applies lowercase, stopword removal, and stemming
- **Language Support**: Spanish stemming via Snowball algorithm

### Indexing
- **Memory Index**: Fast in-memory inverted index using AVL trees
- **Sharded Maps**: Concurrent-safe data structures for high performance
- **Document Storage**: Efficient document metadata storage

### Search Engine
- **Linear Search**: Basic search implementation
- **Boolean Search**: Support for AND, OR, NOT operations
- **Query Parser**: VQL (Visigoth Query Language) for complex queries

### Transport Layer
- **HTTP API**: RESTful interface for web integration
- **TCP Protocol**: Binary protocol for high-performance clients
- **Custom Serialization**: Efficient binary serialization format

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linter (`make lint`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Setup

```bash
# Install git hooks
make setup

# Run tests
make test

# Run linter
make lint

# Format code
make fmt
```

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

## Acknowledgments

- Uses [Snowball stemming algorithm](https://github.com/kljensen/snowball) for Spanish language support
- Inspired by modern search engines like Elasticsearch and Solr
- Built with performance and learning in mind
