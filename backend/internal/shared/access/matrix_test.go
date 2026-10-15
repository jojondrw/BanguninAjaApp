package access

import (
	"net/http"
	"testing"
)

func TestModuleForPathMapsEveryApiPrefix(t *testing.T) {
	cases := map[string]Module{
		"/api/projects/:id/phases":                 ModuleProjects,
		"/api/finance/budgets":                     ModuleFinance,
		"/api/billing/invoices/:id/payments":       ModuleBilling,
		"/api/sales/contracts":                     ModuleSales,
		"/api/procurement/purchase-orders/summary": ModuleProcurement,
		"/api/inventory/materials":                 ModuleInventory,
		"/api/assets/:id":                          ModuleInventory,
		"/api/equipment":                           ModuleInventory,
		"/api/hr/payrolls/:id/pay":                 ModuleHR,
		"/api/locations/saved":                     ModuleLocations,
		"/api/site/evaluate":                       ModuleLocations,
		"/api/scoring/dimensions":                  ModuleLocations,
		"/api/regulation/lookup":                   ModuleLocations,
		"/api/news":                                ModuleLocations,
		"/api/master/accounts":                     ModuleMaster,
		"/api/reports":                             ModuleReports,
		"/api/reports/:id/download":                ModuleReports,
		"/api/users/:id/role":                      ModuleUsers,
		"/api/brand-new-module":                    ModuleUnmapped,
	}
	for path, want := range cases {
		module, guarded := ModuleForPath(path)
		if !guarded || module != want {
			t.Errorf("%s: got %q guarded=%v, want %q", path, module, guarded, want)
		}
	}
}

func TestModuleForPathLeavesAuthOpen(t *testing.T) {
	for _, path := range []string{"/api/auth/login", "/api/auth/me", "/api/auth/refresh", "/health", "/api"} {
		if _, guarded := ModuleForPath(path); guarded {
			t.Errorf("%s should stay open", path)
		}
	}
}

func TestModuleForPathDoesNotMatchPartialSegments(t *testing.T) {
	if module, _ := ModuleForPath("/api/projectsx"); module != ModuleUnmapped {
		t.Fatalf("got %q", module)
	}
}

func TestLevelForMethod(t *testing.T) {
	reads := []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	writes := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
	for _, method := range reads {
		if LevelForMethod(method) != LevelRead {
			t.Errorf("%s should be read", method)
		}
	}
	for _, method := range writes {
		if LevelForMethod(method) != LevelWrite {
			t.Errorf("%s should be write", method)
		}
	}
}

type grantCase struct {
	role    string
	module  Module
	level   Level
	allowed bool
}

func TestRequestedPermissionMatrix(t *testing.T) {
	cases := []grantCase{
		{RoleAdmin, ModuleUsers, LevelWrite, true},
		{RoleAdmin, ModuleHR, LevelWrite, true},
		{RoleAdmin, ModuleUnmapped, LevelWrite, true},

		{RoleProjectManager, ModuleProjects, LevelWrite, true},
		{RoleProjectManager, ModuleLocations, LevelWrite, true},
		{RoleProjectManager, ModuleInventory, LevelWrite, true},
		{RoleProjectManager, ModuleProcurement, LevelRead, true},
		{RoleProjectManager, ModuleProcurement, LevelWrite, false},
		{RoleProjectManager, ModuleHR, LevelRead, false},
		{RoleProjectManager, ModuleUsers, LevelRead, false},

		{RoleFinance, ModuleFinance, LevelWrite, true},
		{RoleFinance, ModuleBilling, LevelWrite, true},
		{RoleFinance, ModuleProjects, LevelRead, true},
		{RoleFinance, ModuleProjects, LevelWrite, false},
		{RoleFinance, ModuleSales, LevelRead, true},
		{RoleFinance, ModuleSales, LevelWrite, false},
		{RoleFinance, ModuleProcurement, LevelRead, true},
		{RoleFinance, ModuleProcurement, LevelWrite, false},
		{RoleFinance, ModuleHR, LevelRead, false},

		{RoleProcurement, ModuleProcurement, LevelWrite, true},
		{RoleProcurement, ModuleInventory, LevelWrite, true},
		{RoleProcurement, ModuleFinance, LevelRead, false},
		{RoleProcurement, ModuleSales, LevelRead, false},

		{RoleSales, ModuleSales, LevelWrite, true},
		{RoleSales, ModuleProjects, LevelRead, true},
		{RoleSales, ModuleBilling, LevelRead, true},
		{RoleSales, ModuleBilling, LevelWrite, false},
		{RoleSales, ModuleFinance, LevelRead, false},

		{RoleHR, ModuleHR, LevelWrite, true},
		{RoleHR, ModuleFinance, LevelRead, false},
		{RoleHR, ModuleSales, LevelRead, false},

		{RoleObserver, ModuleFinance, LevelRead, true},
		{RoleObserver, ModuleHR, LevelRead, true},
		{RoleObserver, ModuleReports, LevelRead, true},
		{RoleObserver, ModuleFinance, LevelWrite, false},
		{RoleObserver, ModuleUsers, LevelRead, false},
		{RoleObserver, ModuleUnmapped, LevelRead, false},

		{"", ModuleProjects, LevelRead, false},
		{"tamu", ModuleProjects, LevelRead, false},
	}
	for _, current := range cases {
		if got := Allows(current.role, current.module, current.level); got != current.allowed {
			t.Errorf("%s %s %s: got %v, want %v", current.role, current.module, current.level, got, current.allowed)
		}
	}
}

func TestObserverReadsEveryModuleButUsers(t *testing.T) {
	for _, module := range Modules() {
		want := module != ModuleUsers
		if got := Allows(RoleObserver, module, LevelRead); got != want {
			t.Errorf("%s: got %v, want %v", module, got, want)
		}
		if Allows(RoleObserver, module, LevelWrite) {
			t.Errorf("observer must not write %s", module)
		}
	}
}

func TestOnlyAdminManagesUsers(t *testing.T) {
	for _, code := range RoleCodes() {
		if got := Allows(code, ModuleUsers, LevelRead); got != (code == RoleAdmin) {
			t.Errorf("%s: got %v", code, got)
		}
	}
}

func TestEveryRoleHasNameAndOnlyKnownModules(t *testing.T) {
	if len(Roles()) != 7 {
		t.Fatalf("got %d roles", len(Roles()))
	}
	for _, role := range Roles() {
		if role.Name == "" || role.Description == "" {
			t.Errorf("%s needs a name and description", role.Code)
		}
		for module := range role.Grants {
			if _, ok := moduleNames[module]; !ok {
				t.Errorf("%s grants unknown module %q", role.Code, module)
			}
		}
	}
}

func TestPermissionsListGrantedModulesOnly(t *testing.T) {
	permissions := Permissions(RoleSales)
	want := map[string]string{"sales": "write", "projects": "read", "billing": "read", "master": "read", "reports": "read"}
	if len(permissions) != len(want) {
		t.Fatalf("got %v", permissions)
	}
	for module, level := range want {
		if permissions[module] != level {
			t.Errorf("%s: got %q, want %q", module, permissions[module], level)
		}
	}
	if len(Permissions("")) != 0 {
		t.Fatal("a user without role has no permissions")
	}
}

func TestDeniedMessageNamesTheMissingAccess(t *testing.T) {
	if got := DeniedMessage(ModuleFinance, LevelWrite); got != "Peran Anda tidak punya akses mengubah data modul Keuangan" {
		t.Fatalf("got %q", got)
	}
	if got := DeniedMessage(ModuleHR, LevelRead); got != "Peran Anda tidak punya akses melihat modul SDM" {
		t.Fatalf("got %q", got)
	}
}
