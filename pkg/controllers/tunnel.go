package controllers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/rs/zerolog"
	"go-server/pkg/services"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
)

type TunnelController struct {
	logger zerolog.Logger

	proxyManager *services.TcpProxyManager
}

func NewTunnelController(logger zerolog.Logger, proxyManager *services.TcpProxyManager) *TunnelController {
	return &TunnelController{logger: logger, proxyManager: proxyManager}
}

func (t *TunnelController) CreateConnection(w http.ResponseWriter, r *http.Request) {
	tq := newTunnelRequest(r)

	body, err := io.ReadAll(r.Body)
	if len(body) != 0 {
		d := json.NewDecoder(bytes.NewReader(body))
		defer r.Body.Close()
		err = d.Decode(&tq)
		if err != nil {
			t.logger.Error().Err(err).Msgf("failed to parse input json")
			w.WriteHeader(500)
			return
		}
	}

	t.createTunnelResponse(w, tq)
	return
}

func (t *TunnelController) TryProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" && r.URL.Query().Has("new") {
		tq := newTunnelRequest(r)
		tq.Name = r.URL.Query().Get("new")

		t.createTunnelResponse(w, tq)
		return
	}

	t.Proxy(w, r)
}

func (t *TunnelController) Proxy(w http.ResponseWriter, r *http.Request) {
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

	err, resp := conn.Proxy(input)
	if err != nil && !err.(*net.OpError).Timeout() {
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

func (t *TunnelController) replicateBody(w http.ResponseWriter, err error, parsedResp *http.Response) {
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

func (t *TunnelController) replicateHeaders(w http.ResponseWriter, parsedResp *http.Response) {
	w.WriteHeader(parsedResp.StatusCode)

	for s := range parsedResp.Header {
		w.Header().Set(s, parsedResp.Header.Get(s))
	}
}

func (t *TunnelController) clearHeaders(w http.ResponseWriter) {
	for k := range w.Header() {
		w.Header().Del(k)
	}
}

func (t *TunnelController) createTunnelResponse(w http.ResponseWriter, tq tunnelRequest) {
	c := t.createTunnel(tq)

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
}

func (t *TunnelController) createTunnel(tq tunnelRequest) *services.TcpProxyInstance {
	if tq.Name == "" {
		tq.Name = services.GenerateTunnelName()
	}

	c := t.proxyManager.New(tq.Name, &services.OriginMeta{
		Url: tq.originURL,
		Ip:  tq.originalIP,
	})
	return c
}
