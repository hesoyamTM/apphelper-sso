package redpanda

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
)

func (c *RedPandaClient) UserRegistered(ctx context.Context, user *UserRegisteredEvent) error {
	const op = "redpanda.RedPandaClient.UserRegistered"

	value, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := c.sendMessage(ctx, userRegisteredTopic, value); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *RedPandaClient) PasswordChanged(ctx context.Context, user *UserRegisteredEvent) error {
	const op = "redpanda.RedPandaClient.PasswordChanged"

	value, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := c.sendMessage(ctx, passwordChangedTopic, value); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *RedPandaClient) VerificationCodeUpdated(ctx context.Context, user *VerificationCodeUpdatedEvent) error {
	const op = "redpanda.RedPandaClient.VerificationCodeUpdated"

	value, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := c.sendMessage(ctx, verificationCodeUpdated, value); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *RedPandaClient) sendMessage(ctx context.Context, topic string, value []byte) error {
	const op = "redpanda.RedPandaClient.sendMessage"

	msg := sarama.ProducerMessage{
		Topic: userRegisteredTopic,
		Value: sarama.ByteEncoder(value),
	}

	select {
	case c.messageChan <- &msg:
		return nil
	case <-c.stopChan:
		return fmt.Errorf("%s: client stopped", op)
	case <-ctx.Done():
		return fmt.Errorf("%s: %w", op, ctx.Err())
	default:
		return fmt.Errorf("%s: failed to send message", op)
	}
}
