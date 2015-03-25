package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	http.ListenAndServe(":8080", New())
}

func New() http.Handler {
	mux := http.NewServeMux()
	log := log.New(os.Stdout, "web ", log.LstdFlags)
	app := &app{mux, log}

	mux.HandleFunc("/foo", app.foo)

	return app
}

type app struct {
	mux *http.ServeMux
	log *log.Logger
}

func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *app) foo(w http.ResponseWriter, r *http.Request) {
	a.log.Println("request to foo")
}
