package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type GatewayHandler struct {
	client        *http.Client
	authPort      string
	rickMortyPort string
}

func NewGatewayHandler(authPort, rickMortyPort string) *GatewayHandler {
	return &GatewayHandler{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		authPort:      authPort,
		rickMortyPort: rickMortyPort,
	}
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

func RickMortyHandler(w http.ResponseWriter, r *http.Request) {
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

	// Crear cliente HTTP con timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Crear contexto con timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Validar token con Auth Service
	authURL := "http://auth:" + os.Getenv("AUTH_SERVICE_PORT") + "/api/v1/validate"
	req, err := http.NewRequestWithContext(ctx, "GET", authURL, nil)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "Error creando petición a Auth Service", nil)
		return
	}

	// Copiar cookies
	for _, c := range r.Cookies() {
		req.AddCookie(c)
	}

	// Validar token
	resp, err := client.Do(req)
	if err != nil {
		sendJSONResponse(w, http.StatusBadGateway, "error", "Error conectando con Auth Service", nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var authResponse Response
		if err := json.Unmarshal(body, &authResponse); err != nil {
			sendJSONResponse(w, resp.StatusCode, "error", "Error de autenticación", nil)
			return
		}
		sendJSONResponse(w, resp.StatusCode, authResponse.Status, authResponse.Message, authResponse.Data)
		return
	}

	// Obtener el path de la URL y limpiarlo (remover api/v1)
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

	// Si es válido, proxy a Rick & Morty Service
	rmURL := "http://rickmorty:" + os.Getenv("RICKMORTY_SERVICE_PORT") + "/api/v1/" + endpoint
	if id := r.URL.Query().Get("id"); id != "" {
		rmURL += "/" + id
	}

	rmReq, err := http.NewRequestWithContext(ctx, "GET", rmURL, nil)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "Error creando petición a Rick & Morty Service", nil)
		return
	}

	// Copiar cookies
	for _, c := range r.Cookies() {
		rmReq.AddCookie(c)
	}

	// Obtener datos
	rmResp, err := client.Do(rmReq)
	if err != nil {
		sendJSONResponse(w, http.StatusBadGateway, "error", "Error conectando con Rick & Morty Service", nil)
		return
	}
	defer rmResp.Body.Close()

	// Leer y enviar respuesta
	body, err := io.ReadAll(rmResp.Body)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "Error leyendo respuesta de Rick & Morty Service", nil)
		return
	}

	// Verificar si la respuesta es un error
	if rmResp.StatusCode != http.StatusOK {
		var errorResponse Response
		if err := json.Unmarshal(body, &errorResponse); err != nil {
			sendJSONResponse(w, rmResp.StatusCode, "error", "Error en el servicio Rick & Morty", nil)
			return
		}
		sendJSONResponse(w, rmResp.StatusCode, errorResponse.Status, errorResponse.Message, errorResponse.Data)
		return
	}

	// Si es una respuesta exitosa, enviar los datos
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "Error procesando respuesta del servicio", nil)
		return
	}

	sendJSONResponse(w, http.StatusOK, "success", "Datos obtenidos exitosamente", data)
}

func (h *GatewayHandler) proxyToRickMorty(w http.ResponseWriter, r *http.Request) {
	// Crear una nueva petición al servicio Rick and Morty
	rickMortyURL := fmt.Sprintf("http://rickmorty:%s%s", h.rickMortyPort, r.URL.Path)
	req, err := http.NewRequest(r.Method, rickMortyURL, nil)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "Error al crear la petición", nil)
		return
	}

	// Copiar los query parameters
	req.URL.RawQuery = r.URL.RawQuery

	// Agregar el header de autenticación interna
	req.Header.Set("X-Internal-Auth", "gateway-service")

	// Realizar la petición
	resp, err := h.client.Do(req)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "Error al comunicarse con el servicio Rick and Morty", nil)
		return
	}
	defer resp.Body.Close()

	// Leer la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "Error al leer la respuesta", nil)
		return
	}

	// Copiar los headers de la respuesta
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Enviar la respuesta al cliente
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// GetCharacters maneja las peticiones de personajes
func (h *GatewayHandler) GetCharacters(w http.ResponseWriter, r *http.Request) {
	// Validar token
	if !h.validateToken(w, r) {
		return
	}
	h.proxyToRickMorty(w, r)
}

// GetCharacter maneja las peticiones de un personaje específico
func (h *GatewayHandler) GetCharacter(w http.ResponseWriter, r *http.Request) {
	// Validar token
	if !h.validateToken(w, r) {
		return
	}
	h.proxyToRickMorty(w, r)
}

// GetLocations maneja las peticiones de ubicaciones
func (h *GatewayHandler) GetLocations(w http.ResponseWriter, r *http.Request) {
	// Validar token
	if !h.validateToken(w, r) {
		return
	}
	h.proxyToRickMorty(w, r)
}

// GetLocation maneja las peticiones de una ubicación específica
func (h *GatewayHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	// Validar token
	if !h.validateToken(w, r) {
		return
	}
	h.proxyToRickMorty(w, r)
}

// GetEpisodes maneja las peticiones de episodios
func (h *GatewayHandler) GetEpisodes(w http.ResponseWriter, r *http.Request) {
	// Validar token
	if !h.validateToken(w, r) {
		return
	}
	h.proxyToRickMorty(w, r)
}

// GetEpisode maneja las peticiones de un episodio específico
func (h *GatewayHandler) GetEpisode(w http.ResponseWriter, r *http.Request) {
	// Validar token
	if !h.validateToken(w, r) {
		return
	}
	h.proxyToRickMorty(w, r)
}

// validateToken valida el token con el servicio de autenticación
func (h *GatewayHandler) validateToken(w http.ResponseWriter, r *http.Request) bool {
	// Obtener el token de la cookie
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		sendJSONResponse(w, http.StatusUnauthorized, "error", "Token no encontrado", nil)
		return false
	}

	// Crear la petición al servicio de autenticación
	authURL := fmt.Sprintf("http://auth:%s/api/v1/validate", h.authPort)
	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "Error al crear la petición de validación", nil)
		return false
	}

	// Agregar la cookie al request
	req.AddCookie(cookie)

	// Realizar la petición
	resp, err := h.client.Do(req)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "Error al comunicarse con el servicio de autenticación", nil)
		return false
	}
	defer resp.Body.Close()

	// Si la validación falla, enviar el error al cliente
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		return false
	}

	return true
}
