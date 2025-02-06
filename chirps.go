package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"github.com/google/uuid"
	"github.com/NachoGz/chirpy/internal/database"
)

func (cfg *apiConfig) handlerChirpCreation(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	ok := validateChirps(params.Body)
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(params.Body, badWords)
	params.Body = cleaned

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:	params.Body,
		UserID:	uuid.NullUUID{UUID:	params.UserID, Valid: true}, // Convert uuid.UUID to uuid.NullUUID
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp", err)
		return
	}

	new_chirp := Chirp {
		ID: 		chirp.ID,
		CreatedAt: 	chirp.CreatedAt,
		UpdatedAt: 	chirp.UpdatedAt,
		Body: 		chirp.Body,
		UserID:		chirp.UserID.UUID,
	}

	respondWithJSON(w, http.StatusCreated, new_chirp)
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}

func validateChirps(chirp string) bool {
	const maxChirpLength = 140
	if len(chirp) > maxChirpLength {
		return false
	}
	return true
}

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}
	retrieved_chirps := []Chirp{}
	for _, chirp := range chirps {
		retrieved_chirps = append(retrieved_chirps, Chirp{
			ID: 		chirp.ID,
			CreatedAt: 	chirp.CreatedAt,
			UpdatedAt: 	chirp.UpdatedAt,
			Body: 		chirp.Body,
			UserID:		chirp.UserID.UUID,
		})
	}
	
	respondWithJSON(w, http.StatusOK, retrieved_chirps)
}

func (cfg *apiConfig) handlerChirpsRetrieveByID(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse chirpID", err)
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't retrieve chirp", err)
		return
	}
	
	respondWithJSON(w, http.StatusOK, Chirp{
		ID: 		chirp.ID,
		CreatedAt: 	chirp.CreatedAt,
		UpdatedAt: 	chirp.UpdatedAt,
		Body: 		chirp.Body,
		UserID:		chirp.UserID.UUID,
	})
}