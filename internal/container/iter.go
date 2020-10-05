package container

import (
	"github.com/sonirico/visigoth/pkg/entities"
)

type ResultIterator struct {
	res     entities.Result
	counter int
	next    *ResultIterator
}

func NewResultIterator(res entities.Result) *ResultIterator {
	return &ResultIterator{
		res:     res,
		counter: -1,
		next:    nil,
	}
}

func (ri *ResultIterator) Next() (entities.Row, bool) {
	if ri.next != nil {
		item, done := ri.next.Next()
		if !done {
			return item, false
		} else if item != nil && done {
			ri.next = nil
			return item, false
		}
	}
	ri.counter++
	done := ri.counter >= ri.res.Len()-1
	return ri.res.Get(ri.counter), done
}

func (ri ResultIterator) Pos() int {
	return ri.counter
}

func (ri *ResultIterator) Iterable() entities.Result {
	return ri.res
}

func (ri *ResultIterator) Chain(iter entities.Iterator) entities.Iterator {
	return &ResultIterator{
		res:     iter.Iterable(),
		counter: iter.Pos(),
		next:    ri,
	}
}
