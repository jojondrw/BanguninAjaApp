package auth

import (
	"time"

	"github.com/google/uuid"
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

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
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

func newUserResponse(user User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}
