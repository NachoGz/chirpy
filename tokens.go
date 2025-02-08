package main

import (
	"github.com/NachoGz/chirpy/internal/auth"
	"time"
	"net/http"
	"log"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get the bearer token", err)
		return
	}
	log.Println("Bearer token:", token)


	ref_token, err := cfg.db.GetRefreshToken(r.Context(), token)
	log.Printf("Looking for refresh token: %v", ref_token.Token)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "The token doesn't exist", err)
		return
	} else if time.Now().After(ref_token.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has expired", nil)
		return
	}  else if ref_token.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token is revoked", nil)
		return
	}

	// Generate JWT with expiration time
	access_token, err := auth.MakeJWT(ref_token.UserID.UUID, cfg.secret, time.Duration(3600)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating JWT", err)
		return
	}

	respondWithJSON(w, http.StatusOK, struct {
		Token	string `json:"token"`
	}{
		Token:	access_token,
	})
}


func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get the bearer token", err)
		return
	}


	if err := cfg.db.RevokeRefreshToken(r.Context(), token); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't revoke the refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	
}