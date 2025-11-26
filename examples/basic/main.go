package main

import (
	"log"
	"net/http"

	"github.com/Mark-Bazylev/routerx"
)

func main() {
	router := routerx.New()

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
