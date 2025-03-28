package inertia

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tortlewortle/yaigo/internal/errflash"
)

var ErrDirtyRender = errors.New("ResponseWriter has already been written to")

// DefaultErrHandler is called when a HandlerFunc returns an error (typically prefer c.Error() over returning an error)
var DefaultErrHandler = func(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

type HandlerFunc func(c *Ctx, request *http.Request) error

// ServeHTTP calls Handler(f)(w, r).
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Handler(f)(w, r)
}

func Handler(handlerFunc HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := &Ctx{
			request: r,
			writer:  w,
		}
		err := handlerFunc(c, r)
		if err != nil {
			DefaultErrHandler(w, r, err)
			return
		}
	}
}

type Ctx struct {
	request *http.Request
	writer  http.ResponseWriter

	dirty bool // keep track on wether this request has been manually written to
}

// compat for http.ResponseWriter
var _ http.ResponseWriter = &Ctx{}

// Header implements http.ResponseWriter.
func (c *Ctx) Header() http.Header {
	return c.writer.Header()
}

// Write implements http.ResponseWriter.
func (c *Ctx) Write(b []byte) (int, error) {
	c.dirty = true

	return c.writer.Write(b)
}

// WriteHeader implements http.ResponseWriter. DO NOT USE TO SET THE STATUS CODE
func (c *Ctx) WriteHeader(statusCode int) {
	c.dirty = true
	c.writer.WriteHeader(statusCode)
}

// Render queues the component for rendering after the handler finishes
func (c *Ctx) Render(page string, props Props) error {
	if c.dirty {
		return ErrDirtyRender
	}
	server, err := getServer(c.request)
	if err != nil {
		return err
	}
	req, err := getRequest(c.request)
	if err != nil {
		return err
	}
	return server.RenderRequest(req, c.writer, c.request, page, props)
}

func (c *Ctx) RenderWithStatus(page string, status int, props Props) error {
	c.Status(status)
	return c.Render(page, props)
}

func (c *Ctx) ErrorWithProps(cause error, status int, pageProps Props) error {
	c.Status(status)
	p := Props{
		"status": status,
	}

	for k, v := range pageProps {
		p[k] = v
	}

	server, err := getServer(c.request)
	if err != nil {
		return err
	}

	if server.IsDevMode() {
		// TODO: pretty stack trace? :)
		// TODO: maybe panic recoverer as well
		fmt.Fprintf(c.writer, "error: %v", cause)
		return nil
	}
	return c.Render("Error", p)
}

// Error renders the Error component and will print out the cause in devmode
func (c *Ctx) Error(cause error, status int) error {
	return c.ErrorWithProps(cause, status, nil)
}

func (c *Ctx) Status(status int) {
	req, err := getRequest(c.request)
	if err != nil {
		panic("Status: could not get *yaigo.Request from *http.Request context, is it wrapped in the middleware?")
	}
	req.SetStatus(status)
}

// ClearHistory tells inertiajs to roll the cache encryption key.
// This can be used to protect any sensitive information from being accessed after logout by using the back button.
func (c *Ctx) ClearHistory() error {
	req, err := getRequest(c.request)
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
	http.SetCookie(c.writer, cookie)
}

func (c *Ctx) Redirect(url string) error {
	http.Redirect(c.writer, c.request, url, http.StatusSeeOther)
	return nil
}

// Redirect instructs inertia to redirect properly using http.StatusSeeOther
func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// Location redirects to external urls
func Location(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (c *Ctx) Location(url string) error {
	c.writer.Header().Set(headerLocation, url)
	c.writer.WriteHeader(http.StatusConflict)
	return nil
}

type FlashErrors = errflash.FlashErrors

func (c *Ctx) Back(errs FlashErrors) error {
	if errs != nil {
		errflash.FlashError(c.writer, c.request, errs)
	}
	http.Redirect(c.writer, c.request, c.request.Referer(), http.StatusSeeOther)
	return nil
}
