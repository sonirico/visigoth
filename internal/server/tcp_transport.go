package server

import (
	"encoding/binary"
	"io"

	"github.com/sonirico/visigoth/pkg/vtp"
)

type compiler interface {
	Compile(io.Writer, vtp.Message) error
}

type parser interface {
	Parse(io.Reader) (vtp.Message, error)
}

type Transport interface {
	compiler
	parser
}

type VTPTransport struct {
	Compiler vtp.ProtoCompiler
	Parser   vtp.ProtoParser
}

func NewVTPTransport() *VTPTransport {
	return &VTPTransport{
		Compiler: vtp.NewVTPCompiler(vtp.NewCompiler(binary.BigEndian)),
		Parser:   vtp.NewVTPParser(vtp.NewParser(binary.BigEndian)),
	}
}

func (v *VTPTransport) Compile(w io.Writer, m vtp.Message) error {
	return v.Compiler.Compile(w, m)
}

func (v *VTPTransport) Parse(r io.Reader) (vtp.Message, error) {
	return v.Parser.Parse(r)
}
