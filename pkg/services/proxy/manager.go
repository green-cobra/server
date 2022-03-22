package proxy

import (
	"github.com/rs/zerolog"
	"go-server/pkg/services/origin"
)

type ConnectionStats struct {
	Addr        string
	Connections int
}

type TcpProxyManager struct {
	logger zerolog.Logger

	conf *Config

	instances map[string]*TcpProxyInstance
}

func NewTcpProxyManager(logger zerolog.Logger, proxyConf *Config) *TcpProxyManager {
	return &TcpProxyManager{
		logger:    logger,
		instances: make(map[string]*TcpProxyInstance),
		conf:      proxyConf,
	}
}

func (t *TcpProxyManager) New(tunnelId string, origin *origin.Meta) *TcpProxyInstance {
	if len(t.instances) >= t.conf.MaxClients() {
		return nil
	}

	t.instances[tunnelId] = NewTcpProxyInstance(t.logger, t.conf, tunnelId, origin)
	go func() {
		<-t.instances[tunnelId].OnClose()
		delete(t.instances, tunnelId)
	}()

	return t.instances[tunnelId]
}

func (t *TcpProxyManager) Exists(host string) bool {
	_, ok := t.instances[host]
	return ok
}

func (t *TcpProxyManager) Get(host string) *TcpProxyInstance {
	v, _ := t.instances[host]
	return v
}

func (t TcpProxyManager) GetRunning() int {
	return len(t.instances)
}

func (t TcpProxyManager) GetConnectionsStats() []ConnectionStats {
	details := make([]ConnectionStats, 0, len(t.instances))

	for _, instance := range t.instances {
		details = append(details, ConnectionStats{
			Connections: instance.Connections(),
			Addr:        instance.GetAddr(),
		})
	}

	return details
}
