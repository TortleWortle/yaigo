package main

import (
	_ "embed"
	"io/fs"
	"log"
	"net/http"
	"os"

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

	mux.HandleFunc("GET /joff", func(w http.ResponseWriter, r *http.Request) {
		inertia.SetProp(r, "age", 32)
		err := inertia.Render(w, r, "User", inertia.Props{
			"user": "Geoffrey 2",
		})

		if err != nil {
			log.Printf("Could not render page: %v", err)
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
			fileServer.ServeHTTP(w, r)
		}
	})

	log.Println("starting listener")
	http.ListenAndServe(":3000", inertiaServer.Middleware(mux))
}
