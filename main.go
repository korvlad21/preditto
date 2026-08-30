package main

import (
	"log"
	"net/http"

	"github.com/korvlad21/firstGoWeb/handler"
)

func main() {
	helloHandler := handler.NewHelloHandler()

	http.HandleFunc("/", helloHandler.Index)
	http.HandleFunc("/hello", helloHandler.Hello)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
