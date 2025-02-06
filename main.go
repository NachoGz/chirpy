package main

import (
	"net/http"
	"log"
	"sync/atomic"
    "github.com/NachoGz/chirpy/internal/database"
    _ "github.com/lib/pq" // PostgreSQL driver
    "github.com/joho/godotenv" // For loading .env files
    "database/sql"
    "os"
	"time"
	"github.com/google/uuid"
)


type apiConfig struct {
	fileserverHits  atomic.Int32
    db         		*database.Queries

}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}


func main() {
    const filepathRoot = "."
	const port = "8080"

    godotenv.Load()
    dbURL := os.Getenv("DB_URL")
    if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
    
    dbConn, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatalf("Could not connect to database: %v", err)
    }
    defer dbConn.Close()
    
    dbQueries := database.New(dbConn)

	apiCfg := apiConfig{
        fileserverHits: atomic.Int32{},
        db: 			dbQueries,
    }

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/", fsHandler)
	
	// endpoints
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
    mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidate)
    mux.HandleFunc("POST /api/users", apiCfg.handlerUserCreation)
	

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())

	return 
}