package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/sonirico/visigoth/pkg/vtp"
)

type Client interface {
	Handle(wire io.ReadWriteCloser, node Node)
}

type TcpClient struct {
	id        uint64
	requests  chan vtp.Message
	responses chan vtp.Message
	transport *VTPTransport
}

func NewTcpClient(id uint64, transport *VTPTransport) *TcpClient {
	return &TcpClient{
		id:        id,
		requests:  make(chan vtp.Message), // TODO: configure size, otherwise new data will not be parsed due to the unbuffered channel
		responses: make(chan vtp.Message),
		transport: transport,
	}
}

func (c *TcpClient) String() string {
	return fmt.Sprintf("client{id=%d,reqBuf=%d,resBuf=%d}",
		c.id, len(c.requests), len(c.responses))
}

func (c *TcpClient) Handle(ctx context.Context, wire io.ReadWriteCloser, node Node) {
	log.Println(c, "connected")
	defer func() {
		err := recover()
		if err != nil {
			log.Println("tcpServerClient, got error", err)
		}
		if err := wire.Close(); err != nil {
			log.Println("tcpServerClient, wire close", err)
		}
	}()
	go c.read(ctx, wire, node)
	c.write(ctx, wire)
}

func (c *TcpClient) read(ctx context.Context, wire io.Reader, node Node) {
	go func() {
		if err := vtp.ParseStream(ctx, wire, c.transport.Parser, c.requests); err != nil {
			close(c.requests)
			close(c.responses)

			if err == io.EOF {
				log.Println(c, "disconnected")
				return
			}
			if err == io.ErrUnexpectedEOF {
				log.Println(fmt.Sprintf("client parser error with id %d, %s", c.id, err.Error()))
				return
			}

			log.Println("tcpServerClient, read", err)
		}
	}()

	node.Run(c.requests, c.responses, &NodeConfig{tracer: c.trace})
}

func (c *TcpClient) write(ctx context.Context, wire io.Writer) {
	buf := new(bytes.Buffer)
	for {
		select {
		case <-ctx.Done():
			return
		case res, ok := <-c.responses:
			buf.Reset()
			if !ok {
				return
			}
			if err := c.transport.Compile(buf, res); err != nil {
				log.Printf("unable to encode response: %s", res)
				continue
			}

			written, err := wire.Write(buf.Bytes())

			if err != nil {
				log.Fatalln(err)
				return
			}
			if written != buf.Len() {
				log.Fatalln("written distinct than encoded")
			}
		}
	}
}

func (c *TcpClient) trace(msg vtp.Message) {
	log.Println(fmt.Sprintf("%s -> %s", c, vtp.MessageToString(msg)))
}
