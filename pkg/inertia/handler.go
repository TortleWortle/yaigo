package inertia

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tortlewortle/yaigo/internal/errflash"
)

var ErrDirtyRender = errors.New("ResponseWriter has already been written to")

var DefaultErrHandler = func(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("handler error", slog.String("err", err.Error()))
	_, _ = fmt.Fprintf(w, "handler error: %v", err)
}

type HandlerFunc func(c *Ctx, request *http.Request) error

func Handler(handlerFunc HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := &Ctx{
			request: r,
			writer:  w,
		}
		err := handlerFunc(c, r)
		if err != nil {
			DefaultErrHandler(w, r, err)
			return
		}
	})
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
	req, err := getResponse(c.request)
	if err != nil {
		return err
	}
	return server.Render(req, c.writer, c.request, page, props)
}

// Error renders the Error component and will print out the cause in devmode (not implemented yet)
func (c *Ctx) Error(cause error, status int) error {
	// todo: check devmode, render pretty component instead of raw error if not devmode
	fmt.Fprintf(c.writer, "error: %v", cause)

	return nil
}

func (c *Ctx) Status(status int) error {
	req, err := getResponse(c.request)
	if err != nil {
		return err
	}
	req.SetStatus(status)
	return nil
}

// ClearHistory tells inertiajs to roll the cache encryption key.
// This can be used to protect any sensitive information from being accessed after logout by using the back button.
func (c *Ctx) ClearHistory() error {
	req, err := getResponse(c.request)
	if err != nil {
		return err
	}
	req.ClearHistory()
	return nil
}

const (
	headerLocation = "X-Inertia-Location"
)

func (c *Ctx) Redirect(url string) error {
	http.Redirect(c.writer, c.request, url, http.StatusSeeOther)
	return nil
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
