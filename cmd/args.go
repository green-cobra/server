package cmd

import "go-server/pkg/services"
import flag "github.com/spf13/pflag"

var (
	minPort = flag.Int("min-port", 30000, "Minimal port to generate socket addrs")
	maxPort = flag.Int("max-port", 30100, "Max port to generate socket addrs")

	maxConnsPerClient = flag.Int("max-client-conns", 10, "Max connections per client")

	baseDomain = flag.String("domain", "", "Domain override for URL")

	listenPort = flag.Int("listen-port", 3001, "Port for API to listen")
	listenHost = flag.String("listen-host", "0.0.0.0", "Host for API to listen")

	timeoutInactiveHours   = flag.Int("timeout-inactive-hours", 24, "Number of hours to wait before closing client sockets")
	timeoutNoActiveSockets = flag.Int("timeout-inactive-sockets", 10, "Number of minutes between checks to wait before treating client as inactive")
	noActiveSocketsChecks  = flag.Int("checks-inactive-sockets", 3, "Number of checks to wait before treating client as inactive")
)

type ServerConfig struct {
	ListenPort int
	ListenHost string
}

func ParseArgs() (*services.ProxyConfig, *ServerConfig) {
	flag.Parse()

	return &services.ProxyConfig{
		MinPort:                       *minPort,
		MaxPort:                       *maxPort,
		BaseDomain:                    *baseDomain,
		MaxConnsPerClient:             *maxConnsPerClient,
		InactiveHoursTimeout:          *timeoutInactiveHours,
		NoActiveSocketsChecks:         *noActiveSocketsChecks,
		NoActiveSocketsMinutesTimeout: *timeoutNoActiveSockets,
	}, &ServerConfig{ListenPort: *listenPort, ListenHost: *listenHost}
}
