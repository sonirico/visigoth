package container

import (
	"github.com/sonirico/visigoth/internal"
)

type ResultIterator struct {
	res     internal.Result
	counter int
}

func NewResultIterator(res internal.Result) *ResultIterator {
	return &ResultIterator{
		res:     res,
		counter: -1,
	}
}

func (ri *ResultIterator) Next() (item internal.Row, done bool) {
	ri.counter++
	return ri.res.Get(ri.counter), ri.counter >= ri.res.Len()-1
}
