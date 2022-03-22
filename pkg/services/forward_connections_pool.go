package services

import (
	"net"
	"sync"
	"time"
)

type forwardConnectionsPool struct {
	m     sync.RWMutex
	conns []forwardConnection
}

func newForwardConnectionsPool() *forwardConnectionsPool {
	p := &forwardConnectionsPool{
		conns: make([]forwardConnection, 0, 10),
		m:     sync.RWMutex{},
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
		v.Close()
	}
}

func (f *forwardConnectionsPool) gcClosedConnections() {
	f.m.RLock()
	defer f.m.RUnlock()

	conns := make([]forwardConnection, 0, 10)
	for _, v := range f.conns {
		if v.InUse() {
			continue
		}

		if v.Alive() {
			conns = append(conns, v)
		}
	}

	f.conns = conns
}

func (f *forwardConnectionsPool) Append(c net.Conn) error {
	f.m.Lock()
	defer f.m.Unlock()

	f.conns = append(f.conns, &tcpForwardConnection{conn: c, inUse: false, alive: true})
	return nil
}

func (f *forwardConnectionsPool) Get() forwardConnection {
	f.m.Lock()
	defer f.m.Unlock()

	for _, v := range f.conns {
		if v.InUse() || !v.Alive() {
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
