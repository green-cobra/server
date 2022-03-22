package tunnel

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

type tunnelRequest struct {
	Name string `json:"name"`

	originalIP net.IP
	originURL  *url.URL
}

func newTunnelRequest(r *http.Request) tunnelRequest {
	tq := tunnelRequest{}
	tq.withMeta(r)
	return tq
}

func (t *tunnelRequest) withMeta(r *http.Request) *tunnelRequest {
	r.URL.Host = r.Host

	//TODO: read from request
	r.URL.Scheme = "http"

	parts := strings.Split(r.RemoteAddr, ":")
	t.originalIP = net.ParseIP(parts[0])

	t.originURL = r.URL

	return t
}
