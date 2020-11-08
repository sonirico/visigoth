package container

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

type mapFactory func() Map

type mapEntry struct {
	key   string
	value interface{}
}

func genUniqueEntries(count int) []*mapEntry {
	c := make([]*mapEntry, count)
	for count := count; count > 0; count-- {
		c[count-1] = &mapEntry{key: fmt.Sprintf("key-%d", count), value: count}
	}
	return c
}

type simpleMap struct {
	L    sync.RWMutex
	data map[string]interface{}
}

func newSimpleMap() *simpleMap {
	return &simpleMap{data: make(map[string]interface{})}
}

func (t *simpleMap) Set(key string, value interface{}) {
	t.L.Lock()
	t.data[key] = value
	t.L.Unlock()
}

func (t *simpleMap) Get(key string) (data interface{}, ok bool) {
	t.L.RLock()
	data, ok = t.data[key]
	t.L.RUnlock()
	return
}

func (t *simpleMap) Del(key string) {
	t.L.RLock()
	delete(t.data, key)
	t.L.RUnlock()
	return
}

func TestShardedMap(t *testing.T) {
	m := NewShardedMap(32)
	m.Set("name", "Pepe")
	data, ok := m.Get("name")
	if !ok || data != "Pepe" {
		log.Fatalf("want Pepe, have %s", data)
	}
	m.Set("name", "Paco")
	data, ok = m.Get("name")
	if !ok || data != "Paco" {
		log.Fatalf("want Paco, have %s", data)
	}
}

func TestShardedMap_Concurrent(t *testing.T) {
	m := NewShardedMap(32)
	wg := sync.WaitGroup{}
	wg.Add(5)
	for _, name := range []string{"1", "2", "3", "4", "5"} {
		go func(x string) {
			m.Set("name", x)
			wg.Done()
		}(name)
	}
	wg.Wait()
	data, ok := m.Get("name")
	log.Println(data, ok)
}

type syncMapAdapter struct {
	*sync.Map
}

func newSyncMapAdapter() *syncMapAdapter {
	return &syncMapAdapter{Map: &sync.Map{}}
}

func (s *syncMapAdapter) Get(k string) (interface{}, bool) {
	return s.Load(k)
}

func (s *syncMapAdapter) Set(k string, v interface{}) {
	s.Store(k, v)
}

func (s *syncMapAdapter) Del(k string) {
	s.Delete(k)
}

func benchmarkMap(b *testing.B, mf mapFactory) {
	entryCardinality := 10000
	m := mf()
	entries := genUniqueEntries(entryCardinality)
	for _, entry := range entries {
		m.Set(entry.key, entry.value)
	}
	b.ResetTimer()
	for j := 0; j < 5; j++ {
		b.Run(strconv.Itoa(j), func(b *testing.B) {
			b.N = 10000
			wg := sync.WaitGroup{}
			wg.Add(2 * b.N)
			for i := 0; i < b.N; i++ {
				entry := entries[rand.Intn(entryCardinality)]

				go func(e *mapEntry) {
					m.Get(e.key)
					wg.Done()
				}(entry)

				go func(e *mapEntry) {
					m.Set(e.key, e.value)
					wg.Done()
				}(entry)
			}
			wg.Wait()
		})
	}
}

func BenchmarkNewShardedMap_32_Shards(b *testing.B) {
	benchmarkMap(b, func() Map {
		return NewShardedMap(1 << 5)
	})
}

func BenchmarkNewShardedMap_256_Shards(b *testing.B) {
	benchmarkMap(b, func() Map {
		return NewShardedMap(1 << 8)
	})
}

func BenchmarkNewShardedMap_1024_Shards(b *testing.B) {
	benchmarkMap(b, func() Map {
		return NewShardedMap(1 << 10)
	})
}

func BenchmarkNewSimpleMap(b *testing.B) {
	benchmarkMap(b, func() Map {
		return newSimpleMap()
	})
}
func BenchmarkNewSyncMap(b *testing.B) {
	benchmarkMap(b, func() Map {
		return newSyncMapAdapter()
	})
}
