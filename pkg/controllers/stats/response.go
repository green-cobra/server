package stats

import "go-server/pkg/services/proxy"

type Response struct {
	ProxiesRunning int                     `json:"proxies_running"`
	Stats          []proxy.ConnectionStats `json:"stats"`
}
