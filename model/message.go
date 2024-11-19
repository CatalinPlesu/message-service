package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Message struct {
	bun.BaseModel `bun:"table:messages"` // This tells Bun ORM to use the "messages" table.

	MessageID   uuid.UUID  `bun:"message_id,pk,type:uuid"` // Primary key, using UUID type.
	ChannelID   uuid.UUID  `bun:"channel_id,type:uuid"`    // Foreign key, using UUID type.
	ParentID    *uuid.UUID `bun:"parent_id,nullzero,type:uuid"` // Nullable field for parent ID.
	UserID      uuid.UUID  `bun:"user_id,type:uuid"`       // Foreign key, using UUID type.
	MessageText string     `bun:"message_text"`            // Column for the message content.
	CreatedAt   *time.Time `bun:"created_at,notnull,default:current_timestamp"` // Timestamp with default value.
	UpdatedAt   *time.Time `bun:"updated_at,notnull,default:current_timestamp"` // Timestamp with default value.
}


type MessageMin struct {
	ChannelID   uuid.UUID  `json:"channel_id"`   
	ParentID    *uuid.UUID `json:"parent_id,omitempty"` 
	UserID      uuid.UUID  `json:"user_id"`      
	MessageText string     `json:"message_text"`  
	CreatedAt   *time.Time `json:"created_at,omitempty"` 
}
