package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HelloHandler struct{}

func NewHelloHandler() *HelloHandler {
	return &HelloHandler{}
}

// Index demo html page.
// @Summary Demo HTML page
// @Tags System
// @Produce html
// @Success 200 {string} string "<div>Hello World</div>"
// @Router / [get]
func (h *HelloHandler) Index(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<div>Hello World</div>")
}

// Hello demo greeting.
// @Summary Demo greeting
// @Tags System
// @Produce json
// @Success 200 {object} apidocs.Hello
// @Router /hello [get]
func (h *HelloHandler) Hello(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Hello from Go!",
	})
}
