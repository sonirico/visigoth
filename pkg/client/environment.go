package client

import "time"

type environment struct {
	Index     *string
	TouchedAt *time.Time
}

func newEnv() *environment {
	return &environment{
		Index:     nil,
		TouchedAt: nil,
	}
}
