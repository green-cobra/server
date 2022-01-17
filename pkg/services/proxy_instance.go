package services

import (
	"fmt"
	"github.com/rs/zerolog"
	"math/rand"
	"net"
	"net/url"
)

type TcpProxyInstance struct {
	Port int
	ID   string

	conf *ProxyConfig

	logger   zerolog.Logger
	connPool *forwardConnectionsPool

	originUrl *url.URL
}

func NewTcpProxyInstance(logger zerolog.Logger, c *ProxyConfig, id string, originUrl *url.URL) *TcpProxyInstance {
	port := c.MinPort + rand.Intn(c.MaxPort-c.MinPort)

	return &TcpProxyInstance{
		Port:      port,
		ID:        id,
		conf:      c,
		logger:    logger,
		originUrl: originUrl,
		connPool:  newForwardConnectionsPool(),
	}
}

func (s *TcpProxyInstance) URL() string {
	domain := s.originUrl.Host
	if s.conf.BaseDomain != "" {
		domain = s.conf.BaseDomain
	}

	return s.originUrl.Scheme + "://" + s.ID + "." + domain
}

func (s *TcpProxyInstance) Proxy(data []byte) (error, []byte) {
	c := s.connPool.Get()
	if c == nil {
		// Todo: handle logic of waiting for connection to be available
		return nil, nil
	}

	defer c.Release()

	//TODO: handle errors
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
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		s.logger.Fatal().Int("port", s.Port).Err(err).Msg("Unable to create listener.")
		return err
	}
	defer listener.Close()
	s.logger.Info().Str("bind", listener.Addr().String()).Str("protocol", "tcp").Msg("Listening...")

	for {
		conn, err := listener.Accept()
		if s.connPool.Size() >= 10 {
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

		//TODO: err
		//TODO: handle connections closing in background

		s.logger.Info().Msg("new connection opened")
		err = s.connPool.Append(conn)
		if err != nil {
			s.logger.Err(err).Msg("new connection opened")
		}
	}
}
