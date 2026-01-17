package v1

import (
	"encoding/json"
	"net/http"

	"github.com/defskela/SocialNetwork/internal/entity"
	"github.com/defskela/SocialNetwork/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Handler struct {
	services  *service.Service
	validator *validator.Validate
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{
		services:  services,
		validator: validator.New(),
	}
}

func (h *Handler) Init(api *chi.Mux) {
	api.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.signUp)
		r.Post("/login", h.signIn)
	})

	api.Route("/users", func(r chi.Router) {
		r.Use(h.userIdentity)
		r.Get("/me", h.getProfile)
		r.Patch("/me", h.updateProfile)

		r.Post("/{id}/follow", h.followUser)
		r.Delete("/{id}/follow", h.unfollowUser)
		r.Get("/{id}/followers", h.getFollowers)
		r.Get("/{id}/following", h.getFollowing)
	})

	api.Route("/posts", func(r chi.Router) {
		r.Use(h.userIdentity)
		r.Post("/", h.createPost)
		r.Get("/{id}", h.getPost)
		r.Patch("/{id}", h.updatePost)
		r.Delete("/{id}", h.deletePost)
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

	if err := h.validator.Struct(input); err != nil {
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
	if err := h.validator.Struct(input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tokens, err := h.services.Auth.SignIn(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(tokens)
}

// @Summary Get user profile
// @Description Get current user profile information
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} entity.User
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /users/me [get]
func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(CtxKeyUserID).(uuid.UUID)
	if !ok {
		http.Error(w, "user id not found", http.StatusInternalServerError)
		return
	}

	var user *entity.User
	var err error
	user, err = h.services.User.GetProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

// @Summary Update user profile
// @Description Update current user profile information
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body service.UpdateUserInput true "Update input"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /users/me [patch]
func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(CtxKeyUserID).(uuid.UUID)
	if !ok {
		http.Error(w, "user id not found", http.StatusInternalServerError)
		return
	}

	var input service.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := h.services.User.UpdateProfile(r.Context(), userID, input); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
