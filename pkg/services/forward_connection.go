package services

import (
	"io"
	"net"
	"time"
)

type forwardConnection struct {
	conn  net.Conn
	inUse bool
	alive bool
}

func (c *forwardConnection) Acquire() {
	c.inUse = true
}

func (c *forwardConnection) Release() {
	c.inUse = false
}

func (c *forwardConnection) Close() {
	if !c.alive {
		return
	}

	_ = c.conn.Close()
	c.alive = false
}

func (c *forwardConnection) updateDeadlines() {
	delay := 100 * time.Millisecond

	c.conn.SetReadDeadline(time.Now().Add(delay))
	c.conn.SetWriteDeadline(time.Now().Add(delay))
}

func (c *forwardConnection) Write(data []byte) error {
	c.updateDeadlines()

	_, err := c.conn.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (c *forwardConnection) MarkClosed() {
	c.alive = false
}

func (c *forwardConnection) Read() (error, []byte) {
	c.updateDeadlines()

	data, err := io.ReadAll(c.conn)
	if err != nil {
		return err, data
	}

	return nil, data
}
