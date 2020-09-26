package client

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/sonirico/visigoth/pkg/vtp"
)

const (
	Version = 0
	Prompt  = "> "
)

type Repl interface {
	Repl(in io.Reader, out io.Writer)
}

type Scanner interface {
	Text() string
	Scan() bool
}

type CmdClient struct {
	client    *TCPClient
	wg        *sync.WaitGroup
	evaluator *commandEvaluator
}

func NewCmdClient(bindTo string) *CmdClient {
	env := newEnv()
	return &CmdClient{
		client: NewTCPClient(&TCPClientConfig{
			bindTo:      bindTo,
			proxyStream: true,
		}),
		wg:        &sync.WaitGroup{},
		evaluator: newCommandEvaluator(env),
	}
}

func (c *CmdClient) eval(cmd string) {
	msgs := c.evaluator.Eval(cmd)
	for _, msg := range msgs {
		c.wg.Add(1)
		c.client.Request(msg)
	}
}

func (c *CmdClient) print(out io.Writer) {
	for res := range c.client.Responses() {
		Print(out, res, c.wg)
	}
}

func (c *CmdClient) read(scanner Scanner) string {
	_ = scanner.Scan()
	return scanner.Text()
}

func (c *CmdClient) Repl(in io.Reader, out io.Writer) {
	ctx := context.Background()
	go c.client.Start(ctx)

	scanner := bufio.NewScanner(in)

	go c.print(out)

	for {
		c.wg.Wait() // Wait for all responses to be processed (eval + print) before reading next line
		PrintEnv(out, c.evaluator.env)
		printSafe(out, Prompt)
		line := c.read(scanner)
		c.eval(line)
	}
}

func Eval(cmd string, cmdClient *CmdClient, group *sync.WaitGroup) {
	msgs := cmdClient.evaluator.Eval(cmd)
	for _, msg := range msgs {
		group.Add(1)
		cmdClient.client.Request(msg)
	}
}

func PrintEnv(out io.Writer, env *environment) {
	index := "<none>"
	var buf bytes.Buffer
	buf.WriteString("\n+-------------------------+\n|")
	if env.Index != nil {
		index = *env.Index
	}
	buf.WriteString(fmt.Sprintf("index: %s |", index))
	buf.WriteString("\n+-------------------------+\n\n")
	printSafe(out, buf.String())
}

func Print(out io.Writer, msg vtp.Message, group *sync.WaitGroup) {
	switch msg.Type() {
	case vtp.ListRes:
		res, _ := msg.(*vtp.ListIndicesResponse)
		for i, index := range res.Indices {
			printSafe(out, fmt.Sprintf("%d) %s\n", i+1, index.Value))
		}
	case vtp.SearchRes:
		switch res := msg.(type) {
		case *vtp.HitsSearchResponse:
			printSafe(out, fmt.Sprintf("\ntotal: %d\n", len(res.Documents)))
			printSafe(out, "---------------\n")
			for _, row := range res.Documents {
				printSafe(out, fmt.Sprintf("{Name=%s,Hits=%d,Content=%s}\n",
					row.Document.Name.Value, row.Hits.Value, row.Document.Content.Value))
			}
			printSafe(out, "---------------\n\n")
		}
	default:
		_, _ = out.Write([]byte(msg.String()))
	}
	group.Done()
}

func printSafe(out io.Writer, msg string) {
	if _, err := out.Write([]byte(msg)); err != nil {
		log.Printf("unable to write '%s'\n", err)
		log.Println(msg)
	}
}
