// Copyright © 2022 Sloan Childers
package sink

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

type ShutdownHandler struct {
	listeners []func()
}

func NewShutdownHandler() *ShutdownHandler {
	return &ShutdownHandler{}
}

func (x *ShutdownHandler) AddListener(f func()) {
	x.listeners = append(x.listeners, f)
}

func (x *ShutdownHandler) Listen() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigs
		log.Printf("[SHUTDOWN] shutting down in 15 seconds: %s", s)
		for _, ShutItDown := range x.listeners {
			ShutItDown()
		}
		os.Exit(0)
	}()
}
