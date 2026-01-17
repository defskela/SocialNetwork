package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/defskela/SocialNetwork/internal/entity"
)

// @Summary Follow a user
// @Description Follow a user by ID
// @Tags followers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID to follow"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /users/{id}/follow [post]
func (h *Handler) followUser(w http.ResponseWriter, r *http.Request) {
	followerID, ok := r.Context().Value(CtxKeyUserID).(uuid.UUID)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	followeeIDStr := chi.URLParam(r, "id")
	followeeID, err := uuid.Parse(followeeIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if followerID == followeeID {
		http.Error(w, "cannot follow yourself", http.StatusBadRequest)
		return
	}

	if err := h.services.Follower.Follow(r.Context(), followerID, followeeID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Unfollow a user
// @Description Unfollow a user by ID
// @Tags followers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID to unfollow"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /users/{id}/follow [delete]
func (h *Handler) unfollowUser(w http.ResponseWriter, r *http.Request) {
	followerID, ok := r.Context().Value(CtxKeyUserID).(uuid.UUID)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	followeeIDStr := chi.URLParam(r, "id")
	followeeID, err := uuid.Parse(followeeIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.services.Follower.Unfollow(r.Context(), followerID, followeeID); err != nil {
		if err.Error() == "relationship not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Get followers
// @Description Get list of users following the specified user
// @Tags followers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Success 200 {array} entity.User
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /users/{id}/followers [get]
func (h *Handler) getFollowers(w http.ResponseWriter, r *http.Request) {
	h.handleGetUsers(w, r, h.services.Follower.GetFollowers)
}

// @Summary Get following
// @Description Get list of users the specified user is following
// @Tags followers
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Success 200 {array} entity.User
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /users/{id}/following [get]
func (h *Handler) getFollowing(w http.ResponseWriter, r *http.Request) {
	h.handleGetUsers(w, r, h.services.Follower.GetFollowing)
}

func (h *Handler) handleGetUsers(
	w http.ResponseWriter,
	r *http.Request,
	getFunc func(context.Context, uuid.UUID) ([]entity.User, error),
) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	users, err := getFunc(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
