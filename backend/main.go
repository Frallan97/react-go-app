package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"msg":"Hello from Go backend!"}`)
	})
	http.ListenAndServe(":8080", nil)
}
