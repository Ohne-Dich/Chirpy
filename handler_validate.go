package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	auth "github.com/Ohne-Dich/Chirpy/internal"
	"github.com/Ohne-Dich/Chirpy/internal/database"
)

func (cfg *apiConfig) handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get token", err)
		return
	}

	uuid, err := auth.ValidateJWT(token, cfg.token_secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate jwt token", err)
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}

	// validating, get the body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
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
		UserID: uuid,
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

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	type returnVal struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get token", err)
		return
	}

	dat, err := cfg.dbQueries.GetRefreshTokenByToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get refresh token", err)
		return
	}

	if dat.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "revoked token", nil)
		return
	}

	if dat.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "expired token", nil)
		return
	}

	new_token, err := auth.MakeJWT(dat.UserID, cfg.token_secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, returnVal{
		Token: new_token,
	})
}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get token", err)
		return
	}

	err = cfg.dbQueries.SetRevokeRefreshToken(r.Context(), bearer)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
