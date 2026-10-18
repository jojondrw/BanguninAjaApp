package project

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

var jpegBytes = []byte("\xFF\xD8\xFF\xE0\x00\x10JFIF\x00\x01\x01\x00\x00\x01\x00\x01\x00\x00")

type photoRepository struct {
	Repository
	created *SitePhoto
	covers  map[uuid.UUID]uuid.UUID
	asked   []uuid.UUID
}

func (r *photoRepository) FindProject(context.Context, uuid.UUID) (Project, error) {
	return Project{Status: statusOngoing}, nil
}

func (r *photoRepository) CreateSitePhoto(_ context.Context, photo *SitePhoto) error {
	r.created = photo
	return nil
}

func (r *photoRepository) ListProjects(context.Context, ProjectFilter) ([]Project, int64, error) {
	first, second := Project{}, Project{}
	first.ID, second.ID = uuid.New(), uuid.New()
	r.covers = map[uuid.UUID]uuid.UUID{first.ID: uuid.New()}
	return []Project{first, second}, 2, nil
}

func (r *photoRepository) LatestSitePhotos(_ context.Context, ids []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	r.asked = ids
	return r.covers, nil
}

func fixedService(repository Repository) *service {
	today := time.Date(2026, time.October, 18, 9, 0, 0, 0, time.UTC)
	return &service{repository: repository, now: func() time.Time { return today }}
}

func TestSitePhotoRejectsInvalidUploads(t *testing.T) {
	tomorrow := time.Date(2026, time.October, 19, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		upload SitePhotoUpload
		want   error
	}{
		{"empty", SitePhotoUpload{}, errSitePhotoEmpty},
		{"too large", SitePhotoUpload{Content: append(append([]byte{}, jpegBytes...), make([]byte, SitePhotoMaxBytes)...)}, errSitePhotoTooLarge},
		{"not an image", SitePhotoUpload{Content: []byte("%PDF-1.4 bukan foto")}, errSitePhotoType},
		{"future date", SitePhotoUpload{Content: jpegBytes, TakenOn: &tomorrow}, errSitePhotoFuture},
		{"long caption", SitePhotoUpload{Content: jpegBytes, Caption: string(bytes.Repeat([]byte("a"), sitePhotoCaptionMax+1))}, errSitePhotoCaption},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repository := &photoRepository{}
			_, err := fixedService(repository).CreateSitePhoto(context.Background(), uuid.New(), uuid.New(), tc.upload)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
			if repository.created != nil {
				t.Fatal("photo must not be stored")
			}
		})
	}
}

func TestSitePhotoStoresSniffedTypeAndDefaultsToToday(t *testing.T) {
	repository := &photoRepository{}
	actorID := uuid.New()

	response, err := fixedService(repository).CreateSitePhoto(context.Background(), uuid.New(), actorID, SitePhotoUpload{
		Content: jpegBytes,
		Caption: "  Pengecoran pelat lantai 2  ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.ContentType != "image/jpeg" || response.Size != len(jpegBytes) {
		t.Fatalf("got type %s size %d", response.ContentType, response.Size)
	}
	if response.Caption != "Pengecoran pelat lantai 2" {
		t.Fatalf("caption not trimmed: %q", response.Caption)
	}
	if !response.TakenOn.Equal(time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("taken on %v, want today", response.TakenOn)
	}
	if repository.created.CreatedBy == nil || *repository.created.CreatedBy != actorID {
		t.Fatal("photo must record its uploader")
	}
}

func TestListProjectsAttachesLatestPhotoAsCover(t *testing.T) {
	repository := &photoRepository{}

	page, err := fixedService(repository).ListProjects(context.Background(), uuid.New(), "admin", ProjectQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repository.asked) != 2 {
		t.Fatalf("asked covers for %d projects, want 2", len(repository.asked))
	}
	first, second := page.Items[0], page.Items[1]
	if first.CoverPhotoID == nil || *first.CoverPhotoID != repository.covers[first.ID] {
		t.Fatal("first project must carry its latest photo")
	}
	if second.CoverPhotoID != nil {
		t.Fatal("project without photos must have no cover")
	}
}
