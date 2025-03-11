package yaigo

import (
	"errors"
	"net/http"

	"github.com/tortlewortle/yaigo/internal/page"
	"github.com/tortlewortle/yaigo/internal/props"
)

func NewResponse() *Response {
	return &Response{
		propBag:  props.NewBag(),
		status:   http.StatusOK,
		pageData: page.New(),
	}
}

type Response struct {
	propBag  *props.Bag
	status   int
	pageData *page.InertiaPage
}

// EncryptHistory enables or disables page history encryption inside inertiajs
func (req *Response) EncryptHistory(encrypt bool) {
	req.pageData.EncryptHistory = encrypt
}

// ClearHistory tells inertiajs to roll the cache encryption key.
// This can be used to protect any sensitive information from being accessed after logout by using the back button.
func (req *Response) ClearHistory() {
	req.pageData.ClearHistory = true
}

// SetStatus of the http response
func (req *Response) SetStatus(status int) {
	req.status = status
}

func (req *Response) SetProp(key string, value any) error {
	switch p := value.(type) {
	case *props.LazyProp:
		if p.IsDeferred() {
			return errors.New("deferred props can only be used on the page render func")
		}
	}
	bag := req.propBag

	bag.Set(key, value)

	return nil
}
