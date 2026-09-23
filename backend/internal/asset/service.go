package asset

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

const (
	assetDisposed      = "disposed"
	equipmentOperating = "operating"
)

var (
	errAssetNotFound            = apperror.NotFound("asset_not_found", "Aset tidak ditemukan")
	errAssetCodeUsed            = apperror.Conflict("asset_code_used", "Kode aset sudah dipakai")
	errAssetDepreciationTooHigh = apperror.Unprocessable("asset_depreciation_exceeds_cost", "Akumulasi penyusutan tidak boleh melebihi nilai perolehan")

	errEquipmentNotFound        = apperror.NotFound("equipment_not_found", "Alat tidak ditemukan")
	errEquipmentCodeUsed        = apperror.Conflict("equipment_code_used", "Kode alat sudah dipakai")
	errEquipmentReference       = apperror.Unprocessable("equipment_reference_not_found", "Aset atau proyek tidak ditemukan")
	errEquipmentProjectRequired = apperror.Unprocessable("equipment_project_required", "Alat yang sedang beroperasi wajib ditugaskan ke proyek")
	errEquipmentAssetDisposed   = apperror.Unprocessable("equipment_asset_disposed", "Aset induk alat sudah dilepas")
)

var (
	assetReadErrors  = database.ErrorMap{NotFound: errAssetNotFound}
	assetWriteErrors = database.ErrorMap{NotFound: errAssetNotFound, Duplicate: errAssetCodeUsed, Invalid: errAssetDepreciationTooHigh}

	equipmentReadErrors   = database.ErrorMap{NotFound: errEquipmentNotFound}
	equipmentWriteErrors  = database.ErrorMap{NotFound: errEquipmentNotFound, Duplicate: errEquipmentCodeUsed, Referenced: errEquipmentReference}
	equipmentParentErrors = database.ErrorMap{NotFound: errEquipmentReference}
)

type Service interface {
	ListAssets(ctx context.Context, query AssetQuery) (AssetPage, error)
	GetAsset(ctx context.Context, id uuid.UUID) (AssetResponse, error)
	CreateAsset(ctx context.Context, request AssetRequest) (AssetResponse, error)
	UpdateAsset(ctx context.Context, id uuid.UUID, request AssetRequest) (AssetResponse, error)
	DeleteAsset(ctx context.Context, id uuid.UUID) error

	ListEquipment(ctx context.Context, query EquipmentQuery) (pagination.Page[EquipmentResponse], error)
	GetEquipment(ctx context.Context, id uuid.UUID) (EquipmentResponse, error)
	CreateEquipment(ctx context.Context, request EquipmentRequest) (EquipmentResponse, error)
	UpdateEquipment(ctx context.Context, id uuid.UUID, request EquipmentRequest) (EquipmentResponse, error)
	DeleteEquipment(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) ListAssets(ctx context.Context, query AssetQuery) (AssetPage, error) {
	filter := AssetFilter{
		Search:   query.Search,
		Category: strings.TrimSpace(query.Category),
		Status:   query.Status,
		Offset:   query.Offset(),
		Limit:    query.Size(),
	}
	assets, total, err := s.repository.ListAssets(ctx, filter)
	if err != nil {
		return AssetPage{}, apperror.Internal(err)
	}

	totals, err := s.repository.SumAssets(ctx, filter)
	if err != nil {
		return AssetPage{}, apperror.Internal(err)
	}
	return AssetPage{
		Page:                         pagination.New(pagination.Map(assets, newAssetResponse), query.Query, total),
		TotalAcquisitionValue:        totals.AcquisitionValue,
		TotalAccumulatedDepreciation: totals.AccumulatedDepreciation,
		TotalBookValue:               bookValue(totals.AcquisitionValue, totals.AccumulatedDepreciation),
	}, nil
}

func (s *service) GetAsset(ctx context.Context, id uuid.UUID) (AssetResponse, error) {
	asset, err := s.repository.FindAsset(ctx, id)
	if err != nil {
		return AssetResponse{}, assetReadErrors.Resolve(err)
	}
	return newAssetResponse(asset), nil
}

func (s *service) CreateAsset(ctx context.Context, request AssetRequest) (AssetResponse, error) {
	if err := validateAsset(request); err != nil {
		return AssetResponse{}, err
	}

	var asset Asset
	applyAssetRequest(&asset, request)
	if err := s.repository.CreateAsset(ctx, &asset); err != nil {
		return AssetResponse{}, assetWriteErrors.Resolve(err)
	}
	return newAssetResponse(asset), nil
}

func (s *service) UpdateAsset(ctx context.Context, id uuid.UUID, request AssetRequest) (AssetResponse, error) {
	if err := validateAsset(request); err != nil {
		return AssetResponse{}, err
	}

	asset, err := s.repository.FindAsset(ctx, id)
	if err != nil {
		return AssetResponse{}, assetReadErrors.Resolve(err)
	}

	applyAssetRequest(&asset, request)
	if err := s.repository.SaveAsset(ctx, &asset); err != nil {
		return AssetResponse{}, assetWriteErrors.Resolve(err)
	}
	return newAssetResponse(asset), nil
}

func (s *service) DeleteAsset(ctx context.Context, id uuid.UUID) error {
	return assetReadErrors.Resolve(s.repository.DeleteAsset(ctx, id))
}

func (s *service) ListEquipment(ctx context.Context, query EquipmentQuery) (pagination.Page[EquipmentResponse], error) {
	equipment, total, err := s.repository.ListEquipment(ctx, EquipmentFilter{
		Search:           query.Search,
		Status:           query.Status,
		ProjectID:        query.ProjectID,
		AssetID:          query.AssetID,
		ServiceDueBefore: query.ServiceDueBefore,
		Offset:           query.Offset(),
		Limit:            query.Size(),
	})
	if err != nil {
		return pagination.Page[EquipmentResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(equipment, newEquipmentResponse), query.Query, total), nil
}

func (s *service) GetEquipment(ctx context.Context, id uuid.UUID) (EquipmentResponse, error) {
	equipment, err := s.repository.FindEquipment(ctx, id)
	if err != nil {
		return EquipmentResponse{}, equipmentReadErrors.Resolve(err)
	}
	return newEquipmentResponse(equipment), nil
}

func (s *service) CreateEquipment(ctx context.Context, request EquipmentRequest) (EquipmentResponse, error) {
	if err := s.validateEquipment(ctx, request); err != nil {
		return EquipmentResponse{}, err
	}

	var equipment Equipment
	applyEquipmentRequest(&equipment, request)
	if err := s.repository.CreateEquipment(ctx, &equipment); err != nil {
		return EquipmentResponse{}, equipmentWriteErrors.Resolve(err)
	}
	return newEquipmentResponse(equipment), nil
}

func (s *service) UpdateEquipment(ctx context.Context, id uuid.UUID, request EquipmentRequest) (EquipmentResponse, error) {
	equipment, err := s.repository.FindEquipment(ctx, id)
	if err != nil {
		return EquipmentResponse{}, equipmentReadErrors.Resolve(err)
	}
	if err := s.validateEquipment(ctx, request); err != nil {
		return EquipmentResponse{}, err
	}

	applyEquipmentRequest(&equipment, request)
	if err := s.repository.SaveEquipment(ctx, &equipment); err != nil {
		return EquipmentResponse{}, equipmentWriteErrors.Resolve(err)
	}
	return newEquipmentResponse(equipment), nil
}

func (s *service) DeleteEquipment(ctx context.Context, id uuid.UUID) error {
	return equipmentReadErrors.Resolve(s.repository.DeleteEquipment(ctx, id))
}

func (s *service) validateEquipment(ctx context.Context, request EquipmentRequest) error {
	if request.Status == equipmentOperating && request.ProjectID == nil {
		return errEquipmentProjectRequired
	}
	if request.AssetID == nil {
		return nil
	}

	parent, err := s.repository.FindAsset(ctx, *request.AssetID)
	if err != nil {
		return equipmentParentErrors.Resolve(err)
	}
	if parent.Status == assetDisposed {
		return errEquipmentAssetDisposed
	}
	return nil
}

func validateAsset(request AssetRequest) error {
	if request.AccumulatedDepreciation > request.AcquisitionValue {
		return errAssetDepreciationTooHigh
	}
	return nil
}

func bookValue(acquisitionValue, accumulatedDepreciation int64) int64 {
	return acquisitionValue - accumulatedDepreciation
}

func monthlyDepreciation(acquisitionValue int64, usefulLifeMonths int) int64 {
	if usefulLifeMonths <= 0 {
		return 0
	}
	return acquisitionValue / int64(usefulLifeMonths)
}

func applyAssetRequest(asset *Asset, request AssetRequest) {
	asset.Code = strings.TrimSpace(request.Code)
	asset.Name = strings.TrimSpace(request.Name)
	asset.Category = strings.TrimSpace(request.Category)
	asset.AcquisitionDate = request.AcquisitionDate
	asset.AcquisitionValue = request.AcquisitionValue
	asset.AccumulatedDepreciation = request.AccumulatedDepreciation
	asset.UsefulLifeMonths = request.UsefulLifeMonths
	asset.Status = request.Status
}

func applyEquipmentRequest(equipment *Equipment, request EquipmentRequest) {
	equipment.Code = strings.TrimSpace(request.Code)
	equipment.Name = strings.TrimSpace(request.Name)
	equipment.AssetID = request.AssetID
	equipment.ProjectID = request.ProjectID
	equipment.OperatingHours = request.OperatingHours
	equipment.NextServiceDate = request.NextServiceDate
	equipment.Status = request.Status
}
