package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"github.com/yourusername/api_ricky_and_morty/internal/rickmorty/handler"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Obtener puerto de las variables de entorno
	port := os.Getenv("RICKMORTY_SERVICE_PORT")
	if port == "" {
		port = "8082"
	}

	// Crear el router
	router := mux.NewRouter()

	// Crear el handler
	rickMortyHandler := handler.NewRickMortyHandler("https://rickandmortyapi.com/api")

	// Configurar rutas
	api := router.PathPrefix("/api/v1").Subrouter()

	// Rutas de personajes
	api.HandleFunc("/characters", rickMortyHandler.GetCharacters).Methods("GET")
	api.HandleFunc("/character", rickMortyHandler.GetCharacter).Methods("GET")
	api.HandleFunc("/character/{id}", rickMortyHandler.GetCharacter).Methods("GET")

	// Rutas de ubicaciones
	api.HandleFunc("/locations", rickMortyHandler.GetLocations).Methods("GET")
	api.HandleFunc("/location", rickMortyHandler.GetLocation).Methods("GET")
	api.HandleFunc("/location/{id}", rickMortyHandler.GetLocation).Methods("GET")

	// Rutas de episodios
	api.HandleFunc("/episodes", rickMortyHandler.GetEpisodes).Methods("GET")
	api.HandleFunc("/episode", rickMortyHandler.GetEpisode).Methods("GET")
	api.HandleFunc("/episode/{id}", rickMortyHandler.GetEpisode).Methods("GET")

	// Configurar CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "X-Internal-Auth"},
		AllowCredentials: true,
	}).Handler(router)

	// Iniciar el servidor
	log.Printf("Rick and Morty service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, corsHandler); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
