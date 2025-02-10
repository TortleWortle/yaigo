package inertia

import (
	"context"
	"errors"
	"github.com/tortlewortle/go-inertia/pkg/yaigo"
	"net/http"
)

type contextKey int

const (
	serverKey contextKey = iota
	requestKey
)

func NewMiddleware(server *yaigo.Server) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = wrapRequest(r, server)
			next.ServeHTTP(w, r)
		})
	}
}

func wrapRequest(r *http.Request, server *yaigo.Server) *http.Request {
	inertiaReq := yaigo.NewRequest()

	ctx := r.Context()
	ctx = context.WithValue(ctx, serverKey, server)
	ctx = context.WithValue(ctx, requestKey, inertiaReq)
	return r.WithContext(ctx)
}

func getServer(r *http.Request) (*yaigo.Server, error) {
	rawVal := r.Context().Value(serverKey)
	if rawVal == nil {
		return nil, errors.New("server not set in context")
	}

	val, ok := rawVal.(*yaigo.Server)
	if !ok {
		return nil, errors.New("server provided but could not be cast")
	}

	return val, nil
}

func getRequest(r *http.Request) (*yaigo.Request, error) {
	rawVal := r.Context().Value(requestKey)
	if rawVal == nil {
		return nil, errors.New("request not set in context")
	}

	val, ok := rawVal.(*yaigo.Request)
	if !ok {
		return nil, errors.New("request provided but could not be cast")
	}

	return val, nil
}
