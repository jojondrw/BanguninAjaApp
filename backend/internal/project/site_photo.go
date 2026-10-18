package project

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

const (
	SitePhotoMaxBytes   = 3 << 20
	sitePhotoListLimit  = 24
	sitePhotoCaptionMax = 160
)

var sitePhotoTypes = map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true}

var (
	errSitePhotoNotFound = apperror.NotFound("site_photo_not_found", "Foto lapangan tidak ditemukan")
	errSitePhotoEmpty    = apperror.Unprocessable("site_photo_empty", "Pilih foto yang akan diunggah")
	errSitePhotoTooLarge = apperror.Unprocessable("site_photo_too_large", "Ukuran foto paling besar 3 MB")
	errSitePhotoType     = apperror.Unprocessable("site_photo_type", "Foto harus berformat JPG, PNG, atau WebP")
	errSitePhotoCaption  = apperror.Unprocessable("site_photo_caption", "Keterangan foto paling panjang 160 karakter")
	errSitePhotoFuture   = apperror.Unprocessable("site_photo_in_future", "Tanggal foto tidak boleh setelah hari ini")

	sitePhotoErrors = database.ErrorMap{NotFound: errSitePhotoNotFound, Referenced: errProjectNotFound}
)

type SitePhoto struct {
	entity.Base
	ProjectID   uuid.UUID  `gorm:"type:uuid;not null;index:idx_site_photo_project"`
	TakenOn     time.Time  `gorm:"type:date;not null"`
	Caption     string     `gorm:"type:varchar(160);not null;default:''"`
	ContentType string     `gorm:"type:varchar(20);not null"`
	Size        int        `gorm:"not null"`
	Content     []byte     `gorm:"type:bytea;not null"`
	CreatedBy   *uuid.UUID `gorm:"type:uuid;index:idx_site_photo_created_by"`
}

func (SitePhoto) TableName() string {
	return "site_photo"
}

type SitePhotoUpload struct {
	Caption string
	TakenOn *time.Time
	Content []byte
}

type SitePhotoResponse struct {
	ID          uuid.UUID  `json:"id"`
	ProjectID   uuid.UUID  `json:"projectId"`
	TakenOn     time.Time  `json:"takenOn"`
	Caption     string     `json:"caption"`
	ContentType string     `json:"contentType"`
	Size        int        `json:"size"`
	CreatedBy   *uuid.UUID `json:"createdBy"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type SitePhotoFile struct {
	ContentType string
	Content     []byte
}

const sitePhotoColumns = "id, project_id, taken_on, caption, content_type, size, created_by, created_at, updated_at"

func newSitePhotoResponse(photo SitePhoto) SitePhotoResponse {
	return SitePhotoResponse{
		ID:          photo.ID,
		ProjectID:   photo.ProjectID,
		TakenOn:     photo.TakenOn,
		Caption:     photo.Caption,
		ContentType: photo.ContentType,
		Size:        photo.Size,
		CreatedBy:   photo.CreatedBy,
		CreatedAt:   photo.CreatedAt,
	}
}

func (s *service) ListSitePhotos(ctx context.Context, projectID uuid.UUID) ([]SitePhotoResponse, error) {
	if _, err := s.repository.FindProject(ctx, projectID); err != nil {
		return nil, projectReadErrors.Resolve(err)
	}
	photos, err := s.repository.ListSitePhotos(ctx, projectID, sitePhotoListLimit)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	responses := make([]SitePhotoResponse, len(photos))
	for index, photo := range photos {
		responses[index] = newSitePhotoResponse(photo)
	}
	return responses, nil
}

func (s *service) CreateSitePhoto(ctx context.Context, projectID, actorID uuid.UUID, upload SitePhotoUpload) (SitePhotoResponse, error) {
	contentType, err := s.validateSitePhoto(upload)
	if err != nil {
		return SitePhotoResponse{}, err
	}
	if _, err := s.repository.FindProject(ctx, projectID); err != nil {
		return SitePhotoResponse{}, projectReadErrors.Resolve(err)
	}

	takenOn := s.today()
	if upload.TakenOn != nil {
		takenOn = dateOnly(*upload.TakenOn)
	}
	photo := SitePhoto{
		ProjectID:   projectID,
		TakenOn:     takenOn,
		Caption:     strings.TrimSpace(upload.Caption),
		ContentType: contentType,
		Size:        len(upload.Content),
		Content:     upload.Content,
		CreatedBy:   &actorID,
	}
	if err := s.repository.CreateSitePhoto(ctx, &photo); err != nil {
		return SitePhotoResponse{}, sitePhotoErrors.Resolve(err)
	}
	return newSitePhotoResponse(photo), nil
}

func (s *service) SitePhotoFile(ctx context.Context, projectID, photoID uuid.UUID) (SitePhotoFile, error) {
	photo, err := s.repository.FindSitePhoto(ctx, projectID, photoID)
	if err != nil {
		return SitePhotoFile{}, sitePhotoErrors.Resolve(err)
	}
	return SitePhotoFile{ContentType: photo.ContentType, Content: photo.Content}, nil
}

func (s *service) DeleteSitePhoto(ctx context.Context, projectID, photoID uuid.UUID) error {
	return sitePhotoErrors.Resolve(s.repository.DeleteSitePhoto(ctx, projectID, photoID))
}

func (s *service) validateSitePhoto(upload SitePhotoUpload) (string, error) {
	if len(upload.Content) == 0 {
		return "", errSitePhotoEmpty
	}
	if len(upload.Content) > SitePhotoMaxBytes {
		return "", errSitePhotoTooLarge
	}
	if len([]rune(strings.TrimSpace(upload.Caption))) > sitePhotoCaptionMax {
		return "", errSitePhotoCaption
	}
	if upload.TakenOn != nil && dateOnly(*upload.TakenOn).After(s.today()) {
		return "", errSitePhotoFuture
	}
	contentType := http.DetectContentType(upload.Content)
	if !sitePhotoTypes[contentType] {
		return "", errSitePhotoType
	}
	return contentType, nil
}

func (s *service) withCoverPhotos(ctx context.Context, projects []ProjectResponse) ([]ProjectResponse, error) {
	if len(projects) == 0 {
		return projects, nil
	}
	ids := make([]uuid.UUID, len(projects))
	for index, project := range projects {
		ids[index] = project.ID
	}
	covers, err := s.repository.LatestSitePhotos(ctx, ids)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	for index := range projects {
		if cover, ok := covers[projects[index].ID]; ok {
			projects[index].CoverPhotoID = &cover
		}
	}
	return projects, nil
}

func (r *gormRepository) ListSitePhotos(ctx context.Context, projectID uuid.UUID, limit int) ([]SitePhoto, error) {
	var photos []SitePhoto
	err := r.db.WithContext(ctx).
		Select(sitePhotoColumns).
		Where("project_id = ?", projectID).
		Order("taken_on DESC, created_at DESC").
		Limit(limit).
		Find(&photos).Error
	return photos, database.Translate(err)
}

func (r *gormRepository) LatestSitePhotos(ctx context.Context, projectIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	var rows []struct {
		ProjectID uuid.UUID
		ID        uuid.UUID
	}
	err := r.db.WithContext(ctx).
		Model(&SitePhoto{}).
		Select("DISTINCT ON (project_id) project_id, id").
		Where("project_id IN ?", projectIDs).
		Order("project_id, taken_on DESC, created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, database.Translate(err)
	}
	covers := make(map[uuid.UUID]uuid.UUID, len(rows))
	for _, row := range rows {
		covers[row.ProjectID] = row.ID
	}
	return covers, nil
}

func (r *gormRepository) FindSitePhoto(ctx context.Context, projectID, id uuid.UUID) (SitePhoto, error) {
	var photo SitePhoto
	err := r.db.WithContext(ctx).Where("id = ? AND project_id = ?", id, projectID).Take(&photo).Error
	return photo, database.Translate(err)
}

func (r *gormRepository) CreateSitePhoto(ctx context.Context, photo *SitePhoto) error {
	return database.Translate(r.db.WithContext(ctx).Create(photo).Error)
}

func (r *gormRepository) DeleteSitePhoto(ctx context.Context, projectID, id uuid.UUID) error {
	return r.deleteWhere(ctx, &SitePhoto{}, "id = ? AND project_id = ?", id, projectID)
}
