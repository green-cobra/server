package tunnel

type tunnelResponse struct {
	Name             string `json:"id,omitempty"`
	ProxyEndpointUrl string `json:"proxy_endpoint_url,omitempty"`
	ClientUrl        string `json:"client_url,omitempty"`
	MaxConns         int    `json:"max_conn_count,omitempty"`
}
