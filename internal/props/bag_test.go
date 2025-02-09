package props

import (
	"slices"
	"testing"
)

func TestBag_Except(t *testing.T) {
	b := NewBag()
	if err := b.Set("username", "john"); err != nil {
		t.Error(err)
	}
	if err := b.Set("age", 32); err != nil {
		t.Error(err)
	}
	if err := b.Set("deferred", NewLazyProp(func() (any, error) {
		return true, nil
	}, "default", true, false)); err != nil {
		t.Error(err)
	}
	if err := b.Set("async", NewLazyProp(func() (any, error) {
		return true, nil
	}, "default", false, false)); err != nil {
		t.Error(err)
	}
	b.Except([]string{"async"})
	props, err := b.GetProps()
	if err != nil {
		t.Error(err)
	}

	if _, ok := props["username"]; !ok {
		t.Error("username must be returned")
	}
	if _, ok := props["age"]; !ok {
		t.Error("age must be returned")
	}
	if _, ok := props["deferred"]; ok {
		t.Error("deferred must not be returned")
	}
	if _, ok := props["async"]; ok {
		t.Error("async must not be returned")
	}
}

func TestBag_Only(t *testing.T) {
	b := NewBag()
	if err := b.Set("username", "john"); err != nil {
		t.Error(err)
	}
	if err := b.Set("age", 32); err != nil {
		t.Error(err)
	}
	if err := b.Set("deferred", NewLazyProp(func() (any, error) {
		return true, nil
	}, "default", true, false)); err != nil {
		t.Error(err)
	}
	if err := b.Set("async", NewLazyProp(func() (any, error) {
		return true, nil
	}, "default", false, false)); err != nil {
		t.Error(err)
	}
	b.Only([]string{"async", "deferred", "age"})
	props, err := b.GetProps()
	if err != nil {
		t.Error(err)
	}

	if _, ok := props["username"]; ok {
		t.Error("username must not be returned")
	}
	if _, ok := props["age"]; !ok {
		t.Error("age must be returned")
	}
	if _, ok := props["deferred"]; !ok {
		t.Error("deferred must be returned")
	}
	if _, ok := props["async"]; !ok {
		t.Error("async must be returned")
	}
}

func TestBag_OnlyExcept(t *testing.T) {
	b := NewBag()
	if err := b.Set("username", "john"); err != nil {
		t.Error(err)
	}
	if err := b.Set("age", 32); err != nil {
		t.Error(err)
	}
	if err := b.Set("deferred", NewLazyProp(func() (any, error) {
		return true, nil
	}, "default", true, false)); err != nil {
		t.Error(err)
	}
	if err := b.Set("async", NewLazyProp(func() (any, error) {
		return true, nil
	}, "default", false, false)); err != nil {
		t.Error(err)
	}

	b.Except([]string{"age", "deferred"})
	b.Only([]string{"async", "deferred", "age"})
	props, err := b.GetProps()
	if err != nil {
		t.Error(err)
	}

	if _, ok := props["username"]; ok {
		t.Error("username must not be returned")
	}
	if _, ok := props["age"]; ok {
		t.Error("age must not be returned")
	}
	if _, ok := props["deferred"]; ok {
		t.Error("deferred must not be returned")
	}
	if _, ok := props["async"]; !ok {
		t.Error("async must be returned")
	}
}

func TestBag_Checkpoint2(t *testing.T) {
	b := NewBag()
	err := b.Set("username", "john")
	if err != nil {
		t.Error(err)
	}
	b.Checkpoint()
	err = b.Set("username", "doe")
	if err != nil {
		t.Error(err)
	}

	props, err := b.GetProps()
	if err != nil {
		t.Error(err)
	}

	username, ok := props["username"]
	if !ok {
		t.Error("username should be set")
	}

	if username != "doe" {
		t.Errorf("username should be 'doe', got: %s", username)
	}
	b.Checkpoint()

	props, err = b.GetProps()
	if err != nil {
		t.Error(err)
	}

	username, ok = props["username"]
	if !ok {
		t.Error("username should be set")
	}

	if username != "john" {
		t.Errorf("username should be 'john', got: %s", username)
	}
}

func TestBag_Checkpoint(t *testing.T) {
	b := NewBag()

	err := b.Set("username", "john")
	if err != nil {
		t.Error(err)
	}

	b.Checkpoint()
	err = b.Set("age", 32)
	if err != nil {
		t.Error(err)
	}
	err = b.Set("deferSync", NewLazyProp(func() (any, error) {
		return true, nil
	}, "default", true, true))
	if err != nil {
		t.Error(err)
	}
	err = b.Set("defer", NewLazyProp(func() (any, error) {
		return true, nil
	}, "default", true, false))
	if err != nil {
		t.Error(err)
	}

	err = b.Set("sync", NewLazyProp(func() (any, error) {
		return true, nil
	}, "default", false, true))
	if err != nil {
		t.Error(err)
	}

	err = b.Set("async", NewLazyProp(func() (any, error) {
		return true, nil
	}, "default", false, false))
	if err != nil {
		t.Error(err)
	}

	b.Checkpoint()
	props, err := b.GetProps()

	username, ok := props["username"]
	if !ok {
		t.Error("username should still be set")
	}
	if username != "john" {
		t.Error("wrong username")
	}

	_, ok = props["age"]
	if ok {
		t.Error("age should be reset")
	}

	_, ok = props["deferSync"]
	if ok {
		t.Error("deferSync should be reset")
	}

	_, ok = props["defer"]
	if ok {
		t.Error("defer should be reset")
	}

	_, ok = props["sync"]
	if ok {
		t.Error("sync should be reset")
	}

	_, ok = props["async"]
	if ok {
		t.Error("async should be reset")
	}

	dfrd := b.GetDeferredProps()
	_, ok = dfrd["default"]
	if ok {
		t.Error("default group should be reset")
	}
}

func TestBag_Set(t *testing.T) {
	b := NewBag()

	err := b.Set("username", "john")
	if err != nil {
		t.Error(err)
	}

	props, err := b.GetProps()
	if err != nil {
		t.Error(err)
	}

	username, ok := props["username"]

	if !ok {
		t.Error("could not retrieve username")
	}

	if username != "john" {
		t.Error("invalid username returned")
	}
}

func TestBag_SetLazySync(t *testing.T) {
	b := NewBag()

	err := b.Set("username", NewLazyProp(func() (any, error) {
		return "john", nil
	}, "default", false, true))
	if err != nil {
		t.Error(err)
	}

	props, err := b.GetProps()
	if err != nil {
		t.Error(err)
	}

	username, ok := props["username"]

	if !ok {
		t.Error("could not retrieve username")
	}

	if username != "john" {
		t.Error("invalid username returned")
	}
}

func TestBag_SetLazy(t *testing.T) {
	b := NewBag()

	err := b.Set("username", NewLazyProp(func() (any, error) {
		return "john", nil
	}, "default", false, false))
	if err != nil {
		t.Error(err)
	}

	props, err := b.GetProps()
	if err != nil {
		t.Error(err)
	}

	username, ok := props["username"]

	if !ok {
		t.Error("could not retrieve username")
	}

	if username != "john" {
		t.Error("invalid username returned")
	}
}

func TestBag_SetDeferredSync(t *testing.T) {
	b := NewBag()

	err := b.Set("username", NewLazyProp(func() (any, error) {
		return "john", nil
	}, "default", true, true))
	if err != nil {
		t.Error(err)
	}

	err = b.Set("age", NewLazyProp(func() (any, error) {
		return 32, nil
	}, "age", true, true))
	if err != nil {
		t.Error(err)
	}

	err = b.Set("hobby", NewLazyProp(func() (any, error) {
		return "knitting", nil
	}, "default", true, true))
	if err != nil {
		t.Error(err)
	}

	props, err := b.GetProps()
	if err != nil {
		t.Error(err)
	}

	_, ok := props["username"]

	if ok {
		t.Error("deferred username should not be ok")
	}

	_, ok = props["age"]

	if ok {
		t.Error("deferred age should not be ok")
	}

	deferred := b.GetDeferredProps()

	defaultGroup, ok := deferred["default"]
	if !ok {
		t.Error("could not retrieve default deferred group")
	}

	if len(defaultGroup) != 2 {
		t.Error("invalid length for default group")
	}

	if !slices.Contains(defaultGroup, "username") || !slices.Contains(defaultGroup, "hobby") {
		t.Error("invalid group prop contents")
	}

	ageGroup, ok := deferred["age"]
	if !ok {
		t.Error("could not retrieve default age group")
	}
	if len(ageGroup) != 1 {
		t.Error("invalid length for age group")
	}

	if ageGroup[0] != "age" {
		t.Error("invalid prop name in age group")
	}
}

func TestBag_SetDeferred(t *testing.T) {
	b := NewBag()

	err := b.Set("username", NewLazyProp(func() (any, error) {
		return "john", nil
	}, "default", true, false))
	if err != nil {
		t.Error(err)
	}

	err = b.Set("age", NewLazyProp(func() (any, error) {
		return 32, nil
	}, "age", true, true))
	if err != nil {
		t.Error(err)
	}

	err = b.Set("hobby", NewLazyProp(func() (any, error) {
		return "knitting", nil
	}, "default", true, true))
	if err != nil {
		t.Error(err)
	}

	props, err := b.GetProps()
	if err != nil {
		t.Error(err)
	}

	_, ok := props["username"]

	if ok {
		t.Error("deferred username should not be ok")
	}

	_, ok = props["age"]

	if ok {
		t.Error("deferred age should not be ok")
	}

	deferred := b.GetDeferredProps()

	defaultGroup, ok := deferred["default"]
	if !ok {
		t.Error("could not retrieve default deferred group")
	}

	if len(defaultGroup) != 2 {
		t.Error("invalid length for default group")
	}

	if defaultGroup[0] != "username" {
		t.Error("invalid prop name in default group")
	}

	if defaultGroup[1] != "hobby" {
		t.Error("invalid prop name in default group")
	}

	ageGroup, ok := deferred["age"]
	if !ok {
		t.Error("could not retrieve default age group")
	}
	if len(ageGroup) != 1 {
		t.Error("invalid length for age group")
	}

	if ageGroup[0] != "age" {
		t.Error("invalid prop name in age group")
	}
}
