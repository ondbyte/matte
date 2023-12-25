package router

import "net/http"

// @Router /hello [get]
func Hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello!"))
}

// @Router /HowAreYou [get]
func HowAreYou(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("I'm fine"))
}
