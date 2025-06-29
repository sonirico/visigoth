package visigoth

type LengthWise interface {
	Len() int
}

type Row interface {
	Doc() Doc
}

type Result interface {
	Len() int
	Get(index int) Row
}
