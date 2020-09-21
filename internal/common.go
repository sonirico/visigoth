package internal

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
	Next() (Row, bool)
}

type Result interface {
	Len() int
	Get(index int) Row
}
