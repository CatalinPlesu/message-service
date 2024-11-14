package model

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	MessageID   uuid.UUID  `json:"message_id"`          
	ChannelID   uuid.UUID  `json:"channel_id"`          
	ParentID    *uuid.UUID `json:"parent_id,omitempty"` 
	UserID      uuid.UUID  `json:"user_id"`             
	MessageText string     `json:"message"`             
	CreatedAt   *time.Time `json:"created_at"`          
	UpdatedAt   *time.Time `json:"updated_at"`          
}
