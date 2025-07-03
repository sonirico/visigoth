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

## Search Algorithms

Visigoth provides two main search algorithms, both implementing AND logic (all query tokens must be present in matching documents):

### HitsSearch vs LinearSearch

| Feature                    | HitsSearch                                   | LinearSearch                  |
| -------------------------- | -------------------------------------------- | ----------------------------- |
| **Algorithm**              | Hit counting + threshold filtering           | Set intersection              |
| **Time Complexity**        | O(T × D + R log R)                           | O(T × D + I)                  |
| **Space Complexity**       | O(R)                                         | O(I)                          |
| **Result Ordering**        | Relevance-based (hit count) + document order | Document index order          |
| **Best For**               | Relevance ranking, scoring                   | Boolean matching, performance |
| **Multi-token Efficiency** | Constant per token                           | Early termination possible    |

Where:
- T = number of search tokens
- D = average documents per token  
- R = number of matching documents
- I = size of intersection sets

### Algorithm Details

#### HitsSearch
```go
// Uses hit counting to rank results by relevance
results := HitsSearch([]string{"programming", "tutorial"}, indexer)
// Returns documents sorted by relevance (most matching tokens first)
```

**Process:**
1. Count hits (unique tokens) per document
2. Filter documents with hits ≥ threshold (AND logic)
3. Sort by hit count (relevance), then by document order (determinism)

**Best for:**
- ✅ Relevance ranking needed
- ✅ User expects "best matches first"
- ✅ Flexible scoring systems
- ✅ Apps with ranked suggestions

#### LinearSearch
```go
// Uses set intersection for exact boolean matching
results := LinearSearch([]string{"programming", "tutorial"}, indexer)
// Returns documents in document index order
```

**Process:**
1. Get document sets for each token
2. Compute intersection of all sets (AND logic)  
3. Return matches in document index order

**Best for:**
- ✅ Pure boolean matching
- ✅ Performance-critical applications
- ✅ Large queries (many tokens)
- ✅ Consistent ordering needed

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](LICENSE)
