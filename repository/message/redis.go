package message

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"

	"github.com/CatalinPlesu/message-service/model"
)

type RedisRepo struct {
	Client *redis.Client
}

func messageIDKey(id uuid.UUID) string {
	return fmt.Sprintf("message:%s", id.String())
}

func (r *RedisRepo) Insert(ctx context.Context, message model.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	key := messageIDKey(message.MessageID)

	txn := r.Client.TxPipeline()

	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set message: %w", err)
	}

	if err := txn.SAdd(ctx, "messages", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add message to set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}

var ErrNotExist = errors.New("message does not exist")

func (r *RedisRepo) FindByID(ctx context.Context, id uuid.UUID) (model.Message, error) {
	key := messageIDKey(id)

	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Message{}, ErrNotExist
	} else if err != nil {
		return model.Message{}, fmt.Errorf("failed to get message: %w", err)
	}

	var message model.Message
	err = json.Unmarshal([]byte(value), &message)
	if err != nil {
		return model.Message{}, fmt.Errorf("failed to decode message json: %w", err)
	}

	return message, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	key := messageIDKey(id)

	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("failed to delete message: %w", err)
	}

	if err := txn.SRem(ctx, "messages", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove message from set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}

func (r *RedisRepo) Update(ctx context.Context, message model.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	key := messageIDKey(message.MessageID)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	return nil
}

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Messages  []model.Message
	Cursor uint64
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "messages", page.Offset, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get message ids: %w", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Messages: []model.Message{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get messages: %w", err)
	}

	messages := make([]model.Message, len(xs))

	for i, x := range xs {
		x := x.(string)
		var message model.Message

		err := json.Unmarshal([]byte(x), &message)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to decode message json: %w", err)
		}

		messages[i] = message
	}

	return FindResult{
		Messages: messages,
		Cursor: cursor,
	}, nil
}
