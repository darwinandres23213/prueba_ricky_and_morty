package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

const baseURL = "https://rickandmortyapi.com/api"

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type RickMortyHandler struct {
	baseURL string
}

func NewRickMortyHandler(baseURL string) *RickMortyHandler {
	return &RickMortyHandler{
		baseURL: baseURL,
	}
}

// validateInternalRequest verifica que la petición venga del Gateway
func (h *RickMortyHandler) validateInternalRequest(r *http.Request) bool {
	// Verificar que la petición venga del Gateway usando un header interno
	internalAuth := r.Header.Get("X-Internal-Auth")
	return internalAuth == "gateway-service"
}

// sendErrorResponse envía una respuesta de error
func (h *RickMortyHandler) sendErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "error",
		"message": message,
	})
}

func sendJSONResponse(w http.ResponseWriter, statusCode int, status, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Status:  status,
		Message: message,
		Data:    data,
	})
}

func (h *RickMortyHandler) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateInternalRequest(r) {
		h.sendErrorResponse(w, http.StatusForbidden, "Acceso directo no permitido. Use el Gateway en el puerto 8080")
		return
	}

	// Configurar CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Manejar preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/")
	if path == "" {
		path = "character" // default endpoint
	}

	// Validar que el endpoint sea válido
	validEndpoints := map[string]bool{
		"character":  true,
		"location":   true,
		"episode":    true,
		"characters": true,
		"locations":  true,
		"episodes":   true,
	}

	// Normalizar el endpoint (remover 's' al final si existe)
	endpoint := strings.TrimSuffix(path, "s")
	if !validEndpoints[endpoint] && !validEndpoints[path] {
		sendJSONResponse(w, http.StatusBadRequest, "error", "Endpoint inválido. Use: character, location, o episode", nil)
		return
	}

	// Construir la URL final
	apiURL := h.baseURL + "/" + endpoint
	if id := r.URL.Query().Get("id"); id != "" {
		apiURL += "/" + id
	}

	// Crear un cliente HTTP con timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Crear el contexto con timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Crear la petición con el contexto
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "Error creando la petición", nil)
		return
	}

	// Realizar la petición
	resp, err := client.Do(req)
	if err != nil {
		sendJSONResponse(w, http.StatusBadGateway, "error", "Error consultando la API pública: "+err.Error(), nil)
		return
	}
	defer resp.Body.Close()

	// Leer el cuerpo de la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "Error leyendo la respuesta", nil)
		return
	}

	// Verificar el status code
	if resp.StatusCode != http.StatusOK {
		sendJSONResponse(w, resp.StatusCode, "error", "Error en la API pública", nil)
		return
	}

	// Parsear la respuesta JSON
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "Error procesando la respuesta", nil)
		return
	}

	// Enviar respuesta exitosa
	sendJSONResponse(w, http.StatusOK, "success", "Datos obtenidos exitosamente", data)
}

// GetCharacters maneja las peticiones de personajes
func (h *RickMortyHandler) GetCharacters(w http.ResponseWriter, r *http.Request) {
	if !h.validateInternalRequest(r) {
		h.sendErrorResponse(w, http.StatusForbidden, "Acceso directo no permitido. Use el Gateway en el puerto 8080")
		return
	}

	h.ProxyHandler(w, r)
}

// GetCharacter maneja las peticiones de un personaje específico
func (h *RickMortyHandler) GetCharacter(w http.ResponseWriter, r *http.Request) {
	if !h.validateInternalRequest(r) {
		h.sendErrorResponse(w, http.StatusForbidden, "Acceso directo no permitido. Use el Gateway en el puerto 8080")
		return
	}

	h.ProxyHandler(w, r)
}

// GetLocations maneja las peticiones de ubicaciones
func (h *RickMortyHandler) GetLocations(w http.ResponseWriter, r *http.Request) {
	if !h.validateInternalRequest(r) {
		h.sendErrorResponse(w, http.StatusForbidden, "Acceso directo no permitido. Use el Gateway en el puerto 8080")
		return
	}

	h.ProxyHandler(w, r)
}

// GetLocation maneja las peticiones de una ubicación específica
func (h *RickMortyHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	if !h.validateInternalRequest(r) {
		h.sendErrorResponse(w, http.StatusForbidden, "Acceso directo no permitido. Use el Gateway en el puerto 8080")
		return
	}

	h.ProxyHandler(w, r)
}

// GetEpisodes maneja las peticiones de episodios
func (h *RickMortyHandler) GetEpisodes(w http.ResponseWriter, r *http.Request) {
	if !h.validateInternalRequest(r) {
		h.sendErrorResponse(w, http.StatusForbidden, "Acceso directo no permitido. Use el Gateway en el puerto 8080")
		return
	}

	h.ProxyHandler(w, r)
}

// GetEpisode maneja las peticiones de un episodio específico
func (h *RickMortyHandler) GetEpisode(w http.ResponseWriter, r *http.Request) {
	if !h.validateInternalRequest(r) {
		h.sendErrorResponse(w, http.StatusForbidden, "Acceso directo no permitido. Use el Gateway en el puerto 8080")
		return
	}

	h.ProxyHandler(w, r)
}
