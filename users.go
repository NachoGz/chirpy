package main

import (
	"net/http"
	"encoding/json"
	"github.com/NachoGz/chirpy/internal/auth"
	"github.com/NachoGz/chirpy/internal/database"
	"time"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
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
		CreatedAt: 		user.CreatedAt,
		UpdatedAt: 		user.UpdatedAt,
		Email: 			user.Email,
		IsChirpyRed:	user.IsChirpyRed,
	}
	respondWithJSON(w, http.StatusCreated, new_user)
}


func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
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
		ID           	uuid.UUID `json:"id"`
		CreatedAt    	time.Time `json:"created_at"`
		UpdatedAt    	time.Time `json:"updated_at"`
		Email        	string    `json:"email"`
		IsChirpyRed		bool	  `json:"is_chirpy_red"`
		Token        	string    `json:"token"`
		RefreshToken 	string    `json:"refresh_token"`
	}{
		ID:           	user.ID,
		CreatedAt:    	user.CreatedAt,
		UpdatedAt:    	user.UpdatedAt,
		Email:        	user.Email,
		IsChirpyRed:	user.IsChirpyRed,
		Token:        	access_token,
		RefreshToken: 	refresh_token,
	}

	respondWithJSON(w, http.StatusOK, response)
}


func (cfg *apiConfig) handleUpdateUserInfo(w http.ResponseWriter, r *http.Request) {
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


	// Extract token from the header
	bearer_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid authorization token", err)
		return
	}


	// Validate the JWT and extract user ID
	userID, err := auth.ValidateJWT(bearer_token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect token", err)
		return
	}


	hashed_passwd, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}


	user, err := cfg.db.ChangeEmailAndPassword(r.Context(), database.ChangeEmailAndPasswordParams{
		ID:					userID,
		Email:				params.Email, 
		HashedPassword:		hashed_passwd, 
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update email or password", err)
		return
	}


	respondWithJSON(w, http.StatusOK, User{
		ID:				userID,
		UpdatedAt:		user.UpdatedAt,
		CreatedAt:		user.CreatedAt,
		Email:			user.Email,
		IsChirpyRed:	user.IsChirpyRed,
	})
}


func (cfg *apiConfig) handleUpgradedToChirpyRed(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	api_key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "ApiKey not found", err)
		return
	}
	if cfg.PolkaKey != api_key {
		respondWithError(w, http.StatusUnauthorized, "Incorrect ApiKey", err)
		return
	}


	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}


	err = cfg.db.UpgradeToChirpyRed(r.Context(), params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}


	w.WriteHeader(http.StatusNoContent)
}