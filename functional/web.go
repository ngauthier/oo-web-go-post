package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request to /")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
