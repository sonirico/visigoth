package client

import (
	"bytes"
	"encoding/binary"
	"github.com/sonirico/visigoth/internal"
	"github.com/sonirico/visigoth/internal/search"
	"github.com/sonirico/visigoth/pkg/vtp"
	"io"
	"log"
	"net"
	"sync"
)

var (
	compiler = vtp.NewCompiler(binary.BigEndian)
	parser   = vtp.NewParser(binary.BigEndian)
)

type Callback func(result vtp.Message)

type TCPClient struct {
	counter        *atomicCounter
	compiler       vtp.Compiler
	parser         vtp.Parser
	bindTo         string
	wg             *sync.WaitGroup
	requests       chan vtp.Message
	responses      chan vtp.Message
	responsesProxy chan vtp.Message
	callbacks      map[uint64]Callback
	callbackLock   sync.RWMutex
}

func NewTCPClient(bindTo string) *TCPClient {
	return &TCPClient{
		requests:       make(chan vtp.Message, 16),
		responses:      make(chan vtp.Message, 16),
		responsesProxy: make(chan vtp.Message, 16),
		bindTo:         bindTo,
		compiler:       compiler,
		parser:         parser,
		wg:             &sync.WaitGroup{},
		counter:        new(atomicCounter),
		callbacks:      make(map[uint64]Callback),
	}
}

func (c *TCPClient) readServer(in io.Reader) {
	readServer(in, c.responses, c.parser)
}

func (c *TCPClient) writeServer(out io.Writer) {
	writeServer(out, c.requests, c.compiler)
}

func (c *TCPClient) registerCallback(id uint64, cb Callback) {
	c.callbackLock.Lock()
	c.callbacks[id] = cb
	c.callbackLock.Unlock()
}

func (c *TCPClient) Start() {
	conn, err := net.Dial("tcp", c.bindTo)
	if err != nil {
		panic(err)
	}

	go c.readServer(conn)
	go c.dispatchResponses()
	c.writeServer(conn)
}

func (c *TCPClient) Drop(index string, cb Callback) {
	msg := vtp.NewDropIndexRequest(c.counter.Inc(), Version, index)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.Id(), cb)
	}
}

func (c *TCPClient) ShowIndices(cb Callback) {
	msg := vtp.NewListIndicesRequest(c.counter.Inc(), Version)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.Id(), cb)
	}
}

func (c *TCPClient) Alias(index, alias string, cb Callback) {
	msg := vtp.NewAliasRequest(c.counter.Inc(), Version, index, alias)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.Id(), cb)
	}
}

func (c *TCPClient) UnAlias(alias string, cb Callback) {
	msg := vtp.NewUnAliasRequest(c.counter.Inc(), Version, alias)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.Id(), cb)
	}
}

func (c *TCPClient) Index(index, name, payload string, format internal.MimeType, cb Callback) {
	msg := vtp.NewIndexRequest(c.counter.Inc(), Version, index, name, payload, format)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.Id(), cb)
	}
}

func (c *TCPClient) Search(index, terms string, engine search.EngineType, cb Callback) {
	msg := vtp.NewSearchRequest(c.counter.Inc(), Version, uint8(engine), index, terms)
	c.requests <- msg
	if cb != nil {
		c.registerCallback(msg.Id(), cb)
	}
}

func (c *TCPClient) Request(msg vtp.Message) {
	c.requests <- msg
}

func (c *TCPClient) Responses() chan vtp.Message {
	return c.responsesProxy
}

func (c *TCPClient) dispatchResponses() {
	for res := range c.responses {
		c.callbackLock.RLock()
		if cb, ok := c.callbacks[res.Id()]; ok {
			cb(res)
		}
		c.callbackLock.RUnlock()
		c.responsesProxy <- res
	}
}

func writeServer(out io.Writer, bus <-chan vtp.Message, compiler vtp.Compiler) {
	buf := new(bytes.Buffer)
	for msg := range bus {
		buf.Reset()
		if err := compiler.Compile(buf, msg); err != nil {
			log.Println(err)
			continue
		}

		if _, err := out.Write(buf.Bytes()); err != nil {
			log.Printf("error when messaging server: %s\n", err)
		}
	}
}

// Read polls from server connection and parses responses into messages to display
func readServer(bus io.Reader, messages chan vtp.Message, parser vtp.Parser) {
	if err := vtp.ParseStream(bus, parser, messages); err != nil {
		close(messages)

		log.Fatalln(err)
	}
}
