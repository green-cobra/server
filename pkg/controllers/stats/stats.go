package stats

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"go-server/pkg/services/proxy"
	"net/http"
)

type Controller struct {
	logger zerolog.Logger

	proxyManager *proxy.TcpProxyManager
}

func NewStatsController(logger zerolog.Logger, proxyManager *proxy.TcpProxyManager) *Controller {
	return &Controller{logger: logger, proxyManager: proxyManager}
}

func (s Controller) Get(w http.ResponseWriter, r *http.Request) {
	response := Response{}

	response.ProxiesRunning = s.proxyManager.GetRunning()
	response.Stats = s.proxyManager.GetConnectionsStats()

	bytes, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(500)
		s.logger.Err(err).Msg("failed to marshal response")

		return
	}

	w.Write(bytes)
}
