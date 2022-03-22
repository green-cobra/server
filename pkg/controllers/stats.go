package controllers

import (
	"github.com/rs/zerolog"
	"go-server/pkg/services/proxy"
)

type StatsController struct {
	logger zerolog.Logger

	proxyManager *proxy.TcpProxyManager
}

func NewStatsController(logger zerolog.Logger, proxyManager *proxy.TcpProxyManager) *StatsController {
	return &StatsController{logger: logger, proxyManager: proxyManager}
}
