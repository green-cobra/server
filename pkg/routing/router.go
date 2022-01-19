package routing

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"go-server/pkg/controllers"
	"go-server/pkg/services"
)

func GetRouter(pc *services.ProxyConfig, logger zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	proxyManager := services.NewTcpProxyManager(logger.With().Str("module", "proxy-manager").Logger(), pc)

	tunnelController := controllers.NewTunnelController(logger.With().Str("module", "controller:tunnel").Logger(), proxyManager)

	r.Post("/api/v1/tunnel", tunnelController.CreateConnection)

	r.Get("/*", tunnelController.TryProxy)
	r.Post("/*", tunnelController.Proxy)
	r.Delete("/*", tunnelController.Proxy)
	r.Patch("/*", tunnelController.Proxy)
	r.Put("/*", tunnelController.Proxy)
	r.Connect("/*", tunnelController.Proxy)
	r.Head("/*", tunnelController.Proxy)

	return r
}
