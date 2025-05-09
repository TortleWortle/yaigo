package inertia

import (
	"context"
	"errors"
	"net/http"

	"github.com/tortlewortle/yaigo/internal/errflash"
	"github.com/tortlewortle/yaigo/pkg/yaigo"
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
			r = wrapRequest(w, r, server, o)
			next.ServeHTTP(w, r)
		})
	}
}

func wrapRequest(w http.ResponseWriter, r *http.Request, server *yaigo.Server, opts *MiddlewareOpts) *http.Request {
	inertiaReq := yaigo.NewRequest()
	inertiaReq.EncryptHistory(opts.EncryptHistory)

	errs := errflash.GetErrors(w, r)
	inertiaReq.SetProp("errors", errs)

	ctx := r.Context()
	ctx = context.WithValue(ctx, serverKey, server)
	ctx = context.WithValue(ctx, requestKey, inertiaReq)
	return r.WithContext(ctx)
}

func getServer(r *http.Request) (*yaigo.Server, error) {
	rawVal := r.Context().Value(serverKey)
	if rawVal == nil {
		return nil, errors.New("*yaigo.Server not set in context")
	}

	val, ok := rawVal.(*yaigo.Server)
	if !ok {
		return nil, errors.New("*yaigo.Server provided but could not be cast")
	}

	return val, nil
}

func getRequest(r *http.Request) (*yaigo.Request, error) {
	rawVal := r.Context().Value(requestKey)
	if rawVal == nil {
		return nil, errors.New("*yaigo.Request not set in context")
	}

	val, ok := rawVal.(*yaigo.Request)
	if !ok {
		return nil, errors.New("*yaigo.Request provided but could not be cast")
	}

	return val, nil
}
