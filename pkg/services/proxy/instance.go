package proxy

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"go-server/pkg/services/forward_connection"
	"go-server/pkg/services/origin"
	"net"
	"strings"
	"time"
)

type TcpProxyInstance struct {
	Port int
	ID   string

	conf *Config

	logger   zerolog.Logger
	connPool *forward_connection.ForwardConnectionsPool

	origin *origin.Meta

	close      chan struct{}
	lastActive time.Time

	listener net.Listener

	onClose chan struct{}
}

func NewTcpProxyInstance(logger zerolog.Logger, port int, c *Config, id string, origin *origin.Meta) *TcpProxyInstance {

	tp := &TcpProxyInstance{
		Port:       port,
		ID:         id,
		conf:       c,
		logger:     logger,
		origin:     origin,
		connPool:   forward_connection.NewForwardConnectionsPool(),
		lastActive: time.Now(),
		close:      make(chan struct{}, 1),
	}

	go func() {
		err := tp.listen()
		if err != nil {
			tp.logger.Err(err).Msg("failed to listen")
		}
	}()

	go func() {
		for {
			<-time.After(30 * time.Minute)
			timeoutDelay := time.Now().Add(time.Duration(-1*tp.conf.InactiveHoursTimeout) * time.Hour)
			if tp.lastActive.Before(timeoutDelay) {
				tp.close <- struct{}{}
				return
			}
		}
	}()

	go func() {
		attempt := 0
		for {
			<-time.After(time.Duration(tp.conf.NoActiveSocketsMinutesTimeout) * time.Minute)

			if tp.connPool.Size() == 0 {
				attempt++
			}

			if attempt >= tp.conf.NoActiveSocketsChecks {
				tp.close <- struct{}{}
				return
			}
		}
	}()

	go func() {
		<-tp.close
		err := tp.listener.Close()
		if err != nil {
			tp.logger.Err(err).Int("port", tp.Port).Msg("failed to close listener")
		}

		tp.connPool.Close()

		tp.onClose <- struct{}{}
	}()

	return tp
}

func (s *TcpProxyInstance) OnClose() <-chan struct{} {
	return s.onClose
}

func (s *TcpProxyInstance) MaxConns() int {
	return s.conf.MaxConnsPerClient
}

func (s *TcpProxyInstance) URL() string {
	domain := s.origin.Host()
	if s.conf.BaseDomain != "" {
		domain = s.conf.BaseDomain
	}

	return s.origin.Scheme() + "://" + s.ID + "." + domain
}

func (s *TcpProxyInstance) Proxy(data []byte) (error, []byte) {
	s.updateActive()

	c := s.connPool.Get()
	attempt := 0
	for c == nil {
		if attempt >= 5 {
			return errors.New("failed to get proxy connection, retries exceeded"), nil
		}

		time.Sleep(200 * time.Millisecond)

		c = s.connPool.Get()
		attempt++
	}

	defer func() {
		c.Release()
		c.Close()
	}()

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

func (s *TcpProxyInstance) updateActive() {
	s.lastActive = time.Now()
}

func (s *TcpProxyInstance) listen() error {
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		s.logger.Fatal().Int("port", s.Port).Err(err).Msg("Unable to create listener.")
		return err
	}
	defer s.listener.Close()
	s.logger.Info().Str("bind", s.listener.Addr().String()).Str("protocol", "tcp").Msg("Listening...")

	for {
		conn, err := s.listener.Accept()
		s.updateActive()

		ra := conn.RemoteAddr().String()
		remoteIP := net.ParseIP(strings.Split(ra, ":")[0])
		if s.origin.IP() != nil && !s.origin.IP().Equal(remoteIP) {
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

		s.logger.Info().Msg("new connection opened")
		err = s.connPool.Append(conn)
		if err != nil {
			s.logger.Err(err).Msg("new connection opened")
		}
	}
}

func (s TcpProxyInstance) GetAddr() string {
	return s.listener.Addr().String()
}

func (s TcpProxyInstance) Connections() int {
	return s.connPool.Size()
}
