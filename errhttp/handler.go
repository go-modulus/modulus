package errhttp

import "net/http"

type Handler func(w http.ResponseWriter, req *http.Request) error
