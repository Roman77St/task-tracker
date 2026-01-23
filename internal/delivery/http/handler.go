package httpHandler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"task-traker/internal/service"
)

type Handler struct {
	service *service.TaskService
}

type contextKey string

const userIDKey contextKey = "user_id"

type CreateTaskRequest struct {
	UserID   int64  `json:"user_id"`
	Title    string `json:"title"`
	Deadline string `json:"deadline"`
}

type loginRequest struct {
	UserID int64  `json:"user_id"`
	Code   string `json:"code"`
}

func NewHandler(s *service.TaskService) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) InitRouter() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /tasks", h.authMiddleware(http.HandlerFunc(h.getTasks)))
	mux.Handle("POST /tasks", h.authMiddleware(http.HandlerFunc(h.createTask)))
	mux.Handle("DELETE /tasks/{id}", h.authMiddleware(http.HandlerFunc(h.deleteTasks)))
	mux.HandleFunc("POST /login", h.login)
	mux.HandleFunc("POST /auth/refresh", h.Refresh)
	mux.Handle("POST /logout", h.authMiddleware(http.HandlerFunc(h.Logout)))


	return LoggingMiddleware(mux)
}

func (h *Handler) getTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		slog.Error("UserID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tasks, err := h.service.Repo.GetTasksByUserID(r.Context(), userID)
	if err != nil {
		slog.Error("HTTP getTasks error", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(tasks)
	if err != nil {
		slog.Error("JSON encode error", "error", err)
	}
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Error("JSON decode error", "error", err)
		http.Error(w, "JSON decode error", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.UserID == 0 || req.Deadline == "" {
		http.Error(w, "Title and UserID are required", http.StatusBadRequest)
		return
	}

	err = h.service.CreateTask(r.Context(), req.UserID, req.Title, req.Deadline)
	if err != nil {
		slog.Error("Service create task error", "error", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"created"}`))
}

func (h Handler) deleteTasks(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Missing task ID", http.StatusBadRequest)
		return
	}
	err := h.service.Repo.DeleteByID(r.Context(), idStr)
	if err != nil {
		slog.Error("Failed to delete task", "id", idStr, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
    var req loginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    tokenPair, err := h.service.Login(r.Context(), req.UserID, req.Code)
    if err != nil {
        slog.Warn("Failed login attempt", "user_id", req.UserID, "error", err)
        http.Error(w, "Invalid or expired code", http.StatusUnauthorized)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tokenPair)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Получаем Refresh токен
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tokens, err := h.service.Refresh(r.Context(), input.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Отдаем новую пару Access + Refresh
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(r.Body).Decode(&input)

	err := h.service.Logout(r.Context(), accessToken, input.RefreshToken)
	if err != nil {
		http.Error(w, "Logout failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
