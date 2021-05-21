package container

import (
	"sync"

	"github.com/cespare/xxhash/v2"
)

type hasher interface {
	HashBytes([]byte) uint64
	HashString(string) uint64
}

type betterHasher struct{}

func (p betterHasher) HashString(key string) uint64 {
	return xxhash.Sum64String(key)
}

func (p betterHasher) HashBytes(key []byte) uint64 {
	return xxhash.Sum64(key)
}

type Map interface {
	Set(string, interface{})
	Get(string) (interface{}, bool)
	Del(string)
}

type shard struct {
	L    sync.RWMutex
	data map[string]interface{}
}

func newShard() *shard {
	return &shard{data: make(map[string]interface{})}
}

func (s *shard) get(k string) (data interface{}, ok bool) {
	s.L.RLock()
	data, ok = s.data[k]
	s.L.RUnlock()
	return
}

func (s *shard) set(k string, value interface{}) {
	s.L.Lock()
	s.data[k] = value
	s.L.Unlock()
}

func (s *shard) del(k string) {
	s.L.Lock()
	delete(s.data, k)
	s.L.Unlock()
}

type shardedMap struct {
	hsh        hasher
	shards     []*shard
	shardCount uint64
}

func (s shardedMap) getShard(h uint64) *shard {
	return s.shards[h%s.shardCount]
}

func (s shardedMap) Get(key string) (interface{}, bool) {
	h := s.hsh.HashString(key)
	return s.getShard(h).get(key)
}

func (s shardedMap) Set(key string, value interface{}) {
	h := s.hsh.HashString(key)
	s.getShard(h).set(key, value)
}

func (s shardedMap) Del(key string) {
	h := s.hsh.HashString(key)
	s.getShard(h).del(key)
}

func NewShardedMap(shards uint64) Map {
	m := &shardedMap{
		hsh:        &betterHasher{},
		shards:     make([]*shard, shards, shards),
		shardCount: shards,
	}
	for i := range m.shards {
		m.shards[i] = newShard()
	}
	return m
}

func NewDefaultShardedMap() Map {
	const shards = 256
	m := &shardedMap{
		hsh:        &betterHasher{},
		shards:     make([]*shard, shards, shards),
		shardCount: shards,
	}
	for i := range m.shards {
		m.shards[i] = newShard()
	}
	return m
}
