package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yourusername/api_ricky_and_morty/internal/auth/service"
	"golang.org/x/crypto/bcrypt"
)

var tokenStore = make(map[string]int)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, "error", "JSON inválido", nil)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "Error en hash", nil)
		return
	}
	if err := service.CreateUser(req.Username, string(hash)); err != nil {
		sendJSONResponse(w, http.StatusConflict, "error", "Usuario ya existe", nil)
		return
	}
	sendJSONResponse(w, http.StatusCreated, "success", "Usuario registrado exitosamente", nil)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, "error", "JSON inválido", nil)
		return
	}
	user, err := service.GetUserByUsername(req.Username)
	if err != nil {
		sendJSONResponse(w, http.StatusUnauthorized, "error", "Usuario o contraseña incorrectos", nil)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		sendJSONResponse(w, http.StatusUnauthorized, "error", "Usuario o contraseña incorrectos", nil)
		return
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "JWT secret not set", nil)
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"username": user.Username,
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "Error generating token", nil)
		return
	}
	tokenStore[tokenString] = 5
	http.SetCookie(w, &http.Cookie{
		Name:     os.Getenv("COOKIE_NAME"),
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
	})
	sendJSONResponse(w, http.StatusOK, "success", "Login exitoso", map[string]interface{}{
		"username": user.Username,
		"message":  "Token guardado en cookie",
	})
}

func ValidateTokenHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(os.Getenv("COOKIE_NAME"))
	if err != nil {
		sendJSONResponse(w, http.StatusUnauthorized, "error", "Token no encontrado en cookie", nil)
		return
	}

	tokenString := cookie.Value
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		sendJSONResponse(w, http.StatusInternalServerError, "error", "JWT secret not set", nil)
		return
	}

	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		sendJSONResponse(w, http.StatusUnauthorized, "error", "Token inválido", nil)
		return
	}

	usos, ok := tokenStore[tokenString]
	if !ok {
		sendJSONResponse(w, http.StatusUnauthorized, "error", "Token expirado", nil)
		return
	}

	if usos <= 0 {
		// Eliminar el token del store
		delete(tokenStore, tokenString)
		// Eliminar la cookie
		http.SetCookie(w, &http.Cookie{
			Name:     os.Getenv("COOKIE_NAME"),
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1, // Eliminar la cookie
		})
		sendJSONResponse(w, http.StatusUnauthorized, "error", "Token expirado por uso máximo alcanzado", nil)
		return
	}

	// Reducir el contador de usos
	tokenStore[tokenString] = usos - 1

	// Si es el último uso, eliminar el token
	if usos-1 == 0 {
		delete(tokenStore, tokenString)
		// Eliminar la cookie
		http.SetCookie(w, &http.Cookie{
			Name:     os.Getenv("COOKIE_NAME"),
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1, // Eliminar la cookie
		})
	}

	sendJSONResponse(w, http.StatusOK, "success", "Token válido", map[string]interface{}{
		"usos_restantes": usos - 1,
		"username":       claims["username"],
		"message":        "Token expirará después de este uso",
	})
}
