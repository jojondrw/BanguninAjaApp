package news

import (
	"context"
)

type Service interface {
	FetchNews(ctx context.Context, query NewsQuery) ([]NewsResponse, error)
}

type service struct {
	client *Client
}

func NewService(client *Client) Service {
	return &service{
		client: client,
	}
}

func (s *service) FetchNews(ctx context.Context, query NewsQuery) ([]NewsResponse, error) {
	return s.client.Fetch(ctx, query.Region, query.District)
}