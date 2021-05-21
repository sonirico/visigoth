package client

import (
	"bytes"
	"context"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/sonirico/visigoth/internal/server"

	"github.com/sonirico/visigoth/internal/search"
	"github.com/sonirico/visigoth/pkg/entities"
	"github.com/sonirico/visigoth/pkg/vtp"
)

type callback func(result vtp.Message)

type errorCallback func(error)

type TCPClientConfig struct {
	BindTo       string
	ReadPoolSize int
	ProxyStream  bool
}

type TCPClient struct {
	counter *atomicCounter

	conn      io.ReadWriteCloser
	transport server.Transport

	bindTo       string
	readPoolSize int
	proxyStream  bool

	close     chan struct{}
	requests  chan vtp.Message
	responses chan vtp.Message

	callbacks       map[uint64]callback
	callbackLock    *sync.Mutex
	onErrorCallback errorCallback
}

func NewTCPClient(config TCPClientConfig) *TCPClient {
	client := &TCPClient{
		requests:  make(chan vtp.Message),
		responses: make(chan vtp.Message),
		close:     make(chan struct{}),

		bindTo:      config.BindTo,
		proxyStream: config.ProxyStream,

		transport:    server.NewVTPTransport(),
		counter:      new(atomicCounter),
		callbacks:    make(map[uint64]callback),
		callbackLock: &sync.Mutex{},
	}

	if config.ReadPoolSize > 0 {
		client.readPoolSize = config.ReadPoolSize
	} else {
		client.readPoolSize = 1
	}

	client.onErrorCallback = func(err error) {
		log.Println("client error => ", err)
	}

	return client
}

func (c *TCPClient) readServer() {
	for {
		select {
		case <-c.close:
			return
		default:
			message, err := c.transport.Parse(c.conn)
			if err != nil {
				c.onErrorCallback(err)
				return
			}
			c.responses <- message
		}
	}
}

func (c *TCPClient) writeServer() {
	defer func() {
		log.Println("write closed")
	}()
	buf := new(bytes.Buffer)
	for {
		select {
		case <-c.close:
			return
		case msg := <-c.requests:
			buf.Reset()
			if err := c.transport.Compile(buf, msg); err != nil {
				c.onErrorCallback(err)
			} else if _, err := c.conn.Write(buf.Bytes()); err != nil {
				log.Printf("error when messaging server: %s\n", err)
				c.onErrorCallback(err)
			}
		}
	}
}

func (c *TCPClient) registerCallback(id uint64, cb callback) {
	c.callbackLock.Lock()
	defer c.callbackLock.Unlock()
	c.callbacks[id] = cb
}

func (c *TCPClient) connect() io.ReadWriteCloser {
	for {
		con, err := net.Dial("tcp", c.bindTo)
		if err != nil {
			// TODO: reconnect on error
			c.onErrorCallback(err)
			time.Sleep(time.Second)
			continue
		}
		if con != nil {
			return con
		}
	}
}

func (c *TCPClient) Start(_ context.Context) {
	c.conn = c.connect()
	if c.conn == nil {
		return
	}

	if !c.proxyStream {
		for i := 0; i < c.readPoolSize; i++ {
			go c.dispatchResponses()
		}
	}

	go c.readServer()
	c.writeServer()
}

func (c *TCPClient) Stop() {
	if err := c.conn.Close(); err != nil {
		c.onErrorCallback(err)
	}
	close(c.close)
}

func (c *TCPClient) Drop(index string, cb callback) {
	msg := vtp.NewDropIndexRequest(c.counter.Inc(), Version, index)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.ID(), cb)
	}
}

func (c *TCPClient) ShowIndices(cb callback) {
	msg := vtp.NewListIndicesRequest(c.counter.Inc(), Version)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.ID(), cb)
	}
}

func (c *TCPClient) Alias(index, alias string, cb callback) {
	msg := vtp.NewAliasRequest(c.counter.Inc(), Version, index, alias)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.ID(), cb)
	}
}

func (c *TCPClient) UnAlias(index, alias string, cb callback) {
	msg := vtp.NewUnAliasRequest(c.counter.Inc(), Version, index, alias)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.ID(), cb)
	}
}

func (c *TCPClient) Index(index, name, payload string, format entities.MimeType, cb callback) {
	msg := vtp.NewIndexRequest(c.counter.Inc(), Version, index, name, payload, format)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.ID(), cb)
	}
}

func (c *TCPClient) Search(index, terms string, engine search.EngineType, cb callback) {
	msg := vtp.NewSearchRequest(c.counter.Inc(), Version, uint8(engine), index, terms)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.ID(), cb)
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

func (c *TCPClient) dispatchResponses() {
	for {
		select {
		case <-c.close:
			return
		case res := <-c.responses:
			c.callbackLock.Lock()
			locked := true
			if cb, ok := c.callbacks[res.ID()]; ok {
				delete(c.callbacks, res.ID())
				// Prevents recursive locking by unlocking
				// prior to the call to 'cb'
				c.callbackLock.Unlock()
				locked = false
				if cb != nil {
					cb(res)
				}
			}
			if locked {
				c.callbackLock.Unlock()
			}
		}
	}
}
