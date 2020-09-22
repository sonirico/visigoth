package container

import (
	"github.com/sonirico/visigoth/pkg/entities"
)

type ResultIterator struct {
	res     entities.Result
	counter int
}

func NewResultIterator(res entities.Result) *ResultIterator {
	return &ResultIterator{
		res:     res,
		counter: -1,
	}
}

func (ri *ResultIterator) Next() (item entities.Row, done bool) {
	ri.counter++
	return ri.res.Get(ri.counter), ri.counter >= ri.res.Len()-1
}
