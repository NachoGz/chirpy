package main

import (
	"net/http"
	"os"
)

// handler function for /admin/reset endpoint
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	platform := os.Getenv("PLATFORM")
	if platform != "dev" {
		respondWithError(w, http.StatusForbidden, "This operation is not allowed in non-development environment", nil)
		return
	}
	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting users", err)
		return
	}


	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("fileserverHits counter reseted succesfully\n"))
}