package redpanda

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/IBM/sarama"
	"github.com/hesoyamTM/apphelper-notification/pkg/redpanda"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"go.uber.org/zap"
)

const (
	userRegisteredTopic     = "sso.auth.registered"
	passwordChangedTopic    = "sso.auth.password.changed"
	verificationCodeUpdated = "sso.auth.code.updated"
)

type RedPandaClient struct {
	producer    sarama.AsyncProducer
	messageChan chan *sarama.ProducerMessage
	stopChan    chan struct{}
}

func NewRedPandaClient(ctx context.Context, cfg redpanda.RedpandaConfig) (*RedPandaClient, error) {
	const op = "redpanda.NewRedPandaClient"

	redpandaCfg := redpanda.NewSaramaConfig(cfg)
	producer, err := redpanda.NewSaramaAsyncProducer(redpandaCfg, cfg.Brokers)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &RedPandaClient{
		producer:    *producer,
		messageChan: make(chan *sarama.ProducerMessage),
		stopChan:    make(chan struct{}),
	}, nil
}

func (c *RedPandaClient) Start(ctx context.Context) error {
	log := logger.GetLoggerFromCtx(ctx)

	var enqueued, successes, failures int32

	go func() {
		for range c.producer.Successes() {
			atomic.AddInt32(&successes, 1)
		}
	}()

	go func() {
		for err := range c.producer.Errors() {
			atomic.AddInt32(&failures, 1)
			log.Error(ctx, "failed to send message to redpanda", zap.Error(err))
		}
	}()

	for {
		select {
		case message := <-c.messageChan:
			select {
			case c.producer.Input() <- message:
				atomic.AddInt32(&enqueued, 1)
			case <-c.stopChan:
				return nil
			}
		case <-c.stopChan:
			return nil
		}
	}
}

func (c *RedPandaClient) Stop(ctx context.Context) error {
	const op = "redpanda.RedPandaClient.Stop"

	close(c.messageChan)
	close(c.stopChan)

	if err := c.producer.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	logger.GetLoggerFromCtx(ctx).Info(ctx, "redpanda client stopped")

	return nil
}
