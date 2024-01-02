package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()
	router.Handle("GET", "/hello", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var err error
		errS := ""
		yaduS := p.ByName("yadu")

		if yaduS == "" {
			errS += fmt.Sprintf("param 'yadu' is required\n")
		}

		yadu := new(string)
		if yaduS != "" {
			err = json.Unmarshal([]byte(yaduS), yadu)
			if err != nil {
				errS += "param value " + yaduS + " cannot be Unmarshalled into type string\n"
			}
		}

		if errS != "" {
			http.Error(w, errS, http.StatusTeapot)
			return
		}

		chinamyaS := p.ByName("chinamya")
		if chinamyaS == "" {
			errS += fmt.Sprintf("param 'chinamya' is required\n")
		}

		chinamya := new(uint)
		if chinamyaS != "" {
			err = json.Unmarshal([]byte(chinamyaS), chinamya)
			if err != nil {
				errS += "param value " + chinamyaS + " cannot be Unmarshalled into type uint\n"
			}
		}

		if errS != "" {
			http.Error(w, errS, http.StatusTeapot)
			return
		}

		yadu2S := p.ByName("yadu2")
		yadu2 := new(int)

		if yadu2S != "" {
			err = json.Unmarshal([]byte(yadu2S), yadu2)
			if err != nil {
				errS += "param value " + yadu2S + " cannot be Unmarshalled into type int\n"
			}
		}

		if errS != "" {
			http.Error(w, errS, http.StatusTeapot)
			return
		}

		yadu.Handle(yadu, chinamya, yadu2)
	})
}
