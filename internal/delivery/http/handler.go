package httpHandler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"task-traker/internal/service"
)


type Handler struct {
	service *service.TaskService
}

type CreateTaskRequest struct {
	UserID int64 `json:"user_id"`
	Title string `json:"title"`
	Deadline string `json:"deadline"`
}

func NewHandler(s *service.TaskService) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) InitRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /tasks", h.getTasks)
	mux.HandleFunc("POST /tasks", h.createTask)
	mux.HandleFunc("DELETE /tasks/{id}", h.deleteTasks)

	return LoggingMiddleWare(mux)
}

func (h *Handler) getTasks(w http.ResponseWriter, r *http.Request)  {
	// нужна авторизация!!! Только для тестов.
	userIDStr := r.URL.Query().Get("user")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

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

func (h Handler) deleteTasks(w http.ResponseWriter, r *http.Request)  {
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