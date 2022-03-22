package proxy

import (
	"github.com/rs/zerolog"
	"go-server/pkg/services/origin"
)

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
