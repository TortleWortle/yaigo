package inertia

import (
	"net/http"
)

func Render(w http.ResponseWriter, r *http.Request, page string, pageProps Props) error {
	server, err := getServer(r)
	if err != nil {
		return err
	}
	req, err := getRequest(r)
	if err != nil {
		return err
	}
	return server.Render(req, w, r, page, pageProps)
}
