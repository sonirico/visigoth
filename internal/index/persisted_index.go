package index

import (
	"log"

	"github.com/sonirico/visigoth/internal/search"
	"github.com/sonirico/visigoth/pkg/entities"
)

type engine interface {
	Insert(key, val string) error
	Search(key string) (string, error)
}

type InMemoryPersistedIndex struct {
	memoryIndex Index
	basePath    string
	engine      engine
}

func (i InMemoryPersistedIndex) Put(doc indexable) Index {
	i.memoryIndex.Put(doc)

	go func() {
		if err := i.engine.Insert(doc.ID(), doc.Statement()); err != nil {
			log.Println("could not insert")
		}
	}()

	return i
}

func (i InMemoryPersistedIndex) Search(terms string, engine search.Engine) entities.Iterator {
	r := i.memoryIndex.Search(terms, engine)
	return r
}
