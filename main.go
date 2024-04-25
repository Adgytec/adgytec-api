package main

import (
	"fmt"
	"net/http"
	"os"
)

func handleReq(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Got req")

}

func main() {
	PORT := "8080"
	if port := os.Getenv("PORT"); port != "" {
		PORT = port
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", handleReq)

	http.ListenAndServe(":"+PORT, mux)
}
