package project

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type ProjectRequest struct {
	Code          string     `json:"code" binding:"required,max=20"`
	Name          string     `json:"name" binding:"required,max=160"`
	RegionID      *uuid.UUID `json:"regionId"`
	Type          string     `json:"type" binding:"required,max=40"`
	StartDate     *time.Time `json:"startDate"`
	TargetEndDate *time.Time `json:"targetEndDate"`
	ContractValue int64      `json:"contractValue" binding:"min=0"`
}

type ProjectStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=planning ongoing on_hold completed cancelled"`
}

type ProjectQuery struct {
	pagination.Query
	Search   string     `form:"search" binding:"omitempty,max=160"`
	Status   string     `form:"status" binding:"omitempty,oneof=planning ongoing on_hold completed cancelled"`
	RegionID *uuid.UUID `form:"regionId,parser=encoding.TextUnmarshaler"`
}

type ProjectResponse struct {
	ID            uuid.UUID  `json:"id"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	RegionID      *uuid.UUID `json:"regionId"`
	Type          string     `json:"type"`
	Status        string     `json:"status"`
	Progress      int        `json:"progress"`
	StartDate     *time.Time `json:"startDate"`
	TargetEndDate *time.Time `json:"targetEndDate"`
	ContractValue int64      `json:"contractValue"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type PhaseRequest struct {
	Name          string     `json:"name" binding:"required,max=160"`
	SortOrder     int        `json:"sortOrder" binding:"required,min=1"`
	StartDate     *time.Time `json:"startDate"`
	TargetEndDate *time.Time `json:"targetEndDate"`
	Progress      int        `json:"progress" binding:"min=0,max=100"`
}

type PhaseResponse struct {
	ID            uuid.UUID  `json:"id"`
	ProjectID     uuid.UUID  `json:"projectId"`
	Name          string     `json:"name"`
	SortOrder     int        `json:"sortOrder"`
	StartDate     *time.Time `json:"startDate"`
	TargetEndDate *time.Time `json:"targetEndDate"`
	Progress      int        `json:"progress"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type BudgetItemRequest struct {
	Code            string    `json:"code" binding:"required,max=20"`
	Description     string    `json:"description" binding:"required,max=200"`
	Volume          float64   `json:"volume" binding:"min=0"`
	UnitOfMeasureID uuid.UUID `json:"unitOfMeasureId" binding:"required"`
	UnitPrice       int64     `json:"unitPrice" binding:"min=0"`
}

type BudgetItemQuery struct {
	pagination.Query
	Search string `form:"search" binding:"omitempty,max=200"`
}

type BudgetItemResponse struct {
	ID              uuid.UUID `json:"id"`
	ProjectID       uuid.UUID `json:"projectId"`
	Code            string    `json:"code"`
	Description     string    `json:"description"`
	Volume          float64   `json:"volume"`
	UnitOfMeasureID uuid.UUID `json:"unitOfMeasureId"`
	UnitPrice       int64     `json:"unitPrice"`
	Total           int64     `json:"total"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type BudgetItemPage struct {
	pagination.Page[BudgetItemResponse]
	GrandTotal int64 `json:"grandTotal"`
}

type PermitRequest struct {
	Type       string     `json:"type" binding:"required,max=60"`
	Number     string     `json:"number" binding:"required,max=80"`
	IssuedDate *time.Time `json:"issuedDate"`
	ValidUntil *time.Time `json:"validUntil"`
	Status     string     `json:"status" binding:"required,oneof=submitted issued expired rejected"`
}

type PermitResponse struct {
	ID         uuid.UUID  `json:"id"`
	ProjectID  uuid.UUID  `json:"projectId"`
	Type       string     `json:"type"`
	Number     string     `json:"number"`
	IssuedDate *time.Time `json:"issuedDate"`
	ValidUntil *time.Time `json:"validUntil"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

func newProjectResponse(project Project) ProjectResponse {
	return ProjectResponse{
		ID:            project.ID,
		Code:          project.Code,
		Name:          project.Name,
		RegionID:      project.RegionID,
		Type:          project.Type,
		Status:        project.Status,
		Progress:      project.Progress,
		StartDate:     project.StartDate,
		TargetEndDate: project.TargetEndDate,
		ContractValue: project.ContractValue,
		CreatedAt:     project.CreatedAt,
		UpdatedAt:     project.UpdatedAt,
	}
}

func newPhaseResponse(phase ProjectPhase) PhaseResponse {
	return PhaseResponse{
		ID:            phase.ID,
		ProjectID:     phase.ProjectID,
		Name:          phase.Name,
		SortOrder:     phase.SortOrder,
		StartDate:     phase.StartDate,
		TargetEndDate: phase.TargetEndDate,
		Progress:      phase.Progress,
		Status:        phase.Status,
		CreatedAt:     phase.CreatedAt,
		UpdatedAt:     phase.UpdatedAt,
	}
}

func newBudgetItemResponse(item BudgetItem) BudgetItemResponse {
	return BudgetItemResponse{
		ID:              item.ID,
		ProjectID:       item.ProjectID,
		Code:            item.Code,
		Description:     item.Description,
		Volume:          item.Volume,
		UnitOfMeasureID: item.UnitOfMeasureID,
		UnitPrice:       item.UnitPrice,
		Total:           item.Total,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func newPermitResponse(permit Permit) PermitResponse {
	return PermitResponse{
		ID:         permit.ID,
		ProjectID:  permit.ProjectID,
		Type:       permit.Type,
		Number:     permit.Number,
		IssuedDate: permit.IssuedDate,
		ValidUntil: permit.ValidUntil,
		Status:     permit.Status,
		CreatedAt:  permit.CreatedAt,
		UpdatedAt:  permit.UpdatedAt,
	}
}
