package origin

import (
	"net"
	"net/url"
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
	return m.url.Host
}

func (m Meta) Scheme() string {
	return m.url.Scheme
}

func (m Meta) IP() net.IP {
	return m.ip
}
