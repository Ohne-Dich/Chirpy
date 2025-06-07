package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Ohne-Dich/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body    string    `json:"body"`
		User_ID uuid.UUID `json:"user_id"`
	}
	type returnVals struct {
		Id         uuid.UUID `json:"id"`
		Created_at time.Time `json:"created_at"`
		Updated_at time.Time `json:"updated_at"`
		Body       string    `json:"body"`
		User_id    uuid.UUID `json:"user_id"`
	}
	// validating, get the body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	// validating, get the body length checked
	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	// validating, get the body cleaned
	cleaned := cleanBody(params.Body)

	//get it saved
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: params.User_ID,
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Chirp wasn't able to be saved", nil)
		return
	}

	//returning
	respondWithJSON(w, http.StatusCreated, returnVals{
		Id:         chirp.ID,
		Created_at: chirp.CreatedAt,
		Updated_at: chirp.UpdatedAt,
		Body:       chirp.Body,
		User_id:    chirp.UserID,
	})

}

func cleanBody(body string) string {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Split(body, " ")
	for i, word := range words {
		lower := strings.ToLower(word)
		if _, ok := badWords[lower]; ok {
			words[i] = "****"
		}
	}
	clean := strings.Join(words, " ")
	return clean
}
