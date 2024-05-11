package main

import (
	"fmt"
	"github.com/erykksc/kwikquiz/internal/routes"
	"log"
	"net/http"
)

func main() {
	router := routes.NewRouter()

	port := 3000
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server listening on http://localhost%s\n", addr)
	err := http.ListenAndServe(addr, router)
	log.Fatal(err)
}
