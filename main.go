package main

import (
	"fmt"
	"net/http"
)

func handleReq(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Got req")

}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handleReq)

	http.ListenAndServe(":8000", mux)
}
