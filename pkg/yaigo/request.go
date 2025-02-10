package yaigo

import (
	"errors"
	"github.com/tortlewortle/go-inertia/internal/page"
	"github.com/tortlewortle/go-inertia/internal/props"
	"net/http"
)

func NewRequest() *Request {
	return &Request{
		propBag:  props.NewBag(),
		status:   http.StatusOK,
		pageData: page.New(),
	}
}

type Request struct {
	propBag  *props.Bag
	status   int
	pageData *page.InertiaPage
}

// EncryptHistory enables or disables page history encryption inside inertiajs
func (req *Request) EncryptHistory(encrypt bool) {
	req.pageData.EncryptHistory = encrypt
}

// ClearHistory tells inertiajs to roll the cache encryption key.
// This can be used to protect any sensitive information from being accessed after logout by using the back button.
func (req *Request) ClearHistory() {
	req.pageData.ClearHistory = true
}

// SetStatus of the http response
func (req *Request) SetStatus(status int) {
	req.status = status
}

func (req *Request) SetProp(key string, value any) error {
	switch value.(type) {
	case *props.LazyProp:
		p, ok := value.(*props.LazyProp)
		if ok {
			if p.IsDeferred() {
				return errors.New("deferred props can only be used on the page render func")
			}
		}
		return errors.New("could not cast LazyProp")
	}
	bag := req.propBag

	return bag.Set(key, value)
}
