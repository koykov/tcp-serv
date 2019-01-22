package tcp_serv

import (
	"io"
	"net"
	"time"
)

// Wrapper over TCP connection with support of idle timeout and incoming data limit.
// updateTimeout() method should be called explicitly after each read/write operation.
type Conn struct {
	net.Conn
	idleTimeout time.Duration
	bytesLimit  int64
}

// The constructor.
func NewConn(raw *net.Conn, idleTimeout time.Duration, bytesLimit int64) *Conn {
	conn := Conn{
		Conn:        *raw,
		idleTimeout: idleTimeout,
		bytesLimit:  bytesLimit,
	}
	_ = conn.SetDeadline(time.Now().Add(conn.idleTimeout))

	return &conn
}

// Write data to connection.
func (c *Conn) Write(p []byte) (n int, err error) {
	if err = c.updateTimeout(); err != nil {
		return
	}
	n, err = c.Conn.Write(p)
	return
}

// Read data from connection.
func (c *Conn) Read(b []byte) (n int, err error) {
	if err = c.updateTimeout(); err != nil {
		return
	}
	r := io.LimitReader(c.Conn, c.bytesLimit)
	n, err = r.Read(b)
	return
}

// Close the connection.
func (c *Conn) Close() (err error) {
	err = c.Conn.Close()
	return
}

// Update timeout value.
func (c *Conn) updateTimeout() error {
	idleDeadline := time.Now().Add(c.idleTimeout)
	return c.Conn.SetDeadline(idleDeadline)
}
