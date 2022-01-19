package services

import (
	"github.com/rs/zerolog"
	"net"
	"net/url"
)

type ProxyConfig struct {
	MinPort           int
	MaxPort           int
	BaseDomain        string
	MaxConnsPerClient int

	InactiveHoursTimeout          int
	NoActiveSocketsMinutesTimeout int
	NoActiveSocketsChecks         int
}

func (pc *ProxyConfig) MaxClients() int {
	return pc.MaxPort - pc.MinPort
}

type OriginMeta struct {
	Url *url.URL
	Ip  net.IP
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

func (t *TcpProxyManager) New(tunnelId string, origin *OriginMeta) *TcpProxyInstance {
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
