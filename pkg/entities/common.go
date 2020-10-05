package entities

type Serializer interface {
	Serialize(item Row) []byte
}

type LengthWise interface {
	Len() int
}

type Row interface {
	Doc() Doc
	Ser(serializer Serializer) []byte
}

type Iterator interface {
	Pos() int
	Iterable() Result
	Next() (Row, bool)
	Chain(Iterator) Iterator
}

type Result interface {
	Len() int
	Get(index int) Row
}
