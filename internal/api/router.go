package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func NewRouter(handler *Handler) http.Handler {
	r := mux.NewRouter()

	// API v1 routes
	v1 := r.PathPrefix("/api/v1").Subrouter()

	// TVL endpoints
	v1.HandleFunc("/tvl", handler.GetTotalTVL).Methods("GET")
	v1.HandleFunc("/tvl/{protocol}", handler.GetProtocolTVL).Methods("GET")
	v1.HandleFunc("/tvl/{protocol}/history", handler.GetHistoricalTVL).Methods("GET")

	// Protocol endpoints
	v1.HandleFunc("/protocols", handler.GetProtocols).Methods("GET")

	// Chain endpoints
	v1.HandleFunc("/chains", handler.GetChains).Methods("GET")

	// Stats endpoint
	v1.HandleFunc("/stats", handler.GetStats).Methods("GET")

	// Health check
	v1.HandleFunc("/health", handler.HealthCheck).Methods("GET")

	// Apply middleware
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		ExposedHeaders: []string{"X-Cache"},
	}).Handler(r)

	return corsHandler
}
