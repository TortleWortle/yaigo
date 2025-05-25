package yaigo

import (
	"github.com/tortlewortle/yaigo/internal/errflash"
	"github.com/tortlewortle/yaigo/internal/page"
	"github.com/tortlewortle/yaigo/internal/props"
	"net/http"
	"sync"
)

type MiddlewareOpts struct {
	EncryptHistory bool
}

func WithHistoryEncryption(encrypt bool) func(*MiddlewareOpts) {
	return func(opt *MiddlewareOpts) {
		opt.EncryptHistory = encrypt
	}
}

// Middleware provides the context with a config, RequestURI and a prop bag, also handles version conflicts
func Middleware(config *Config, opts ...func(*MiddlewareOpts)) func(http.Handler) http.Handler {
	o := &MiddlewareOpts{}
	for _, fn := range opts {
		fn(o)
	}
	bagPool := sync.Pool{
		New: func() interface{} {
			return props.NewBag()
		},
	}
	infoPool := sync.Pool{
		New: func() interface{} {
			return &RequestInfo{}
		},
	}
	inertiaPagePool := sync.Pool{
		New: func() interface{} {
			return page.New()
		},
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := infoPool.Get().(*RequestInfo)
			info.Fill(r)

			if info.RedirectIfVersionConflict(w, config.manifestVersion) {
				errflash.Reflash(w, r)
				w.Header().Set(HeaderLocation, r.URL.String())
				w.WriteHeader(http.StatusConflict)
				return
			}
			bag := bagPool.Get().(*props.Bag)
			pageData := inertiaPagePool.Get().(*page.InertiaPage)

			pageData.Version = config.manifestVersion
			pageData.Url = r.RequestURI
			pageData.EncryptHistory = o.EncryptHistory

			errs := errflash.GetErrors(w, r)
			bag.Set("errors", errs)

			ctx := WithConfig(r.Context(), config)
			ctx = WithRequestInfo(ctx, info)
			ctx = WithPropBag(ctx, bag)
			ctx = WithInertiaPage(ctx, pageData)
			next.ServeHTTP(w, r.WithContext(ctx))

			// empty and return values to pool
			info.Empty()
			infoPool.Put(info)

			bag.Empty()
			bagPool.Put(bag)

			pageData.Reset()
			inertiaPagePool.Put(pageData)
		})
	}
}
