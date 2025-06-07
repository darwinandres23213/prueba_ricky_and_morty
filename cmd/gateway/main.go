package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"github.com/yourusername/api_ricky_and_morty/internal/gateway/handler"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Obtener puertos de las variables de entorno
	gatewayPort := os.Getenv("GATEWAY_SERVICE_PORT")
	if gatewayPort == "" {
		gatewayPort = "8080"
	}
	authPort := os.Getenv("AUTH_SERVICE_PORT")
	if authPort == "" {
		authPort = "8081"
	}
	rickMortyPort := os.Getenv("RICKMORTY_SERVICE_PORT")
	if rickMortyPort == "" {
		rickMortyPort = "8082"
	}

	// Crear el router
	router := mux.NewRouter()

	// Crear el handler
	gatewayHandler := handler.NewGatewayHandler(authPort, rickMortyPort)

	// Configurar rutas
	api := router.PathPrefix("/api/v1").Subrouter()

	// Rutas de personajes
	api.HandleFunc("/characters", gatewayHandler.GetCharacters).Methods("GET")
	api.HandleFunc("/character", gatewayHandler.GetCharacter).Methods("GET")
	api.HandleFunc("/character/{id}", gatewayHandler.GetCharacter).Methods("GET")

	// Rutas de ubicaciones
	api.HandleFunc("/locations", gatewayHandler.GetLocations).Methods("GET")
	api.HandleFunc("/location", gatewayHandler.GetLocation).Methods("GET")
	api.HandleFunc("/location/{id}", gatewayHandler.GetLocation).Methods("GET")

	// Rutas de episodios
	api.HandleFunc("/episodes", gatewayHandler.GetEpisodes).Methods("GET")
	api.HandleFunc("/episode", gatewayHandler.GetEpisode).Methods("GET")
	api.HandleFunc("/episode/{id}", gatewayHandler.GetEpisode).Methods("GET")

	// Configurar CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}).Handler(router)

	// Iniciar el servidor
	log.Printf("Gateway service starting on port %s", gatewayPort)
	if err := http.ListenAndServe(":"+gatewayPort, corsHandler); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
