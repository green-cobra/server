package services

import (
	"io"
	"net"
	time "time"
)

type forwardConnection interface {
	Acquire()
	Release()
	Close() error
	Write(data []byte) error
	Read() (error, []byte)
	Alive() bool
	InUse() bool
}

type tcpForwardConnection struct {
	conn  net.Conn
	inUse bool
	alive bool
}

func (c *tcpForwardConnection) Alive() bool {
	return c.alive
}

func (c *tcpForwardConnection) InUse() bool {
	return c.inUse
}

func (c *tcpForwardConnection) Acquire() {
	c.inUse = true
}

func (c *tcpForwardConnection) Release() {
	c.inUse = false
}

func (c *tcpForwardConnection) Close() error {
	if !c.alive {
		return nil
	}
	c.markClosed()

	err := c.conn.SetDeadline(time.Now())
	if err != nil {
		return err
	}
	_ = c.conn.Close()
	return nil
}

func (c *tcpForwardConnection) updateDeadlines() {
	delay := 100 * time.Millisecond

	c.conn.SetReadDeadline(time.Now().Add(delay))
	c.conn.SetWriteDeadline(time.Now().Add(delay))
}

func (c *tcpForwardConnection) Write(data []byte) error {
	c.updateDeadlines()

	_, err := c.conn.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (c *tcpForwardConnection) markClosed() {
	c.alive = false
}

func (c *tcpForwardConnection) Read() (error, []byte) {
	c.updateDeadlines()

	data, err := io.ReadAll(c.conn)
	if err != nil {
		return err, data
	}

	return nil, data
}
