package main

import (
	"errors"
	"net/http"
	"time"

	auth "github.com/Ohne-Dich/Chirpy/internal"
	"github.com/google/uuid"
)

type returnVals struct {
	Id         uuid.UUID `json:"id"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	Body       string    `json:"body"`
	User_id    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {

	chirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
		return
	}
	var arr []returnVals

	for _, chirp := range chirps {
		arr = append(arr, returnVals{
			Id:         chirp.ID,
			Created_at: chirp.CreatedAt,
			Updated_at: chirp.UpdatedAt,
			Body:       chirp.Body,
			User_id:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, arr)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	value := r.PathValue("id")

	uuid, err := uuid.Parse(value)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get valid uuid from body", err)
		return
	}

	chirp, err := cfg.dbQueries.GetChirp(r.Context(), uuid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		Id:         chirp.ID,
		Created_at: chirp.CreatedAt,
		Updated_at: chirp.UpdatedAt,
		Body:       chirp.Body,
		User_id:    chirp.UserID,
	})
}

func (cfg *apiConfig) handleDeleteChirps(w http.ResponseWriter, r *http.Request) {
	// authentication
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get token", err)
		return
	}
	id, err := auth.ValidateJWT(token, cfg.token_secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate jwt token", err)
		return
	}

	// checking ownership
	value := r.PathValue("chirpID")
	chirpUuid, err := uuid.Parse(value)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get valid uuid from body", err)
		return
	}
	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpUuid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}
	if chirp.UserID != id {
		respondWithError(w, http.StatusForbidden, "Not owner of the chirp", errors.New("not owner of the chirp"))
		return
	}

	//actually deleting
	err = cfg.dbQueries.DeleteChirpById(r.Context(), chirpUuid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
