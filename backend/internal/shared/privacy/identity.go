package privacy

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
)

const (
	visibleIdentityDigits = 4
	identityMask          = "•"
)

func MaskIdentity(value string) string {
	runes := []rune(value)
	if len(runes) <= visibleIdentityDigits {
		return strings.Repeat(identityMask, len(runes))
	}
	hidden := len(runes) - visibleIdentityDigits
	return strings.Repeat(identityMask, hidden) + string(runes[hidden:])
}

func RevealsIdentity(ctx *gin.Context, module access.Module) bool {
	role, _ := middleware.AuthenticatedRole(ctx)
	return access.Allows(role, module, access.LevelWrite)
}
