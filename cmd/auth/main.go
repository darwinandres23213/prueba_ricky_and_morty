package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/yourusername/api_ricky_and_morty/internal/auth/handler"
	"github.com/yourusername/api_ricky_and_morty/internal/auth/service"
)

func main() {
	_ = godotenv.Load()
	if err := service.InitDB(); err != nil {
		log.Fatalf("No se pudo inicializar la base de datos: %v", err)
	}
	r := mux.NewRouter()

	// Crear subrouter para api/v1
	apiV1 := r.PathPrefix("/api/v1").Subrouter()

	// Endpoints de autenticación bajo api/v1
	apiV1.HandleFunc("/login", handler.LoginHandler).Methods("POST")
	apiV1.HandleFunc("/validate", handler.ValidateTokenHandler).Methods("GET")
	apiV1.HandleFunc("/register", handler.RegisterHandler).Methods("POST")

	port := os.Getenv("AUTH_SERVICE_PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("[AUTH] Running on :%s", port)
	h := cors.Default().Handler(r)
	log.Fatal(http.ListenAndServe(":"+port, h))
}
