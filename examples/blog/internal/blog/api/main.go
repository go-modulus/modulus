package api

import (
	"encoding/json"
	mHttp "github.com/go-modulus/modulus/http"
	"net/http"
)

type Main struct {
}

func NewMain() *Main {
	return &Main{}
}

func NewMainRoute(handler *Main) mHttp.RouteProvider {
	return mHttp.ProvideInputRoute(
		"GET",
		"/",
		handler.Handle,
	)
}

type MainInput struct {
	// Use https://github.com/ggicci/httpin to define the input parameters
	// Example:
	// Name string    `in:"query=name"`

}

type MainResponse struct {
	Ok bool `json:"ok"`
}

func (h *Main) Handle(rw http.ResponseWriter, r mHttp.RequestWithInput[MainInput]) error {

	return json.NewEncoder(rw).Encode(MainResponse{Ok: true})
}
