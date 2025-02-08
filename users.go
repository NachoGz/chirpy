package main

import (
	"net/http"
	"encoding/json"
	"github.com/NachoGz/chirpy/internal/auth"
	"github.com/NachoGz/chirpy/internal/database"
	"time"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUserCreation(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	hashed_passwd, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:		params.Email,
		HashedPassword:	hashed_passwd,		
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating user", err)
		return
	}
	new_user := User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}
	respondWithJSON(w, http.StatusCreated, new_user)
}


func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	
	// Set expiration time
	// If expires_in_seconds is not specified or greater than 3600 seconds, default to 3600
	// expirationTime := 3600 // Default to 1 hour
	// if params.ExpiresInSeconds > 0 && params.ExpiresInSeconds <= 3600 {
	// 	expirationTime = params.ExpiresInSeconds
	// }

	// Generate JWT with expiration time
	access_token, err := auth.MakeJWT(user.ID, cfg.secret, time.Duration(3600)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating JWT", err)
		return
	}

	// Create an refresh token
	refresh_token, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating refresh token", err)
		return
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token: refresh_token,
		UserID: uuid.NullUUID{UUID:	user.ID, Valid: true},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating refresh token", err)
		return
	}


	response := struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        access_token,
		RefreshToken: refresh_token,
	}

	respondWithJSON(w, http.StatusOK, response)

	
}