package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/diamondoughnut/httpChirpy/internal/auth"
	"github.com/diamondoughnut/httpChirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Configuration struct holding application state and database connection
type apiConfig struct {
	fileserverHits atomic.Int32
	databaseQueries *database.Queries
	platform string
	secretKey string
	userId uuid.UUID
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token	  string	`json:"token"`
}

func main() {
	// Load environment variables and establish database connection
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secretKey := os.Getenv("JWT_SECRET_KEY")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	// Initialize application configuration with database queries
	apiCfg := &apiConfig{databaseQueries: dbQueries, platform: platform, secretKey: secretKey}
	// Set up HTTP router and register route handlers
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirpById)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/users", apiCfg.handlerRegister)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	// Configure and start HTTP server
	srv := http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	log.Fatal(srv.ListenAndServe())
	
}

// Health check endpoint returning 200 OK status
func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

// Admin metrics page displaying current hit count in HTML format
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())))
}

// Admin endpoint to reset hit counter to zero
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	err := cfg.databaseQueries.DeleteUsers(r.Context())
	if err != nil {
		log.Printf("Error deleting users: %s", err.Error())
		marshallError(w, err, 500)
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Hits reset to 0"))
}

// Middleware that increments hit counter for each request before passing to next handler
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}



func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	// decode JSON body
	decoder := json.NewDecoder(r.Body)
	params := database.CreateChirpParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err.Error())
		marshallError(w, err, 500)
		return
	}
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting bearer token: %s", err.Error())
		marshallError(w, err, 401)
		return
	}
	userId, err := auth.ValidateJWT(bearerToken, cfg.secretKey)
	// Validate chirp length (140 character limit)
	respBody, err := validate(params)
	if err != nil {
		log.Printf("Error validating chirp: %s", err.Error())
		marshallError(w, err, 400)
		return
	}
	// Create chirp in database
	chirp, err := cfg.databaseQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: respBody, UserID: cfg.userId})
	if err != nil {
		log.Printf("Error creating chirp: %s", err.Error())
		marshallError(w, err, 500)
		return
	}
	type response struct {
		ID uuid.UUID `json:"id"`
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	resp := response{
		ID: chirp.ID,
		Body: respBody,
		UserID: userId,
	}
	// Marshal response to JSON
	dat, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling response body: %s", err.Error())
		marshallError(w, err, 500)
		return
	}
	// Send successful JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
}

// helper functio nto validate and clean chirp messages, rejecting those over 140 characters
func validate(params database.CreateChirpParams) (string, error) {
	if len(params.Body) > 140 {
		err := fmt.Errorf("chirp is too long")
		return "", err
	}
	// Build response string with cleaned chirp content
	
	respBody := cleanString(params.Body)
	
	return respBody, nil
}

// Helper function to marshal and send error responses with specified status code
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

// Replaces profane words with asterisks and returns cleaned string
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

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
		ExpiresInSeconds int `json:"expires_in_seconds"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err.Error())
		marshallError(w, err, 500)
		return
	}
	// Validate user credentials
	user, err := cfg.databaseQueries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error getting user: %s", err.Error())
		marshallError(w, err, 404)
		return
	}
	err = auth.CheckHashPassword(params.Password, user.HashedPassword)
	if err != nil {
		log.Printf("Error checking password: %s", err.Error())
		marshallError(w, err, 401)
		return
	}
	if params.ExpiresInSeconds == 0 || params.ExpiresInSeconds > 3600 {
		params.ExpiresInSeconds = 3600
	}
	token, err := auth.MakeJWT(user.ID, cfg.secretKey, time.Duration(params.ExpiresInSeconds * int(time.Second)))
	response := User{
		ID: user.ID,
		Email: user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Token: token,
	}
	cfg.userId = user.ID
	// Marshal response to JSON
	dat, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling response body: %s", err.Error())
		marshallError(w, err, 500)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *apiConfig) handlerRegister(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err.Error())
		marshallError(w, err, 500)
		return
	}
	if cfg.platform != "dev" {
		log.Printf("Error: register endpoint only available in dev mode")
		marshallError(w, err, 403)
	}
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err.Error())
		marshallError(w, err, 500)
		return
	}
	user, err := cfg.databaseQueries.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: hashedPassword})
	if err != nil {
		log.Printf("Error creating user: %s", err.Error())
		marshallError(w, err, 500)
		return
	}
	data := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	cfg.userId = user.ID
	newUser, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling response body: %s", err.Error())
		marshallError(w, err, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(newUser)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.databaseQueries.GetChirps(r.Context())
	if err != nil {
		log.Printf("Error getting chirps: %s", err.Error())
		marshallError(w, err, 404)
		return
	}
	type responseItem struct {
		ID uuid.UUID `json:"id"`
		Body string `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	var responseItems []responseItem
	for _, chirp := range chirps {
		item := responseItem{
			ID: chirp.ID,
			Body: chirp.Body,
			UserId: chirp.UserID,
		}
		responseItems = append(responseItems, item)
	}
	// Marshal response to JSON
	dat, err := json.Marshal(responseItems)
	if err != nil {
		log.Printf("Error marshalling response body: %s", err.Error())
		marshallError(w, err, 500)
		return
	}
	// Send successful JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *apiConfig) handlerGetChirpById(w http.ResponseWriter, r *http.Request) {
	pathValue := r.PathValue("chirpID")
	path, err := uuid.Parse(pathValue)
	if err != nil {
		log.Printf("Error parsing chirp ID: %s", err.Error())
		marshallError(w, err, 400)
		return
	}
	chirp, err := cfg.databaseQueries.GetChirpById(r.Context(), path)
	if err != nil {
		log.Printf("Error getting chirp: %s", err.Error())
		marshallError(w, err, 404)
		return
	}
	type response struct {
		ID uuid.UUID `json:"id"`
		Body string `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	resp := response{
		ID: chirp.ID,
		Body: chirp.Body,
		UserId: chirp.UserID,
	}
	// Marshal response to JSON
	dat, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling response body: %s", err.Error())
		marshallError(w, err, 500)
		return
	}
	// Send successful JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}