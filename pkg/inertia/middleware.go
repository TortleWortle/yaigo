package inertia

import (
	"context"
	"errors"
	"github.com/tortlewortle/yaigo/pkg/yaigo"
	"net/http"
)

type contextKey int

const (
	serverKey contextKey = iota
	requestKey
)

type MiddlewareOpts struct {
	EncryptHistory bool
}

func WithHistoryEncryption(opt *MiddlewareOpts) {
	opt.EncryptHistory = true
}

func NewMiddleware(server *yaigo.Server, opts ...func(*MiddlewareOpts)) func(next http.Handler) http.Handler {
	o := &MiddlewareOpts{}
	for _, fn := range opts {
		fn(o)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = wrapRequest(r, server, o)
			next.ServeHTTP(w, r)
		})
	}
}

func wrapRequest(r *http.Request, server *yaigo.Server, opts *MiddlewareOpts) *http.Request {
	inertiaReq := yaigo.NewResponse()
	inertiaReq.EncryptHistory(opts.EncryptHistory)

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

func getResponse(r *http.Request) (*yaigo.Response, error) {
	rawVal := r.Context().Value(requestKey)
	if rawVal == nil {
		return nil, errors.New("request not set in context")
	}

	val, ok := rawVal.(*yaigo.Response)
	if !ok {
		return nil, errors.New("request provided but could not be cast")
	}

	return val, nil
}
