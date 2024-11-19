package message

import (
	"context"
	"fmt"

	"github.com/CatalinPlesu/message-service/model"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PostgresRepo struct {
	DB *bun.DB
}

func NewPostgresRepo(db *bun.DB) *PostgresRepo {
	return &PostgresRepo{DB: db}
}

func (p *PostgresRepo) Migrate(ctx context.Context) error {
	_, err := p.DB.NewCreateTable().
		Model((*model.Message)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create messages table: %w", err)
	}
	return nil
}

func (p *PostgresRepo) Insert(ctx context.Context, message model.Message) error {
	_, err := p.DB.NewInsert().Model(&message).Exec(ctx)
	if err != nil {
		p.Migrate(ctx)
		return fmt.Errorf("failed to insert message: %w", err)
	}
	return nil
}

func (p *PostgresRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Message, error) {
	var message model.Message
	err := p.DB.NewSelect().Model(&message).Where("message_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find message by ID: %w", err)
	}
	return &message, nil
}

func (p *PostgresRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	_, err := p.DB.NewDelete().Model((*model.Message)(nil)).Where("message_id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

func (p *PostgresRepo) Update(ctx context.Context, message *model.Message) error {
	_, err := p.DB.NewUpdate().Model(message).Where("message_id = ?", message.MessageID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	return nil
}

type MessagePage struct {
	Messages []model.Message
	Cursor   uint64
}

func (r *PostgresRepo) FindAll(ctx context.Context, page FindAllPage) (MessagePage, error) {
	var messages []model.Message

	query := r.DB.NewSelect().
		Model(&messages).
		Order("message_id ASC").
		Limit(int(page.Size))

	if page.Offset > 0 {
		query.Where("message_id > ?", page.Offset)
	}

	err := query.Scan(ctx)
	if err != nil {
		return MessagePage{}, fmt.Errorf("failed to retrieve messages: %w", err)
	}

	if len(messages) == 0 {
		return MessagePage{
			Messages: []model.Message{},
			Cursor:   0,
		}, nil
	}

	return MessagePage{
		Messages: messages,
		Cursor:   page.Size + 100,
	}, nil
}

func (p *PostgresRepo) FindByChannelID(ctx context.Context, channelID uuid.UUID, page FindAllPage) ([]model.Message, uint64, error) {
	var messages []model.Message

	query := p.DB.NewSelect().
		Model(&messages).
		Where("channel_id = ?", channelID). 
		Order("created_at ASC").           
		Limit(int(page.Size))               

	err := query.Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve messages: %w", err)
	}

	var newCursor uint64
	if len(messages) > 0 {
		newCursor = page.Size + 100
	}

	return messages, newCursor, nil
}

func (p *PostgresRepo) FindByParentID(ctx context.Context, parentID uuid.UUID, page FindAllPage) ([]model.Message, uint64, error) {
	var messages []model.Message

	query := p.DB.NewSelect().
		Model(&messages).
		Where("parent_id = ?", parentID).
		Order("created_at DESC").
		Limit(int(page.Size))

	err := query.Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve child messages: %w", err)
	}

	var newCursor uint64
	if len(messages) > 0 {
		newCursor = page.Size + 100
	}

	return messages, newCursor, nil
}
