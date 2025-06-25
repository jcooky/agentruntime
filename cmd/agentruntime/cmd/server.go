package cmd

import (
	"context"
	"encoding/json"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/habiliai/agentruntime/entity"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Model struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Message struct {
	Model

	ThreadID uint                        `json:"thread_id"`
	Content  string                      `json:"content"`
	User     string                      `json:"user"`
	Actions  datatypes.JSONSlice[Action] `json:"tool_calls"`
	Thread   Thread                      `json:"-" gorm:"foreignKey:ThreadID"`
}

type Action struct {
	Name   string          `json:"name"`
	Args   json.RawMessage `json:"args"`
	Result json.RawMessage `json:"result"`
}

type Thread struct {
	Model

	Instruction  string                      `json:"instruction"`
	Participants datatypes.JSONSlice[string] `json:"participants"`

	History []Message `json:"messages" gorm:"foreignKey:ThreadID"`
}

func createThreadsRouter(router *mux.Router, db *gorm.DB, messageCh chan *Message) {

	// Create a new thread
	router.HandleFunc("/threads", func(w http.ResponseWriter, r *http.Request) {
		tx := db.WithContext(r.Context())

		var thread Thread
		if err := json.NewDecoder(r.Body).Decode(&thread); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := tx.Create(&thread).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]any{
			"id": thread.ID,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("POST")

	router.HandleFunc("/threads", func(w http.ResponseWriter, r *http.Request) {
		tx := db.WithContext(r.Context())

		var threads []Thread
		if err := tx.Find(&threads).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(threads); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("GET")

	router.HandleFunc("/threads/{id}", func(w http.ResponseWriter, r *http.Request) {
		tx := db.WithContext(r.Context())

		vars := mux.Vars(r)
		id := vars["id"]

		if err := tx.Delete(&Thread{}, id).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("DELETE")

	// Get a thread by id
	router.HandleFunc("/threads/{id}", func(w http.ResponseWriter, r *http.Request) {
		tx := db.WithContext(r.Context())

		vars := mux.Vars(r)
		id := vars["id"]

		var thread Thread
		if err := tx.First(&thread, "id = ?", id).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]any{
			"id":           thread.ID,
			"instruction":  thread.Instruction,
			"participants": thread.Participants,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("GET")

	router.HandleFunc("/threads/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
		tx := db.WithContext(r.Context())

		vars := mux.Vars(r)
		id := vars["id"]

		var req struct {
			Message string `json:"message"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		threadId, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		msg := Message{
			ThreadID: uint(threadId),
			Content:  req.Message,
			User:     "USER",
		}

		if err := tx.Create(&msg).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		messageCh <- &msg
		w.WriteHeader(http.StatusOK)

	}).Methods("POST")

	router.HandleFunc("/threads/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
		tx := db.WithContext(r.Context())

		vars := mux.Vars(r)
		id := vars["id"]

		var messages []Message
		if r := tx.Find(&messages, "thread_id = ?", id); r.Error != nil {
			http.Error(w, r.Error.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(messages); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("GET")
}

func createServerHandler(agents map[string]entity.Agent, db *gorm.DB, logger *slog.Logger, messageCh chan *Message) (http.Handler, error) {
	router := mux.NewRouter()
	createThreadsRouter(router, db, messageCh)
	router.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		agents := slices.Collect(maps.Values(agents))

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(agents); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("GET")

	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)
	recovery := handlers.RecoveryHandler(handlers.PrintRecoveryStack(true), handlers.RecoveryLogger(slog.NewLogLogger(logger.Handler(), slog.LevelError)))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		router.ServeHTTP(w, r.WithContext(ctx))
	})

	return cors(recovery(handler)), nil
}
