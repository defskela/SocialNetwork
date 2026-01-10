package v1

import (
	"encoding/json"
	"net/http"

	"github.com/defskela/SocialNetwork/internal/service"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) Init(api *chi.Mux) {
	api.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.signUp)
		r.Post("/login", h.signIn)
	})
}

// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param input body service.SignUpInput true "Sign up input"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /auth/register [post]
func (h *Handler) signUp(w http.ResponseWriter, r *http.Request) {
	var input service.SignUpInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.services.Auth.SignUp(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
	})
}

// @Summary Sign in
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param input body service.SignInInput true "Sign in input"
// @Success 200 {object} service.Tokens
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /auth/login [post]
func (h *Handler) signIn(w http.ResponseWriter, r *http.Request) {
	var input service.SignInInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokens, err := h.services.Auth.SignIn(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(tokens)
}
