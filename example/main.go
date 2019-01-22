package main

import (
	"github.com/koykov/tcp-serv"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Simple echo handler.
// Just returns back what you said.
type EchoHandler struct {}

// The constructor.
func NewEchoHandler() *EchoHandler {
	h := EchoHandler{}
	return &h
}

// Main handler method.
func (h *EchoHandler) Handle(data []byte) (out []byte, err error) {
	out = []byte("you said: " + string(data))
	return
}

func main() {
	// Init.
	logger := io.Writer(os.Stdout)
	h := NewEchoHandler()
	s := tcp_serv.NewServer(":9000", time.Second * 5, tcp_serv.BufSize)
	s.SetLogger(&logger)

	// Listen.
	err := s.ListenAndServe(h)
	if err != nil {
		log.Println(err)
	}

	// Graceful shutdown.
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	go func(s *tcp_serv.Server) {
		<-c
		s.Shutdown()
		os.Exit(0)
	}(s)
}
