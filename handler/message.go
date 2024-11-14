package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/CatalinPlesu/message-service/model"
	"github.com/CatalinPlesu/message-service/repository/message"
)

type Message struct {
	Repo *message.RedisRepo
}

func (h *Message) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ChannelID   uuid.UUID  `json:"channel_id"`
		ParentID    *uuid.UUID `json:"parent_id,omitempty"`
		UserID      uuid.UUID  `json:"user_id"`
		MessageText string     `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	theMessage := model.Message{
		MessageID:   uuid.New(),
		ChannelID:   body.ChannelID,
		ParentID:    body.ParentID,
		UserID:      body.UserID,
		MessageText: body.MessageText,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	err := h.Repo.Insert(r.Context(), theMessage)
	if err != nil {
		fmt.Println("failed to insert message:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(theMessage)
	if err != nil {
		fmt.Println("failed to marshal message:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}

func (h *Message) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	const decimal = 10
	const bitSize = 64
	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	const size = 50
	res, err := h.Repo.FindAll(r.Context(), message.FindAllPage{
		Offset: cursor,
		Size:   size,
	})
	if err != nil {
		fmt.Println("failed to find all messages:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Items []model.Message `json:"items"`
		Next  uint64          `json:"next,omitempty"`
	}
	response.Items = res.Messages
	response.Next = res.Cursor

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("failed to marshal messages:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (h *Message) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	messageID, err := uuid.Parse(idParam) // Parse as UUID
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	theMessage, err := h.Repo.FindByID(r.Context(), messageID)
	if errors.Is(err, message.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find message by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(theMessage); err != nil {
		fmt.Println("failed to marshal message:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Message) UpdateByID(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MessageText string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idParam := chi.URLParam(r, "id")

	messageID, err := uuid.Parse(idParam) // Parse as UUID
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	theMessage, err := h.Repo.FindByID(r.Context(), messageID)
	if errors.Is(err, message.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find message by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	now := time.Now().UTC()
	if body.MessageText != "" {
		theMessage.MessageText = *&body.MessageText
	}
	theMessage.UpdatedAt = &now

	err = h.Repo.Update(r.Context(), theMessage)
	if err != nil {
		fmt.Println("failed to update message:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(theMessage); err != nil {
		fmt.Println("failed to marshal message:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Message) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	messageID, err := uuid.Parse(idParam) // Parse as UUID
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.Repo.DeleteByID(r.Context(), messageID)
	if errors.Is(err, message.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to delete message by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
