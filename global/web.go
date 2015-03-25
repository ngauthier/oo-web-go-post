package main

import (
	"log"
	"net/http"
	"os"
)

var (
	logger *log.Logger
)

func main() {
	logger = log.New(os.Stdout, "web ", log.LstdFlags)

	server := &http.Server{
		Addr:    ":8080",
		Handler: routes(),
	}

	server.ListenAndServe()
}

func routes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("/foo", foo)

	return r
}

func foo(w http.ResponseWriter, r *http.Request) {
	logger.Println("request to foo")
}
