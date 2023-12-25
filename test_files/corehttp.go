package corehttp

import "net/http"

func HandlePost(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ondbyte"))
}

func HandleGet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ondbyte"))
}
