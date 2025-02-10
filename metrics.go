package main

import (
	"fmt"
	"net/http"
)


// handle function for /admin/metrics endpoint
func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
    res := fmt.Sprintf(
        "<html>" +
            "<body>" +
                "<h1>Welcome, Chirpy Admin</h1>" +
                "<p>Chirpy has been visited %d times!</p>" + 
            "</body>" + 
        "</html>", int(cfg.fileserverHits.Load()))
    w.Write([]byte(res))

}


func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
	
}