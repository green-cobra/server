package services

import (
	"github.com/rs/zerolog"
	"net/url"
)

type ProxyConfig struct {
	MinPort    int
	MaxPort    int
	BaseDomain string
}

type TcpProxyManager struct {
	logger zerolog.Logger

	conf *ProxyConfig

	instances map[string]*TcpProxyInstance
}

func NewTcpProxyManager(logger zerolog.Logger, proxyConf *ProxyConfig) *TcpProxyManager {
	return &TcpProxyManager{
		logger:    logger,
		instances: make(map[string]*TcpProxyInstance),
		conf:      proxyConf,
	}
}

func (t *TcpProxyManager) New(tunnelId string, originUrl *url.URL) *TcpProxyInstance {
	// TODO: close connection after timeout

	t.instances[tunnelId] = NewTcpProxyInstance(t.logger, t.conf, tunnelId, originUrl)
	go t.instances[tunnelId].listen()

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
