package tcp_serv

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const (
	BufSize = 1024
)

// A Handler responds to TCP request.
type Handler interface {
	Handle([]byte) ([]byte, error)
}

// TCP server.
type Server struct {
	addr        string
	idleTimeout time.Duration
	bytesLimit  int64

	listener    net.Listener
	connections map[*Conn]bool
	mux         sync.Mutex
	stop        bool
	logger      *io.Writer
}

// The constructor.
func NewServer(address string, idleTimeout time.Duration, bytesLimit int64) *Server {
	s := Server{
		addr:        address,
		idleTimeout: idleTimeout,
		bytesLimit:  bytesLimit,
		connections: make(map[*Conn]bool),
		logger:      nil,
	}

	return &s
}

// Logger setter.
func (s *Server) SetLogger(logger *io.Writer) {
	s.logger = logger
}

// Listen given TCP address and run corresponding handler.
// ListenAndServe always returns a non-nil error.
func (s *Server) ListenAndServe(h Handler) (err error) {
	// Listen given port.
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return
	}
	defer func() {
		_ = s.listener.Close()
	}()

	s.Log(fmt.Sprintf("waiting for requests on %s", s.addr))

	// Waiting for new connections.
	for {
		if s.stop {
			break
		}

		// Accept new connection.
		connRaw, err := s.listener.Accept()
		if err != nil {
			s.Log(err)
			continue
		}
		s.Log(fmt.Sprintf("connection accepted from %s", connRaw.RemoteAddr()))
		conn := NewConn(&connRaw, s.idleTimeout, s.bytesLimit)
		s.addConn(conn)

		// Process each connection concurrently.
		go func(conn *Conn) {
			defer func(conn *Conn) {
				s.closeConn(conn)
				s.Log("connection closed")
			}(conn)

			r := bufio.NewReader(conn)
			w := bufio.NewWriter(conn)
			cbuf := make(chan []byte)
			timeout := time.After(s.idleTimeout)
			for {
				// Waiting for reading from the connection.
				go func(cbuf chan []byte, r *bufio.Reader) {
					buf := make([]byte, BufSize)
					_, err := r.Read(buf)
					if err != nil {
						s.Log(err)
					}
					buf = bytes.Trim(buf, "\x00")
					cbuf <- buf
				}(cbuf, r)

				select {
				case <-timeout:
					// Oops, we caught a timeout.
					err = io.EOF
					return
				case buf := <-cbuf:
					// Red buffer isn't empty, process the data.
					out, err := h.Handle(buf)
					if err != nil {
						s.Log(err)
						return
					}
					// Write response to connection.
					_, err = w.Write(out)
					if err != nil {
						s.Log(err)
						return
					}
					// Flush the connection.
					err = w.Flush()
					if err != nil {
						s.Log(err)
					}
					// Update timeout,
					timeout = time.After(s.idleTimeout)
				}
			}
		}(conn)
	}

	return
}

// Add new connection to the list.
func (s *Server) addConn(conn *Conn) {
	s.mux.Lock()
	s.connections[conn] = true
	s.mux.Unlock()
}

// Close connection.
func (s *Server) closeConn(conn *Conn) {
	_ = conn.Close()
	s.mux.Lock()
	delete(s.connections, conn)
	s.mux.Unlock()
}

// Shutdown the server.
func (s *Server) Shutdown() {
	s.stop = true
	s.Log("shutting down at", time.Now())
	_ = s.listener.Close()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.Log(fmt.Sprintf("waiting on %v connections", len(s.connections)))
		}
		if len(s.connections) == 0 {
			return
		}
	}
}

// Send message to the logger.
func (s *Server) Log(a ...interface{}) {
	if s.logger == nil {
		return
	}
	_, _ = fmt.Fprintln(*s.logger, a...)
}
