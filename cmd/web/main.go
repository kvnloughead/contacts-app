package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", index)
	mux.HandleFunc("/contacts/view", contactView)
	mux.HandleFunc("/contacts/create", contactCreate)

	// Print a log a message to say that the server is starting.
	log.Print("starting server on :4000")

	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
