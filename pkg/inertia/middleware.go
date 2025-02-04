package inertia

import (
	"context"
	"errors"
	"net/http"
)

type contextKey int

const (
	serverKey contextKey = iota
	requestKey
	statusKey
)

func (s *Server) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inertiaReq := s.requestPool.Get().(*request)

		ctx := r.Context()
		ctx = context.WithValue(ctx, serverKey, s)
		ctx = context.WithValue(ctx, requestKey, inertiaReq)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

		inertiaReq.Reset()
		s.requestPool.Put(inertiaReq)
	})
}

func SetStatus(r *http.Request, status int) error {
	return setStatusCtx(r.Context(), status)
}

func setStatusCtx(ctx context.Context, status int) error {
	req, err := getRequestCtx(ctx)
	if err != nil {
		return err
	}
	req.status = status
	return nil
}

func getServer(r *http.Request) (*Server, error) {
	rawVal := r.Context().Value(serverKey)
	if rawVal == nil {
		return nil, errors.New("server not set in context")
	}

	val, ok := rawVal.(*Server)
	if !ok {
		return nil, errors.New("server provided but could not be cast")
	}

	return val, nil
}

func getRequest(r *http.Request) (*request, error) {
	return getRequestCtx(r.Context())
}

func getRequestCtx(ctx context.Context) (*request, error) {
	rawVal := ctx.Value(requestKey)
	if rawVal == nil {
		return nil, errors.New("request not set in context")
	}

	val, ok := rawVal.(*request)
	if !ok {
		return nil, errors.New("request provided but could not be cast")
	}

	return val, nil
}
