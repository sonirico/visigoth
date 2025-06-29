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
- Multiple search algorithms (Linear, Hits-based, Noop)
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
    "context"
    "fmt"
    "log"
    "github.com/sonirico/visigoth"
)

type VisigothSearcher struct {
    repo visigoth.Repo
}

type SearchPayload struct {
    Index string
    Terms string
}

func (v *VisigothSearcher) Search(p SearchPayload) error {
    stream, err := v.repo.Search(p.Index, p.Terms, visigoth.HitsSearch)
    if err != nil {
        return err
    }
    
    // Process results from stream
    for result := range stream.Chan() {
        fmt.Printf("Document: %s, Hits: %d\n", 
            result.Doc().ID(), result.Hits)
    }
    
    return nil
}

func main() {
    // Create tokenization pipeline with Spanish support
    tokenizer := visigoth.NewKeepAlphanumericTokenizer()
    pipeline := visigoth.NewTokenizationPipeline(
        tokenizer,
        visigoth.NewLowerCaseTokenizer(),
        visigoth.NewStopWordsFilter(visigoth.SpanishStopWords),
        visigoth.NewSpanishStemmer(true),
    )
    
    // Create repository with memory index builder
    repo := visigoth.NewIndexRepo(visigoth.NewMemoryIndexBuilder(pipeline))
    
    // Create searcher
    searcher := &VisigothSearcher{repo: repo}
    
    // Index some documents
    repo.Put("courses", visigoth.NewDocRequest("java-course", "Curso de programación en Java"))
    repo.Put("courses", visigoth.NewDocRequest("go-course", "Curso de programación en Go"))
    repo.Put("courses", visigoth.NewDocRequest("python-course", "Curso de programación en Python"))
    
    // Search
    if err := searcher.Search(SearchPayload{
        Index: "courses",
        Terms: "programación java",
    }); err != nil {
        log.Fatal(err)
    }
}
```

## Architecture

The package is organized into focused components:

- **analyze** - Text processing pipeline
- **index** - Document indexing and storage  
- **search** - Query processing and ranking
- **stemmer** - Language-specific text normalization
- **repos** - Data persistence layer with aliasing
- **loaders** - Document loading utilities
- **entities** - Core data structures

## License

Licensed under the [MIT License](LICENSE).
