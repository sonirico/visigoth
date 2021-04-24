package engines

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestNewLSMTEngine_Write(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	engi := NewLSMTEngine(lsmtOptions{SegmentFactory: newBasicLSMTSegment})
	go engi.run(ctx)
	<-time.After(time.Second * 1)
	if err := engi.Insert("temperatura", "10"); err != nil {
		log.Println(err)
	}
	<-time.After(time.Second * 1)
	if r, err := engi.Search("temperaturas"); err != nil {
		log.Println(err)
	} else {
		log.Println("result found", r)
	}

	<-time.After(time.Second * 10)
}

func TestNewLSMTEngine_Read(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*50)
	engi := NewLSMTEngine(lsmtOptions{SegmentFactory: newBasicLSMTSegment})
	go engi.run(ctx)
	<-time.After(time.Second * 1)
	if r, err := engi.Search("temperaturas"); err != nil {
		log.Println("search error", err)
	} else {
		log.Println("result found", r)
	}

	<-time.After(time.Second * 10000)
}
