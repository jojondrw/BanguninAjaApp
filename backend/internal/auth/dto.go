package auth

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=120"`
	Email    string `json:"email" binding:"required,email,max=160"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=160"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required,max=72"`
	NewPassword     string `json:"newPassword" binding:"required,min=8,max=72"`
}

type RoleResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type UserResponse struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Email       string            `json:"email"`
	Active      bool              `json:"active"`
	Role        *RoleResponse     `json:"role"`
	Permissions map[string]string `json:"permissions"`
	CreatedAt   time.Time         `json:"createdAt"`
}

const pendingRoleFilter = "none"

type UserQuery struct {
	pagination.Query
	Search string `form:"search" binding:"omitempty,max=160"`
	Role   string `form:"role" binding:"omitempty,max=30"`
	Active *bool  `form:"active"`
}

type ChangeRoleRequest struct {
	Role string `json:"role" binding:"required,max=30"`
}

type ChangeActiveRequest struct {
	Active *bool `json:"active" binding:"required"`
}

type ManagedUserResponse struct {
	ID        uuid.UUID     `json:"id"`
	Name      string        `json:"name"`
	Email     string        `json:"email"`
	Active    bool          `json:"active"`
	Role      *RoleResponse `json:"role"`
	CreatedAt time.Time     `json:"createdAt"`
}

type SessionResponse struct {
	AccessToken          string       `json:"accessToken"`
	AccessTokenExpiresAt time.Time    `json:"accessTokenExpiresAt"`
	User                 UserResponse `json:"user"`
}

type Session struct {
	Response         SessionResponse
	RefreshToken     string
	RefreshExpiresAt time.Time
}

func newUserResponse(user User, role *Role) UserResponse {
	response := UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Active:      user.Active,
		Permissions: map[string]string{},
		CreatedAt:   user.CreatedAt,
	}
	if role != nil {
		response.Role = &RoleResponse{Code: role.Code, Name: role.Name}
		response.Permissions = access.Permissions(role.Code)
	}
	return response
}

func newManagedUserResponse(row UserRow) ManagedUserResponse {
	response := ManagedUserResponse{
		ID:        row.ID,
		Name:      row.Name,
		Email:     row.Email,
		Active:    row.Active,
		CreatedAt: row.CreatedAt,
	}
	if row.RoleCode != nil {
		response.Role = &RoleResponse{Code: *row.RoleCode, Name: valueOf(row.RoleName)}
	}
	return response
}

func valueOf(text *string) string {
	if text == nil {
		return ""
	}
	return *text
}
