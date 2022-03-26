package origin

import (
	"net"
	"net/url"
	"strings"
)

type Meta struct {
	url *url.URL
	ip  net.IP
}

func NewMeta(url *url.URL, ip net.IP) *Meta {
	return &Meta{
		url: url,
		ip:  ip,
	}
}

func (m Meta) Host() string {
	// Host format is: host.com:443
	parts := strings.Split(m.url.Host, ":")

	return parts[0]
}
func (m Meta) Port() string {
	// Host format is: host.com:443
	parts := strings.Split(m.url.Host, ":")

	return parts[1]
}

func (m Meta) Scheme() string {
	return m.url.Scheme
}

func (m Meta) IP() net.IP {
	return m.ip
}
