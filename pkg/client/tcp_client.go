package client

import (
	"bytes"
	"context"
	"github.com/sonirico/visigoth/internal/server"
	"io"
	"log"
	"net"
	"sync"

	"github.com/sonirico/visigoth/internal/search"
	"github.com/sonirico/visigoth/pkg/entities"
	"github.com/sonirico/visigoth/pkg/vtp"
)

type callback func(result vtp.Message)

type errorCallback func(error)

type TCPClientConfig struct {
	bindTo       string
	readPoolSize *int
	proxyStream  bool
}

type TCPClient struct {
	counter *atomicCounter

	transport server.Transport

	bindTo       string
	readPoolSize int

	requests    chan vtp.Message
	responses   chan vtp.Message
	proxyStream bool

	callbacks       map[uint64]callback
	callbackLock    sync.RWMutex
	onErrorCallback errorCallback
}

func NewTCPClient(config *TCPClientConfig) *TCPClient {
	client := &TCPClient{
		requests:  make(chan vtp.Message),
		responses: make(chan vtp.Message),

		bindTo:      config.bindTo,
		proxyStream: config.proxyStream,

		transport: server.NewVTPTransport(),
		counter:   new(atomicCounter),
		callbacks: make(map[uint64]callback),
	}

	if config.readPoolSize != nil {
		client.readPoolSize = *config.readPoolSize
	} else {
		client.readPoolSize = 1
	}

	client.onErrorCallback = func(err error) { log.Println("client error", err) }

	return client
}

func (c *TCPClient) readServer(in io.Reader) error {
	for {
		message, err := c.transport.Parse(in)
		if err != nil {
			return err
		}
		c.responses <- message
	}
}

func (c *TCPClient) writeServer(out io.Writer) {
	buf := new(bytes.Buffer)
	for msg := range c.requests {
		buf.Reset()
		if err := c.transport.Compile(buf, msg); err != nil {
			log.Println(err)
			continue
		}

		if _, err := out.Write(buf.Bytes()); err != nil {
			log.Printf("error when messaging server: %s\n", err)
		}
	}
}

func (c *TCPClient) registerCallback(id uint64, cb callback) {
	c.callbackLock.Lock()
	c.callbacks[id] = cb
	c.callbackLock.Unlock()
}

func (c *TCPClient) connect(ctx context.Context) io.ReadWriteCloser {
	conn, err := net.Dial("tcp", c.bindTo)
	if err != nil {
		// TODO: reconnect on error
		c.onErrorCallback(err)
		return nil
	}
	return conn
}

func (c *TCPClient) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	conn, err := net.Dial("tcp", c.bindTo)
	if err != nil {
		// TODO: reconnect on error
		c.onErrorCallback(err)
		return
	}

	if !c.proxyStream {
		for i := 0; i < c.readPoolSize; i++ {
			go c.dispatchResponses(ctx)
		}
	}

	go func() {
		if err := c.readServer(conn); err != nil {
			c.onErrorCallback(err)
			return
		}
	}()
	c.writeServer(conn)
}

func (c *TCPClient) Drop(index string, cb callback) {
	msg := vtp.NewDropIndexRequest(c.counter.Inc(), Version, index)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.Id(), cb)
	}
}

func (c *TCPClient) ShowIndices(cb callback) {
	msg := vtp.NewListIndicesRequest(c.counter.Inc(), Version)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.Id(), cb)
	}
}

func (c *TCPClient) Alias(index, alias string, cb callback) {
	msg := vtp.NewAliasRequest(c.counter.Inc(), Version, index, alias)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.Id(), cb)
	}
}

func (c *TCPClient) UnAlias(alias string, cb callback) {
	msg := vtp.NewUnAliasRequest(c.counter.Inc(), Version, alias)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.Id(), cb)
	}
}

func (c *TCPClient) Index(index, name, payload string, format entities.MimeType, cb callback) {
	msg := vtp.NewIndexRequest(c.counter.Inc(), Version, index, name, payload, format)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.Id(), cb)
	}
}

func (c *TCPClient) Search(index, terms string, engine search.EngineType, cb callback) {
	msg := vtp.NewSearchRequest(c.counter.Inc(), Version, uint8(engine), index, terms)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.Id(), cb)
	}
}

func (c *TCPClient) Request(msg vtp.Message) {
	c.requests <- msg
}

func (c *TCPClient) Responses() <-chan vtp.Message {
	return c.responses
}

func (c *TCPClient) Error(fn errorCallback) {
	c.onErrorCallback = fn
}

func (c *TCPClient) dispatchResponses(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case res := <-c.responses:
			c.callbackLock.RLock()
			if cb, ok := c.callbacks[res.Id()]; ok {
				cb(res)
			}
			c.callbackLock.RUnlock()
		}
	}
}
