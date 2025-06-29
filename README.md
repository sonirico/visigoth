<div align="center">
  <img src="visigoth.png" alt="Visigoth" width="200"/>
  
  # Visigoth
  
  A Go package for full-text search and indexing
  
  [![Go](https://img.shields.io/badge/Go-1.23%2B-blue)](https://golang.org/)
  [![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
  [![CI](https://github.com/sonirico/visigoth/actions/workflows/test.yml/badge.svg)](https://github.com/sonirico/visigoth/actions/workflows/test.yml)
  
</div>

## Features

- Text analysis with tokenization, filtering, and stemming
- Memory-efficient inverted index
- Multiple search algorithms
- Spanish language support with Snowball stemming
- Repository pattern for data persistence

## Installation

```bash
go get github.com/sonirico/visigoth
```

## Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/sonirico/visigoth"
)

func main() {
    // Create analyzer
    analyzer := visigoth.NewStandardAnalyzer()
    
    // Create index
    idx := visigoth.NewMemoryIndex("example", analyzer)
    
    // Index documents
    documents := []string{
        "El gato subió al tejado",
        "Los perros corren en el parque",
        "La casa tiene ventanas grandes",
    }
    
    for i, doc := range documents {
        if err := idx.Index(i, doc); err != nil {
            log.Printf("Error indexing document %d: %v", i, err)
        }
    }
    
    // Search
    results := visigoth.LinearSearch(idx, "gato casa")
    
    fmt.Printf("Found %d results\n", len(results))
    for _, result := range results {
        fmt.Printf("Document %d: %.2f\n", result.DocID, result.Score)
    }
}
```

## Architecture

The package is organized into focused components:

- **analyze** - Text processing pipeline
- **index** - Document indexing and storage  
- **search** - Query processing and ranking
- **stemmer** - Language-specific text normalization
- **repos** - Data persistence layer
- **loaders** - Document loading utilities
- **entities** - Core data structures

## License

Licensed under the [MIT License](LICENSE).
