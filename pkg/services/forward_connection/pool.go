package forward_connection

import (
	"net"
	"sync"
	"time"
)

type ForwardConnectionsPool struct {
	m     sync.RWMutex
	conns []ForwardConnection
}

func NewForwardConnectionsPool() *ForwardConnectionsPool {
	p := &ForwardConnectionsPool{
		conns: make([]ForwardConnection, 0, 10),
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

func (f *ForwardConnectionsPool) Close() {
	f.m.Lock()
	defer f.m.Unlock()

	for _, v := range f.conns {
		v.Close()
	}
}

func (f *ForwardConnectionsPool) gcClosedConnections() {
	f.m.RLock()
	defer f.m.RUnlock()

	conns := make([]ForwardConnection, 0, 10)
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

func (f *ForwardConnectionsPool) Append(c net.Conn) error {
	f.m.Lock()
	defer f.m.Unlock()

	f.conns = append(f.conns, &tcpForwardConnection{conn: c, inUse: false, alive: true})
	return nil
}

func (f *ForwardConnectionsPool) Get() ForwardConnection {
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

func (f *ForwardConnectionsPool) Size() int {
	return len(f.conns)
}
