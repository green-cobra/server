package proxy

import (
	"github.com/rs/zerolog"
	"go-server/pkg/services"
	"go-server/pkg/services/origin"
	"sync"
)

type ConnectionStats struct {
	Addr        string
	Connections int
}

type TcpProxyManager struct {
	logger zerolog.Logger

	conf *Config

	createMut  *sync.RWMutex
	instances  map[string]*TcpProxyInstance
	takenPorts map[int]string
}

func NewTcpProxyManager(logger zerolog.Logger, proxyConf *Config) *TcpProxyManager {
	return &TcpProxyManager{
		logger:     logger,
		instances:  make(map[string]*TcpProxyInstance),
		takenPorts: make(map[int]string),
		conf:       proxyConf,
		createMut:  &sync.RWMutex{},
	}
}

func (t *TcpProxyManager) New(tunnelId string, origin *origin.Meta) *TcpProxyInstance {
	if len(t.instances) >= t.conf.MaxClients() {
		return nil
	}

	t.createMut.Lock()
	defer t.createMut.Unlock()
	port := services.GenerateOpenedPortNumber(t.conf.MinPort, t.conf.MaxPort)
	for {
		_, ok := t.takenPorts[port]
		if !ok {
			break
		}

		port = services.GenerateOpenedPortNumber(t.conf.MinPort, t.conf.MaxPort)
	}
	t.takenPorts[port] = tunnelId

	t.instances[tunnelId] = NewTcpProxyInstance(t.logger, port, t.conf, tunnelId, origin)
	go func() {
		<-t.instances[tunnelId].OnClose()
		delete(t.instances, tunnelId)
	}()

	return t.instances[tunnelId]
}

func (t *TcpProxyManager) Exists(host string) bool {
	t.createMut.RLock()
	defer t.createMut.RUnlock()
	_, ok := t.instances[host]
	return ok
}

func (t *TcpProxyManager) Get(host string) *TcpProxyInstance {
	t.createMut.RLock()
	defer t.createMut.RUnlock()
	v, _ := t.instances[host]
	return v
}

func (t TcpProxyManager) GetRunning() int {
	t.createMut.RLock()
	defer t.createMut.RUnlock()
	return len(t.instances)
}

func (t TcpProxyManager) GetConnectionsStats() []ConnectionStats {
	t.createMut.RLock()
	defer t.createMut.RUnlock()

	details := make([]ConnectionStats, 0, len(t.instances))

	for _, instance := range t.instances {
		details = append(details, ConnectionStats{
			Connections: instance.Connections(),
			Addr:        instance.GetAddr(),
		})
	}

	return details
}
