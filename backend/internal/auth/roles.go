package auth

import (
	"context"
	"fmt"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
)

func EnsureRoles(ctx context.Context, repository Repository) (int, error) {
	definitions := access.Roles()
	rows := make([]Role, 0, len(definitions))
	for _, definition := range definitions {
		rows = append(rows, Role{Code: definition.Code, Name: definition.Name, Description: definition.Description})
	}

	if err := repository.UpsertRoles(ctx, rows); err != nil {
		return 0, fmt.Errorf("upsert roles: %w", err)
	}
	return len(rows), nil
}

func AssignAdminToUnassignedUsers(ctx context.Context, repository Repository) (int64, error) {
	assigned, err := repository.AssignRoleToUnassigned(ctx, access.RoleAdmin)
	if err != nil {
		return 0, fmt.Errorf("assign admin role to unassigned users: %w", err)
	}
	return assigned, nil
}
