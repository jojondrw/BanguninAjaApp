package privacy

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
)

func TestMaskIdentityKeepsOnlyTheLastFourCharacters(t *testing.T) {
	cases := map[string]string{
		"3273011504900003": "••••••••••••0003",
		"DEMO-KTP-0003":    "•••••••••0003",
		"1234":             "••••",
		"":                 "",
	}
	for value, want := range cases {
		if got := MaskIdentity(value); got != want {
			t.Errorf("MaskIdentity(%q) = %q, want %q", value, got, want)
		}
	}
}

func TestOnlyWritersOfTheModuleSeeTheFullIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		role string
		want bool
	}{
		{access.RoleAdmin, true},
		{access.RoleSales, true},
		{access.RoleObserver, false},
		{access.RoleFinance, false},
		{"", false},
	}
	for _, tc := range cases {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		if tc.role != "" {
			ctx.Set("authenticated_role", tc.role)
		}
		if got := RevealsIdentity(ctx, access.ModuleSales); got != tc.want {
			t.Errorf("role %q: got %v, want %v", tc.role, got, tc.want)
		}
	}
}
