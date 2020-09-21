package vtp

type MetaType uint8

const (
	Varchar MetaType = iota
	Blob
	I8
	I32
	I64
)

type Type interface {
	Type() Type
	Len() int
}

type StringType struct {
	Value string
}

func (sv StringType) Type() MetaType {
	return Varchar
}

func (sv StringType) Len() int {
	return len(sv.Value)
}

type ByteType struct {
	Value uint8
}

func (b ByteType) Type() MetaType {
	return I8
}

func (b ByteType) Len() int {
	return 1
}

func (b ByteType) Clone() *ByteType {
	return &ByteType{Value: b.Value}
}

type UInt32Type struct {
	Value uint32
}

func (u32 UInt32Type) Type() MetaType {
	return I32
}

func (u32 UInt32Type) Len() int {
	return 4
}

type UInt64Type struct {
	Value uint64
}

func (i UInt64Type) Type() MetaType {
	return I64
}

func (i UInt64Type) Len() int {
	return 8
}

type BlobType struct {
	Value []byte
}

func (b BlobType) Type() MetaType {
	return Blob
}

func (b BlobType) Len() int {
	return len(b.Value)
}
