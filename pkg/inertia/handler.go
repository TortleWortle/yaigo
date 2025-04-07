package inertia

import (
	"errors"
	"fmt"
	"github.com/tortlewortle/yaigo/pkg/yaigo"
	"net/http"
	"runtime/debug"

	"github.com/tortlewortle/yaigo/internal/errflash"
)

var ErrDirtyRender = errors.New("ResponseWriter has already been written to")

var DefaultErrorComponent = "Error"

// DefaultErrHandler is called when a HandlerFunc returns an error (typically prefer c.Error() over returning an error) or when a handler panics.
var DefaultErrHandler = func(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

type HandlerFunc func(c *Ctx, request *http.Request) error

// ServeHTTP calls Handler(f)(w, r).
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Handler(f)(w, r)
}

// Handler takes in an inertia.HandlerFunc and returns a http.HandlerFunc.
//
// Panics if the inertia.NewMiddleware() middleware is not used to inject a *yaigo.Server and *yaigo.Request
func Handler(handlerFunc HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server, err := getServer(r)
		if err != nil {
			panic(fmt.Sprintf("could not get *yaigo.Server for inertia.Handler: %v", err))
		}

		request, err := getRequest(r)
		if err != nil {
			panic(fmt.Sprintf("could not get *yaigo.Request for inertia.Handler: %v", err))
		}

		defer func() {
			if rec := recover(); rec != nil {
				if server.IsDevMode() {
					fmt.Fprintf(w, "<h1>Recovered from panic: %v</h1>\n<pre>%s</pre>", rec, debug.Stack())
				} else {
					DefaultErrHandler(w, r, err)
				}
			}
		}()

		c := &Ctx{
			httpRequest:    r,
			responseWriter: w,

			yaigoServer:  server,
			yaigoRequest: request,
		}
		err = handlerFunc(c, r)
		if err != nil {
			if server.IsDevMode() {
				fmt.Fprintf(w, "<h1>inertia.Handler() returned err: %v</h1>\n<pre>%s</pre>", err, debug.Stack())
			} else {
				DefaultErrHandler(w, r, err)
			}
			return
		}
	}
}

type Ctx struct {
	httpRequest    *http.Request
	responseWriter http.ResponseWriter

	yaigoServer  *yaigo.Server
	yaigoRequest *yaigo.Request

	dirty bool // keep track on wether this httpRequest has been manually written to
}

// compat for http.ResponseWriter
var _ http.ResponseWriter = &Ctx{}

// Header implements http.ResponseWriter.
func (c *Ctx) Header() http.Header {
	return c.responseWriter.Header()
}

// Write implements http.ResponseWriter.
func (c *Ctx) Write(b []byte) (int, error) {
	c.dirty = true

	return c.responseWriter.Write(b)
}

// WriteHeader implements http.ResponseWriter. DO NOT USE TO SET THE STATUS CODE
func (c *Ctx) WriteHeader(statusCode int) {
	c.dirty = true
	c.responseWriter.WriteHeader(statusCode)
}

// Render queues the component for rendering after the handler finishes
func (c *Ctx) Render(page string, props Props) error {
	if c.dirty {
		return ErrDirtyRender
	}
	return c.yaigoServer.RenderRequest(c.yaigoRequest, c.responseWriter, c.httpRequest, page, props)
}

func (c *Ctx) RenderWithStatus(page string, status int, props Props) error {
	c.Status(status)
	return c.Render(page, props)
}

// Error renders the DefaultErrorComponent with the status as prop, sheds any existing props prior to call
func (c *Ctx) Error(cause error, status int, pageProps Props) error {
	c.Status(status)
	props := Props{
		"status": status,
	}

	for k, v := range pageProps {
		props[k] = v
	}

	if c.yaigoServer.IsDevMode() {
		fmt.Fprintf(c.responseWriter, "error: %v", cause)
		return nil
	}

	return c.yaigoServer.Render(c.responseWriter, c.httpRequest, DefaultErrorComponent, props)
}

func (c *Ctx) Status(status int) {
	req, err := getRequest(c.httpRequest)
	if err != nil {
		panic("Status: could not get *yaigo.Request from *http.Request context, is it wrapped in the middleware?")
	}
	req.SetStatus(status)
}

// ClearHistory tells inertiajs to roll the cache encryption key.
// This can be used to protect any sensitive information from being accessed after logout by using the back button.
func (c *Ctx) ClearHistory() error {
	req, err := getRequest(c.httpRequest)
	if err != nil {
		return err
	}
	req.ClearHistory()
	return nil
}

const (
	headerLocation = "X-Inertia-Location"
)

func (c *Ctx) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.responseWriter, cookie)
}

func (c *Ctx) Redirect(url string) error {
	http.Redirect(c.responseWriter, c.httpRequest, url, http.StatusSeeOther)
	return nil
}

// Redirect instructs inertia to redirect properly using http.StatusSeeOther
func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// Location redirects to external urls
func Location(w http.ResponseWriter, r *http.Request, url string) {
	w.Header().Set(headerLocation, url)
	w.WriteHeader(http.StatusConflict)
}

func (c *Ctx) Location(url string) error {
	c.responseWriter.Header().Set(headerLocation, url)
	c.responseWriter.WriteHeader(http.StatusConflict)
	return nil
}

type FlashErrors = errflash.FlashErrors

func (c *Ctx) Back(errs FlashErrors) error {
	if errs != nil {
		errflash.FlashError(c.responseWriter, c.httpRequest, errs)
	}
	http.Redirect(c.responseWriter, c.httpRequest, c.httpRequest.Referer(), http.StatusSeeOther)
	return nil
}
