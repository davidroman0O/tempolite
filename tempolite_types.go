package tempolite

import (
	"context"
	"log"
	"reflect"
	"time"
)

type HandlerInfo struct {
	Handler    interface{}
	ParamType  reflect.Type
	ReturnType reflect.Type
	NumIn      int
	NumOut     int
}

type SageInfo struct{}

type HandlerContext struct {
	context.Context
	tp                 *Tempolite
	taskID             string
	executionContextID string
}

func (c *HandlerContext) GetID() string {
	return c.taskID
}

type EnqueueOption func(*enqueueOptions)

type enqueueOptions struct {
	maxDuration    time.Duration
	timeLimit      time.Duration
	immediate      bool
	panicOnTimeout bool
}

func WithMaxDuration(duration time.Duration) EnqueueOption {
	return func(o *enqueueOptions) {
		log.Printf("Setting max duration for enqueue option: %v", duration)
		o.maxDuration = duration
	}
}

func WithTimeLimit(limit time.Duration) EnqueueOption {
	return func(o *enqueueOptions) {
		log.Printf("Setting time limit for enqueue option: %v", limit)
		o.timeLimit = limit
	}
}

func WithImmediateRetry() EnqueueOption {
	return func(o *enqueueOptions) {
		log.Printf("Enabling immediate retry for enqueue option")
		o.immediate = true
	}
}

func WithPanicOnTimeout() EnqueueOption {
	return func(o *enqueueOptions) {
		log.Printf("Enabling panic on timeout for enqueue option")
		o.panicOnTimeout = true
	}
}
