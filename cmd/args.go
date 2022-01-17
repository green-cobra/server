package cmd

import "go-server/pkg/services"
import flag "github.com/spf13/pflag"

var (
	minPort = flag.Int("min-port", 30000, "Minimal port to generate socket addrs")
	maxPort = flag.Int("max-port", 30100, "Max port to generate socket addrs")

	baseDomain = flag.String("domain", "", "Domain override for URL")

	listenPort = flag.Int("listen-port", 3000, "Port for API to listen")
)

type ServerConfig struct {
	ListenPort int
}

func ParseArgs() (*services.ProxyConfig, *ServerConfig) {
	flag.Parse()

	return &services.ProxyConfig{
		MinPort:    *minPort,
		MaxPort:    *maxPort,
		BaseDomain: *baseDomain,
	}, &ServerConfig{ListenPort: *listenPort}
}
