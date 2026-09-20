package util

import (
	"context"
	"sync"
)

var (
	shutDownContext *ShutdownContext
	onceLock        sync.Once
)

type ShutdownContext struct {
	Context context.Context
	Cancel  context.CancelFunc
}

func initialize() {
	onceLock.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		shutDownContext = &ShutdownContext{ctx, cancel}
	})
}

func GetShutDownCtx() *ShutdownContext {
	initialize()
	return shutDownContext
}
