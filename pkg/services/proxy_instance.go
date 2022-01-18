package services

import (
	"fmt"
	"github.com/rs/zerolog"
	"net"
	"strings"
	"time"
)

type TcpProxyInstance struct {
	Port int
	ID   string

	conf *ProxyConfig

	logger   zerolog.Logger
	connPool *forwardConnectionsPool

	origin *OriginMeta
}

func NewTcpProxyInstance(logger zerolog.Logger, c *ProxyConfig, id string, origin *OriginMeta) *TcpProxyInstance {
	port := GenerateOpenedPortNumber(c.MinPort, c.MaxPort)

	tp := &TcpProxyInstance{
		Port:     port,
		ID:       id,
		conf:     c,
		logger:   logger,
		origin:   origin,
		connPool: newForwardConnectionsPool(),
	}

	go func() {
		err := tp.listen()
		if err != nil {
			tp.logger.Err(err).Msg("failed to listen")
		}
	}()
	return tp

}

func (s *TcpProxyInstance) MaxConns() int {
	return s.conf.MaxConnsPerClient
}

func (s *TcpProxyInstance) URL() string {
	domain := s.origin.Url.Host
	if s.conf.BaseDomain != "" {
		domain = s.conf.BaseDomain
	}

	return s.origin.Url.Scheme + "://" + s.ID + "." + domain
}

func (s *TcpProxyInstance) Proxy(data []byte) (error, []byte) {
	c := s.connPool.Get()
	attempt := 0
	for c == nil {
		if attempt >= 5 {
			return nil, nil
		}

		time.Sleep(200 * time.Millisecond)

		c = s.connPool.Get()
		attempt++
	}

	defer c.Release()

	err := c.Write(data)
	if err != nil {
		return err, nil
	}

	err, resp := c.Read()
	if err != nil {
		return err, resp
	}

	return nil, resp
}

func (s *TcpProxyInstance) listen() error {
	//TODO: close connection after some inactivity period
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		s.logger.Fatal().Int("port", s.Port).Err(err).Msg("Unable to create listener.")
		return err
	}
	defer listener.Close()
	s.logger.Info().Str("bind", listener.Addr().String()).Str("protocol", "tcp").Msg("Listening...")

	for {
		conn, err := listener.Accept()

		ra := conn.RemoteAddr().String()
		raParts := net.ParseIP(strings.Split(ra, ":")[0])
		if !s.origin.Ip.Equal(raParts) {
			s.logger.Warn().Err(err).Msg("Closing connection as IP doesnt match origin IP")
			err := conn.Close()
			if err != nil {
				s.logger.Error().Err(err).Msg("Error while closing connection.")
			}
			continue
		}

		if s.connPool.Size() >= s.conf.MaxConnsPerClient {
			// reject connection after 10 are opened
			s.logger.Debug().Err(err).Msg("Closing connection as there are too many opened connections for client")
			err := conn.Close()
			if err != nil {
				s.logger.Error().Err(err).Msg("Error while closing connection.")
			}
			continue
		}
		if err != nil {
			s.logger.Debug().Err(err).Msg("Error while accepting connection.")
		}

		//TODO: handle connections closing in background
		s.logger.Info().Msg("new connection opened")
		err = s.connPool.Append(conn)
		if err != nil {
			s.logger.Err(err).Msg("new connection opened")
		}
	}
}
