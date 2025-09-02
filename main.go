package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	apiCfg := &apiConfig{}
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidate)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	srv := http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	log.Fatal(srv.ListenAndServe())
	
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Hits reset to 0"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func handlerValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err.Error())
		marshallError(w, err, 500)
		return
	}
	if len(params.Body) > 140 {
		err = fmt.Errorf("chirp is too long")
		marshallError(w, err, 400)
		return
	}
	type response struct {
		ValidChirp string `json:"cleaned_body"`
	}
	respBody := response{
		ValidChirp: cleanString(params.Body),
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling response body: %s", err.Error())
		marshallError(w, err, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func marshallError(w http.ResponseWriter, err error, code int) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	w.WriteHeader(code)
	dat, err := json.Marshal(errorResponse{
		Error: err.Error(),
	})
	if err != nil {
		log.Printf("Error marshalling error response: %s", err.Error())
		return
	}
	w.Write(dat)
}

func cleanString(s string) string {
	var result string
	words := strings.Split(s, " ")
	for _, word := range words {
		if strings.ToLower(word) == "kerfuffle" || strings.ToLower(word) == "sharbert" || strings.ToLower(word) == "fornax" {
			word = "****"
		}
		result += word + " "
	}
	result = strings.TrimRight(result, " ")
	return result
}