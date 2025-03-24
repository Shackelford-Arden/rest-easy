package api

import "net/http"

type CustomHandler struct {
	Handler http.Handler

	resp *http.ResponseWriter
}

func (h *CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Handler.ServeHTTP(w, r)
}

func (h *CustomHandler) respond() error {

	return nil
}
