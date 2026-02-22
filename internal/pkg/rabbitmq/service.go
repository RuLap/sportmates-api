package rabbitmq

import (
	"context"

	"github.com/RuLap/sportmates-api/internal/pkg/events"
)

type Service struct {
	mqClient *Client
}

func NewService(mqClient *Client) *Service {
	return &Service{mqClient: mqClient}
}

func (s *Service) PublishEmail(event events.EmailEvent) error {
	return s.mqClient.PublishEvent(event)
}

func (s *Service) ConsumeEmailEvents(ctx context.Context, handler func(events.EmailEvent) error) error {
	return s.mqClient.ConsumeEmailEvents(ctx, handler)
}
