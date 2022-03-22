package tunnel

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"go-server/pkg/services"
	"go-server/pkg/services/origin"
	"go-server/pkg/services/proxy"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
)

type Controller struct {
	logger zerolog.Logger

	proxyManager *proxy.TcpProxyManager
}

func NewTunnelController(logger zerolog.Logger, proxyManager *proxy.TcpProxyManager) *Controller {
	return &Controller{logger: logger, proxyManager: proxyManager}
}

func (t *Controller) CreateConnection(w http.ResponseWriter, r *http.Request) {
	tq := newTunnelRequest(r)

	body, err := io.ReadAll(r.Body)
	if len(body) != 0 {
		d := json.NewDecoder(bytes.NewReader(body))
		defer r.Body.Close()
		err = d.Decode(&tq)
		if err != nil {
			t.logger.Error().Err(err).Msgf("create connection: failed to parse input json")
			w.WriteHeader(500)
			return
		}
	}

	t.createTunnelResponse(w, tq)
	return
}

func (t Controller) TryProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" && r.URL.Query().Has("new") {
		tq := newTunnelRequest(r)
		tq.Name = r.URL.Query().Get("new")

		t.createTunnelResponse(w, tq)
		return
	}

	t.Proxy(w, r)
}

func (t Controller) Proxy(w http.ResponseWriter, r *http.Request) {
	tunnelId := services.GetTunnelNameFromHost(r.Host)

	ok := t.proxyManager.Exists(tunnelId)
	if !ok {
		w.Write([]byte("not found"))
		w.WriteHeader(404)
		return
	}

	conn := t.proxyManager.Get(tunnelId)
	if conn == nil {
		w.Write([]byte("no tunnel connection available"))
		w.WriteHeader(504)
		return
	}

	input, err := httputil.DumpRequest(r, true)
	if err != nil {
		t.logger.Error().Err(err).Msg("failed to read client data")
		w.WriteHeader(500)
		return
	}

	// err might be non nil because of response read timeout
	// Timeout will be triggered because of deadline exceed
	// timeout exceed can be caused by 2 reasons:
	// 1. Real client timeout
	// 2. Read data timeout caused by data receive finish
	// Here it ignores err in case response size is non 0
	err, resp := conn.Proxy(input)
	if err != nil && len(resp) == 0 {
		t.logger.Error().Err(err).Msg("failed to proxy data")
		w.WriteHeader(500)
		return
	}

	parsedResp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(resp)), r)
	if err != nil {
		netErr, ok := err.(*net.OpError)
		if !ok || (ok && !netErr.Timeout()) {
			t.logger.Error().Err(err).Msg("failed to parse req")
			w.WriteHeader(500)
			return
		}
	}

	t.clearHeaders(w)
	t.replicateHeaders(w, parsedResp)

	t.replicateBody(w, err, parsedResp)

}

func (t Controller) replicateBody(w http.ResponseWriter, err error, parsedResp *http.Response) {
	body, err := io.ReadAll(parsedResp.Body)
	if err != nil {
		t.logger.Error().Err(err).Msg("failed to parse req body")
	}

	_, err = w.Write(body)
	if err != nil {
		t.logger.Error().Err(err).Msg("failed to write data")
		w.WriteHeader(500)
		return
	}
}

func (Controller) replicateHeaders(w http.ResponseWriter, parsedResp *http.Response) {
	w.WriteHeader(parsedResp.StatusCode)

	for s := range parsedResp.Header {
		w.Header().Set(s, parsedResp.Header.Get(s))
	}
}

func (Controller) clearHeaders(w http.ResponseWriter) {
	for k := range w.Header() {
		w.Header().Del(k)
	}
}

func (t Controller) createTunnelResponse(w http.ResponseWriter, tq tunnelRequest) {
	c := t.createTunnel(tq)
	if c == nil {
		t.logger.Error().Msgf("failed to create proxy for request: %+v", tq)
		w.WriteHeader(500)
		return
	}

	tr := tunnelResponse{
		Name:     c.ID,
		Port:     c.Port,
		Url:      c.URL(),
		MaxConns: c.MaxConns(),
	}

	enc := json.NewEncoder(w)
	err := enc.Encode(tr)

	if err != nil {
		t.logger.Error().Err(err).Msg("failed to encode response")
		w.WriteHeader(500)
		return
	}

	t.logger.Info().Str("name", c.ID).Int("port", c.Port).Str("url", c.URL()).Msg("opened new tunnel")
}

func (t Controller) createTunnel(tq tunnelRequest) *proxy.TcpProxyInstance {
	if tq.Name == "" {
		tq.Name = services.GenerateTunnelName()
	}

	for t.proxyManager.Exists(tq.Name) {
		tq.Name = services.GenerateTunnelName()
	}

	c := t.proxyManager.New(tq.Name, origin.NewMeta(tq.originURL, tq.originalIP))
	return c
}

func (t *Controller) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	tunnelId := chi.URLParam(r, "id")

	ok := t.proxyManager.Exists(tunnelId)
	if !ok {
		w.Write([]byte("not found"))
		w.WriteHeader(404)
		return
	}

	connection := t.proxyManager.Get(tunnelId)

	creatorIP := connection.GetCreatorIP()
	requestIP := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
	if !creatorIP.Equal(requestIP) {
		w.WriteHeader(404)
		return
	}

	connection.RequestClose()
	w.WriteHeader(200)
	w.Write([]byte("{}"))
}
