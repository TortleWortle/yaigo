package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/tortlewortle/go-inertia/examples/basic/web"
	"github.com/tortlewortle/go-inertia/pkg/inertia"
)

func main() {
	log.Println("preparing dist filesystem")
	frontend, err := web.FrontendFS()
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
			inertia.WithViteDevServer("http://localhost:5173", false),
		)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("creating PRODUCTION inertia server")
		inertiaServer, err = inertia.NewServer(frontend)
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

	mux.HandleFunc("GET /redirect", func(w http.ResponseWriter, r *http.Request) {
		inertia.Redirect(w, r, "/test")
	})

	mux.HandleFunc("GET /location", func(w http.ResponseWriter, r *http.Request) {
		inertia.Location(w, "https://google.com")
	})

	mux.HandleFunc("GET /404", func(w http.ResponseWriter, r *http.Request) {
		inertia.SetStatus(r, http.StatusNotFound)
		err := inertia.Render(w, r, "Error", inertia.Props{
			"status": http.StatusNotFound,
		})

		if err != nil {
			log.Printf("Could not render page: %v", err)
		}
	})

	mux.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		timeOfRender := time.Now().Format(time.TimeOnly)
		inertia.SetProp(r, "helperProp", "32 "+timeOfRender)
		err := inertia.Render(w, r, "TestPage", inertia.Props{
			"inlineProp": "Geoffrey " + timeOfRender,
			"time":       timeOfRender,
			// simple deferred prop, this request will take approx 500ms
			"deferredProp": inertia.DeferSync(func() (any, error) {
				time.Sleep(time.Millisecond * 500)
				return "deferred prop", nil
			}),
			// three props part of a propgroup, these will be fetched in a separate request.
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
		if err != nil {
			log.Printf("Could not render page: %v", err)
		}
	})

	mux.HandleFunc("GET /benchme", func(w http.ResponseWriter, r *http.Request) {
		inertia.SetProp(r, "helperProp", "32 ")
		err := inertia.Render(w, r, "BenchPage", inertia.Props{
			"inlineProp": "Geoffrey ",
			"time":       "go to bed",
			// two "long-running" prop resolves, so we run them concurrently
			"concurrentProp": inertia.Resolve(func() (any, error) {
				return "one!", nil
			}),
			"concurrentProp2": inertia.Resolve(func() (any, error) {
				return "two!", nil
			}),
		})
		if err != nil {
			log.Printf("Could not render page: %v", err)
		}
	})

	mux.HandleFunc("GET /clear_history", func(w http.ResponseWriter, r *http.Request) {
		inertia.ClearHistory(r)
		err := inertia.Render(w, r, "About", nil)
		if err != nil {
			log.Printf("Could not render page: %v", err)
		}
	})

	mux.HandleFunc("GET /brokenprop", func(w http.ResponseWriter, r *http.Request) {
		timeOfRender := time.Now().Format(time.TimeOnly)
		inertia.SetProp(r, "helperProp", "32 "+timeOfRender)
		defErr := inertia.SetProp(r, "deferredFromHelper", inertia.DeferSync(func() (any, error) {
			time.Sleep(time.Millisecond * 500)
			return "deferred prop from helper", nil
		}))
		if defErr == nil {
			fmt.Println("NO DEF ERROR")
		}
		err := inertia.Render(w, r, "TestPage", inertia.Props{
			"inlineProp": "Geoffrey " + timeOfRender,
			"time":       timeOfRender,
			"deferredPropInGroup": inertia.Defer(func() (any, error) {
				time.Sleep(time.Millisecond * 250)
				return nil, errors.New("uh oh!")
			}).Group("propgroup"),
		})
		if err != nil {
			log.Printf("Could not render page: %v", err)
			log.Println("rendering Error page")
			inertia.SetStatus(r, http.StatusInternalServerError)
			err = inertia.Render(w, r, "Error", inertia.Props{
				"status": http.StatusInternalServerError,
			})
			if err != nil {
				log.Printf("Could not render error page: %v", err)
			}
		}
	})

	mux.HandleFunc("GET /about", func(w http.ResponseWriter, r *http.Request) {
		err := inertia.Render(w, r, "About", nil)

		if err != nil {
			log.Printf("Could not render page: %v", err)
		}
	})

	fileServer := http.FileServer(http.FS(frontend))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			err := inertia.Render(w, r, "Index", nil)

			if err != nil {
				log.Printf("Could not render page: %v", err)
			}
		} else {
			_, err := fs.Stat(frontend, path.Clean(r.URL.Path)[1:])
			if errors.Is(err, os.ErrNotExist) {
				_ = inertia.Render(w, r, "Error", inertia.Props{
					"status": 404,
				})
				return
			}
			fileServer.ServeHTTP(w, r)
		}
	})

	log.Println("starting listener")
	http.ListenAndServe(":3000", logRequests(inertiaServer.Middleware(encryptsHistory(mux))))
}

func encryptsHistory(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inertia.EncryptHistory(r, true)
		next.ServeHTTP(w, r)
	})
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s - %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
