package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"go.tortle.tech/go-inertia/examples/basic/web"
	"go.tortle.tech/go-inertia/pkg/inertia"
)

func main() {
	log.Println("preparing dist filesystem")
	dist := web.Dist
	frontend, err := fs.Sub(dist, "dist")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("configuring vite dev")
	appEnv := os.Getenv("APP_ENV")

	// I don't really like this
	var inertiaServer *inertia.Server
	if appEnv == "local" {
		log.Println("creating LOCAL inertia server")
		inertiaServer, err = inertia.NewServer(frontend,
			inertia.WithViteDevServer("http://localhost:5173"),
		)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("creating PRODUCTION inertia server")
		inertiaServer, err = inertia.NewServer(frontend,
			inertia.WithSSR("http://127.0.0.1:13714", "bundle.mjs"),
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("creating servemux")
	mux := http.NewServeMux()

	mux.HandleFunc("GET /news", func(w http.ResponseWriter, r *http.Request) {
		err := inertia.Render(w, r, "News", nil)

		if err != nil {
			log.Printf("Could not render page: %v", err)
		}
	})

	mux.HandleFunc("GET /jeff", func(w http.ResponseWriter, r *http.Request) {
		inertia.SetProp(r, "user", "Jeffrey")
		err := inertia.Render(w, r, "User", nil)

		if err != nil {
			log.Printf("Could not render page: %v", err)
		}
	})

	mux.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		timeOfRender := time.Now().Format(time.TimeOnly)
		inertia.SetProp(r, "helperProp", "32 "+timeOfRender)
		inertia.SetProp(r, "deferredFromHelper", inertia.DeferSync(func() (any, error) {
			time.Sleep(time.Millisecond * 500)
			return "deferred prop from helper", nil
		}))
		err := inertia.Render(w, r, "TestPage", inertia.Props{
			"inlineProp": "Geoffrey " + timeOfRender,
			"time":       timeOfRender,
			"deferredProp": inertia.DeferSync(func() (any, error) {
				time.Sleep(time.Millisecond * 500)
				return "deferred prop", nil
			}),
			"deferredPropInGroup": inertia.Defer(func() (any, error) {
				return nil, errors.New("uh oh!")
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
		if err != nil {
			log.Printf("Could not render page: %v", err)
			inertia.Render(w, r, "About", nil)
		}
	})

	mux.HandleFunc("GET /about", func(w http.ResponseWriter, r *http.Request) {
		err := inertia.Render(w, r, "About", nil)

		if err != nil {
			log.Printf("Could not render page: %v", err)
		}
	})

	mux.HandleFunc("GET /404", func(w http.ResponseWriter, r *http.Request) {
		err := inertia.Render(w, r, "About", nil)

		if err != nil {
			log.Printf("Could not render page: %v", err)
		}
		w.WriteHeader(404)
	})

	fileServer := http.FileServer(http.FS(frontend))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			err := inertia.Render(w, r, "Index", nil)

			if err != nil {
				log.Printf("Could not render page: %v", err)
			}
		} else {
			fileServer.ServeHTTP(w, r)
		}
	})

	log.Println("starting listener")
	http.ListenAndServe(":3000", logRequests(inertiaServer.Middleware(mux)))
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s - %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
