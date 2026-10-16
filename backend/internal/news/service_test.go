package news

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
)

type countingFetcher struct {
	calls int
	err   error
}

func (f *countingFetcher) Fetch(_ context.Context, region, district string) ([]NewsResponse, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return []NewsResponse{{Title: "Berita " + region + district}}, nil
}

func newTestService(fetcher Fetcher, clock *time.Time) *service {
	news := NewService(fetcher).(*service)
	news.now = func() time.Time { return *clock }
	return news
}

func TestFetchNewsReusesArticlesForTheSameAreaWhileFresh(t *testing.T) {
	clock := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	fetcher := &countingFetcher{}
	news := newTestService(fetcher, &clock)

	first, err := news.FetchNews(context.Background(), NewsQuery{Region: "Kota Bandung", District: "Coblong"})
	if err != nil {
		t.Fatalf("first fetch: %v", err)
	}
	clock = clock.Add(14 * time.Minute)
	second, err := news.FetchNews(context.Background(), NewsQuery{Region: " kota bandung ", District: "COBLONG"})
	if err != nil {
		t.Fatalf("second fetch: %v", err)
	}

	if fetcher.calls != 1 {
		t.Fatalf("Google News was called %d times for the same area within 15 minutes", fetcher.calls)
	}
	if len(second) != 1 || second[0].Title != first[0].Title {
		t.Fatalf("cached articles = %+v, want %+v", second, first)
	}

	clock = clock.Add(time.Minute)
	if _, err := news.FetchNews(context.Background(), NewsQuery{Region: "Kota Bandung", District: "Coblong"}); err != nil {
		t.Fatalf("fetch after expiry: %v", err)
	}
	if fetcher.calls != 2 {
		t.Fatalf("Google News was called %d times, want a fresh call after 15 minutes", fetcher.calls)
	}
}

func TestFetchNewsKeepsDifferentAreasApart(t *testing.T) {
	clock := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	fetcher := &countingFetcher{}
	news := newTestService(fetcher, &clock)

	bandung, _ := news.FetchNews(context.Background(), NewsQuery{Region: "Kota Bandung"})
	surabaya, _ := news.FetchNews(context.Background(), NewsQuery{Region: "Kota Surabaya"})

	if fetcher.calls != 2 || bandung[0].Title == surabaya[0].Title {
		t.Fatalf("calls = %d, articles %q and %q", fetcher.calls, bandung[0].Title, surabaya[0].Title)
	}
}

func TestFetchNewsDoesNotCacheFailuresAndReportsThemAsUnavailable(t *testing.T) {
	clock := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	fetcher := &countingFetcher{err: errors.New("google news returned status 503")}
	news := newTestService(fetcher, &clock)

	_, err := news.FetchNews(context.Background(), NewsQuery{Region: "Kota Surabaya"})
	var appError *apperror.Error
	if !errors.As(err, &appError) || appError.Status != 502 || appError.Code != "news_unavailable" {
		t.Fatalf("err = %v, want 502 news_unavailable", err)
	}

	fetcher.err = nil
	articles, err := news.FetchNews(context.Background(), NewsQuery{Region: "Kota Surabaya"})
	if err != nil || len(articles) != 1 {
		t.Fatalf("retry after a failure = %+v, %v", articles, err)
	}
	if fetcher.calls != 2 {
		t.Fatalf("Google News was called %d times, want a retry after the failure", fetcher.calls)
	}
}

func TestFetchNewsReturnsACopyThatCannotChangeTheCache(t *testing.T) {
	clock := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	news := newTestService(&countingFetcher{}, &clock)

	first, _ := news.FetchNews(context.Background(), NewsQuery{Region: "Kota Bandung"})
	first[0].Title = "diubah pemanggil"
	second, _ := news.FetchNews(context.Background(), NewsQuery{Region: "Kota Bandung"})

	if second[0].Title == "diubah pemanggil" {
		t.Fatal("a caller changed the cached articles")
	}
}

func TestFetchNewsStopsCachingNewAreasWhenFullOfFreshEntries(t *testing.T) {
	clock := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	news := newTestService(&countingFetcher{}, &clock)

	for i := 0; i < maxCachedQueries+10; i++ {
		_, _ = news.FetchNews(context.Background(), NewsQuery{Region: "Wilayah", District: string(rune('A'+i%26)) + time.Duration(i).String()})
	}
	if len(news.cache) > maxCachedQueries {
		t.Fatalf("cache holds %d areas, limit is %d", len(news.cache), maxCachedQueries)
	}

	clock = clock.Add(freshFor)
	_, _ = news.FetchNews(context.Background(), NewsQuery{Region: "Wilayah baru"})
	if len(news.cache) != 1 {
		t.Fatalf("cache holds %d areas after every old entry expired, want only the new one", len(news.cache))
	}
}
