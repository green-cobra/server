package controllers

type tunnelResponse struct {
	Name     string `json:"id,omitempty"`
	Port     int    `json:"port,omitempty"`
	Url      string `json:"url,omitempty"`
	MaxConns int    `json:"max_conn_count,omitempty"`
}
