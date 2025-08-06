// Package auth implements HTTP handlers for authentication.
// This is part of the Adapters layer in Hexagonal Architecture.
// HTTP adapters handle HTTP requests and delegate business logic to application services.
package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"chess-backend/internal/utils"
	"chess-backend/internal/ports/services"

	"github.com/gorilla/mux"
)

// Handler handles HTTP requests for authentication
type Handler struct {
	authService services.AuthService
}

// NewHandler creates a new authentication HTTP handler
func NewHandler(authService services.AuthService) *Handler {
	return &Handler{
		authService: authService,
	}
}

// RegisterRoutes registers authentication routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	auth := router.PathPrefix("/auth").Subrouter()

	// Public routes
	auth.HandleFunc("/signup", h.SignupHandler).Methods("POST")
	auth.HandleFunc("/signin", h.SigninHandler).Methods("POST")
	auth.HandleFunc("/logout", h.LogoutHandler).Methods("POST")
	auth.HandleFunc("/me", h.GetCurrentUserHandler).Methods("GET")
	auth.HandleFunc("/refresh", h.RefreshSessionHandler).Methods("POST")
}

// SignupHandler handles user registration
func (h *Handler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	var req services.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Response.WriteBadRequest(w, "Invalid request body")
		return
	}

	resp, err := h.authService.Signup(r.Context(), req)
	if err != nil {
		utils.Response.WriteBadRequest(w, err.Error())
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    resp.SessionID,
		Path:     "/",
		MaxAge:   int((24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	utils.Response.WriteCreated(w, resp.Message, resp)
}

// SigninHandler handles user login
func (h *Handler) SigninHandler(w http.ResponseWriter, r *http.Request) {
	var req services.SigninRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Response.WriteBadRequest(w, "Invalid request body")
		return
	}

	resp, err := h.authService.Signin(r.Context(), req)
	if err != nil {
		utils.Response.WriteUnauthorized(w, err.Error())
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    resp.SessionID,
		Path:     "/",
		MaxAge:   int((24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	utils.Response.WriteSuccess(w, resp.Message, resp)
}

// LogoutHandler handles user logout
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		utils.Response.WriteBadRequest(w, "No active session")
		return
	}

	err = h.authService.Logout(r.Context(), cookie.Value)
	if err != nil {
		utils.Response.WriteInternalServerError(w, err.Error())
		return
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	utils.Response.WriteSuccess(w, "Logout successful", nil)
}

// GetCurrentUserHandler returns current user information
func (h *Handler) GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		utils.Response.WriteUnauthorized(w, "No active session")
		return
	}

	user, err := h.authService.GetCurrentUser(r.Context(), cookie.Value)
	if err != nil {
		utils.Response.WriteUnauthorized(w, "Invalid session")
		return
	}

	utils.Response.WriteSuccess(w, "User retrieved successfully", map[string]interface{}{
		"authenticated": true,
		"user":          user,
	})
}

// RefreshSessionHandler extends the current session
func (h *Handler) RefreshSessionHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		utils.Response.WriteBadRequest(w, "No active session")
		return
	}

	err = h.authService.RefreshSession(r.Context(), cookie.Value)
	if err != nil {
		utils.Response.WriteUnauthorized(w, err.Error())
		return
	}

	// Update session cookie with new expiration
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    cookie.Value,
		Path:     "/",
		MaxAge:   int((24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	utils.Response.WriteSuccess(w, "Session refreshed successfully", nil)
}