package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Merhaba! Ben canlı bir API sunucusuyum.")
	})

	fmt.Println("API backend çalışıyor :5678")
	http.ListenAndServe(":5678", nil)
}
