# go-inertia
Go adapter for InertiaJS

## Client usage
### Starter kits
Since the client-side installation deviates a little bit from the normal setup starter kits will be made available for use.
(soon)
### Installation
Client side usage and instructions can be read here: https://inertiajs.com/

## Basic usage
These docs are yet incomplete and not the clearest, it is recommended to take a peep at the example folder or basic examples!
### Install
`go get go.tortle.tech/go-inertia`

### Setup the server
```go
// base server with the frontend filesystem
inertiaServer, err = inertia.NewServer(frontend)

// with the vite dev server attached
inertiaServer, err = inertia.NewServer(frontend,
    inertia.WithViteDevServer("http://localhost:5173"),
)

// with server side rendering (coming soon)
inertiaServer, err = inertia.NewServer(frontend,
    inertia.WithSSR("http://127.0.0.1:13714", "bundle.mjs"),
)
```

### Running the middleware
You should wrap your servemux in the middleware to make use of the helper functions.

Without using the middleware you will not be able to benefit from the helper functions and will need to render using the InertiaServer instance.
```go
http.ListenAndServe(":3000", inertiaServer.Middleware(mux))
```
### Providing props
You can provide props from middleware or within the handler using one of the helpers or providing them to the page directly.

```go
func(w http.ResponseWriter, r *http.Request) {
    inertia.SetProp(r, "username", "John Doe")
}
```

### Rendering pages
#### Without props
```go
func(w http.ResponseWriter, r *http.Request) {
	// using the helper
    inertia.Render(w, r, "Welcome", nil)
	// using your own instance
    inertiaServer.Render(w, r, "Welcome", nil)
}
```
#### With props
```go
func(w http.ResponseWriter, r *http.Request) {
    err := inertia.Render(w, r, "BenchPage", inertia.Props{
        "username": "John Doe ",
    })
}
```

#### Resolving props concurrently in the same request
```go
    func(w http.ResponseWriter, r *http.Request) {
    err := inertia.Render(w, r, "BenchPage", inertia.Props{
        // This request will take 50ms
        // two "long-running" prop resolves, so we run them concurrently
        "concurrentProp": inertia.Resolve(func() (any, error) {
            timeOfRender := time.Now().Format(time.StampMilli)
            time.Sleep(time.Millisecond * 50)
            return timeOfRender, nil
        }),
        "concurrentProp2": inertia.Resolve(func() (any, error) {
            time.Sleep(time.Millisecond * 50)
            timeOfRender := time.Now().Format(time.StampMilli)
            return timeOfRender, nil
        }),
    })
}
```

#### Deferring props
```go
func(w http.ResponseWriter, r *http.Request) {
	// this initial response will happen instantly and not wait.
	// inertia will then fire off two separate requests to load each prop group.
    err := inertia.Render(w, r, "BenchPage", inertia.Props{
        // simple deferred prop, this request will take approx 500ms
        "deferredProp": inertia.DeferSync(func() (any, error) {
            time.Sleep(time.Millisecond * 500)
            return "deferred prop", nil
        }),
        // three props part of a propgroup, these will be fetched in a separate request from the other deferred prop.
        // this request will take approx 750ms even though we wait for a total of 2000ms
        // This is because we execute the Defer() calls concurrently
        // The DeferSync() callbacks are run after starting the Defer() callbacks
        // This makes the total time 750ms instead of 2000ms
        "deferredPropInGroup": inertia.Defer(func() (any, error) {
            time.Sleep(time.Millisecond * 750)
            return "one!", nil
        }).Group("propgroup"),
        "deferredPropInGroup2": inertia.Defer(func() (any, error) {
            time.Sleep(time.Millisecond * 750)
            return "two!", nil
        }).Group("propgroup"),
        // sync prop in same group as unsynced for extra dramatic effect
        "deferredPropInGroup3": inertia.DeferSync(func() (any, error) {
            time.Sleep(time.Millisecond * 500)
            return "three!", nil
        }).Group("propgroup"),
    })
}
```

### Encrypting & clearing history
https://inertiajs.com/history-encryption

#### Enabling encryption
```go
func encryptsHistory(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inertia.EncryptHistory(r, true)
		next.ServeHTTP(w, r)
	})
}

http.ListenAndServe(":3000", inertiaServer.Middleware(encryptsHistory(mux)))
```

#### Clearing history
```go
func(w http.ResponseWriter, r *http.Request) {
    inertia.ClearHistory(r)
}
```

### Redirecting
https://inertiajs.com/redirects

#### Redirecting client
```go
// redirects using status 303 to ensure GET requests
func(w http.ResponseWriter, r *http.Request) {
    inertia.Redirect(w, r, "/")
}
```
#### Redirecting to external websites
```go
// sends a 409 conflict with the X-Inertia-Location header.
func(w http.ResponseWriter, r *http.Request) {
    inertia.Location(w, "https://google.com")
}
```

## Todo
### Priorities
- [ ] SSR
- [ ] test coverage, the state of tests is sad right now
- [ ] test helpers for testing your own application
- [ ] frontend starter kits for vue, react and svelte
### Maybe
- [ ] more complete starter kits (with ssr).

## Why
I liked using InertiaJS and I want to use it with Go.

## Goals
1. Understand how InertiaJS works.
2. Being able to use InertiaJS for quick project prototyping.

## Credit
This package uses ideas from both [romsar/gonertia](https://github.com/romsar/gonertia) and [petaki/inertia-go](https://github.com/petaki/inertia-go).
