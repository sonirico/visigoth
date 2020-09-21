package server

import (
	"io"
	"log"
	"net"
)

type Server interface {
	Serve()
}

type handler interface {
	Handle(io.ReadWriteCloser, Node)
}

type TcpServer struct {
	url       string
	node      Node
	transport *VTPTransport
}

func NewTcpServer(url string, node Node, tr *VTPTransport) *TcpServer {
	return &TcpServer{url: url, node: node, transport: tr}
}

func (s *TcpServer) Serve() {
	link, err := net.Listen("tcp", s.url)
	if err != nil {
		log.Fatalln(err)
	}

	defer func(listener net.Listener) {
		if err := listener.Close(); err != nil {
			log.Fatalln(err)
		}
	}(link)

	var clientCounter uint64 = 0

	for {
		conn, err := link.Accept()
		if err != nil {
			log.Println("error on accept")
			log.Fatalln(err)
		}
		client := NewTcpClient(clientCounter, s.transport)
		go s.accept(conn, client)
		clientCounter++
	}
}

func (s *TcpServer) accept(wire io.ReadWriteCloser, h handler) {
	h.Handle(wire, s.node)
}
