package frontend

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux) {
	fs := http.FileServer(http.Dir("frontend-stuff/static/"))
	mux.Handle("/", fs)
}
