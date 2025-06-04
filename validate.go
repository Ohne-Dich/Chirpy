package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnErr struct {
		Err string `json:"error"`
	}
	type returnValid struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respBody := returnErr{
			Err: "Something went wrong",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			fmt.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(500)
		w.Write(dat)
		return
	}

	if len(params.Body) > 140 {
		respBody := returnErr{
			Err: "Chirp is too long",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			fmt.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(400)
		w.Write(dat)
		return
	}
	respBody := returnValid{
		Valid: true,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(dat)
}
