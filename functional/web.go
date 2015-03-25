package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request to /")
	})

	http.ListenAndServe(":8080", nil)
}
