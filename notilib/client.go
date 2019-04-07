package notilib

import "net/http"

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type dispatcher interface {
	dispatch(*http.Request) (*http.Response, error)
}

type clientHandler struct {
	client httpClient
}

func newClientHandler(c httpClient) dispatcher {
	return &clientHandler{client: c}
}

// dispatch the http request to the client
func (hdlr *clientHandler) dispatch(req *http.Request) (*http.Response, error) {
	return hdlr.client.Do(req)
}
