package news

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
)

const (
	freshFor         = 15 * time.Minute
	maxCachedQueries = 256
)

var errNewsUnavailable = apperror.New(502, "news_unavailable", "Berita sekitar lokasi belum bisa diambil, coba lagi nanti")

type Service interface {
	FetchNews(ctx context.Context, query NewsQuery) ([]NewsResponse, error)
}

type Fetcher interface {
	Fetch(ctx context.Context, region, district string) ([]NewsResponse, error)
}

type cachedArticles struct {
	articles  []NewsResponse
	fetchedAt time.Time
}

type service struct {
	client Fetcher
	now    func() time.Time

	mu    sync.Mutex
	cache map[string]cachedArticles
}

func NewService(client Fetcher) Service {
	return &service{
		client: client,
		now:    time.Now,
		cache:  map[string]cachedArticles{},
	}
}

func (s *service) FetchNews(ctx context.Context, query NewsQuery) ([]NewsResponse, error) {
	key := cacheKey(query)
	if articles, ok := s.cached(key); ok {
		return articles, nil
	}

	articles, err := s.client.Fetch(ctx, query.Region, query.District)
	if err != nil {
		return nil, errNewsUnavailable.WithCause(err)
	}

	s.remember(key, articles)
	return copyArticles(articles), nil
}

func (s *service) cached(key string) ([]NewsResponse, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.cache[key]
	if !ok || s.now().Sub(entry.fetchedAt) >= freshFor {
		return nil, false
	}
	return copyArticles(entry.articles), true
}

func (s *service) remember(key string, articles []NewsResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if len(s.cache) >= maxCachedQueries {
		for cachedKey, entry := range s.cache {
			if now.Sub(entry.fetchedAt) >= freshFor {
				delete(s.cache, cachedKey)
			}
		}
	}
	if len(s.cache) >= maxCachedQueries {
		return
	}
	s.cache[key] = cachedArticles{articles: copyArticles(articles), fetchedAt: now}
}

func cacheKey(query NewsQuery) string {
	return strings.ToLower(strings.TrimSpace(query.Region)) + "|" + strings.ToLower(strings.TrimSpace(query.District))
}

func copyArticles(articles []NewsResponse) []NewsResponse {
	return append([]NewsResponse{}, articles...)
}
