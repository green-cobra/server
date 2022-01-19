package services

import (
	"net"
	"sync"
	"time"
)

type forwardConnectionsPool struct {
	m     sync.Mutex
	conns []*forwardConnection
}

func newForwardConnectionsPool() *forwardConnectionsPool {
	p := &forwardConnectionsPool{
		conns: make([]*forwardConnection, 0, 10),
		m:     sync.Mutex{},
	}

	go func() {
		for {
			<-time.After(3 * time.Second)
			p.gcClosedConnections()
		}
	}()

	return p
}

func (f *forwardConnectionsPool) Close() {
	f.m.Lock()
	defer f.m.Unlock()

	for _, v := range f.conns {
		v.MarkClosed()
		v.conn.SetDeadline(time.Now())
		v.conn.Close()
	}
}

func (f *forwardConnectionsPool) gcClosedConnections() {
	f.m.Lock()
	defer f.m.Unlock()

	conns := make([]*forwardConnection, 0, 10)
	for _, v := range f.conns {
		if v.inUse {
			continue
		}

		if v.alive {
			conns = append(conns, v)
		}
	}

	f.conns = conns
}

func (f *forwardConnectionsPool) Append(c net.Conn) error {
	f.m.Lock()
	defer f.m.Unlock()

	f.conns = append(f.conns, &forwardConnection{conn: c, inUse: false, alive: true})
	return nil
}

func (f *forwardConnectionsPool) Get() *forwardConnection {
	f.m.Lock()
	defer f.m.Unlock()

	for _, v := range f.conns {
		if v.inUse || !v.alive {
			continue
		}

		v.Acquire()
		return v
	}

	return nil
}

func (f *forwardConnectionsPool) Size() int {
	return len(f.conns)
}
