package internal

import "hash/fnv"

type Hashable interface {
	Hash() HashKey
}

type HashKey struct {
	value uint64
}

func (d Doc) Hash() HashKey {
	hash := fnv.New64()
	_, _ = hash.Write([]byte(d.Id()))
	return HashKey{value: hash.Sum64()}
}
