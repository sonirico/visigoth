package server

import (
	"context"
	"io"
	"log"
	"net"
)

type Server interface {
	Serve(ctx context.Context)
}

type handler interface {
	Handle(context.Context, io.ReadWriteCloser, Node)
}

type TCPServer struct {
	url       string
	node      Node
	transport *VTPTransport
}

func NewTCPServer(url string, node Node, tr *VTPTransport) *TCPServer {
	return &TCPServer{url: url, node: node, transport: tr}
}

func (s *TCPServer) Serve(ctx context.Context) {
	link, err := net.Listen("tcp", s.url)
	if err != nil {
		log.Fatalln("tcpServer, listen", err)
	}

	defer func(listener net.Listener) {
		if err := listener.Close(); err != nil {
			log.Fatalln("tcpServer, serve", err)
		}
	}(link)

	var clientCounter uint64
	for {
		conn, err := link.Accept()
		select {
		case <-ctx.Done():
			if err := conn.Close(); err != nil {
				log.Println("tcpserver, shutdown", err)
			}
			return
		default:
			if err != nil {
				log.Println("error on accept")
				log.Fatalln(err)
			}
			client := NewTCPClient(clientCounter, s.transport)
			go s.accept(ctx, conn, client)
			clientCounter++
		}
	}
}

func (s *TCPServer) accept(ctx context.Context, wire io.ReadWriteCloser, h handler) {
	h.Handle(ctx, wire, s.node)
}
