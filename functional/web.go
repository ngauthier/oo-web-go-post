package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		log.Println("request to foo")
	})

	http.ListenAndServe(":8080", nil)
}
