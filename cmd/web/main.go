package main

import "net/http"

func main() {
	http.HandleFunc("/targets", func(rw http.ResponseWriter, r *http.Request) {
		for _, t := range ex.Targets() {
			rw.Write([]byte(t))
		}
	})
	http.ListenAndServe(":8080", nil)
}
