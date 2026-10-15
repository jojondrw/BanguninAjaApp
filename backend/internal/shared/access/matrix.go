package access

import (
	"net/http"
	"sort"
	"strings"
)

type Module string

type Level int

const (
	LevelNone Level = iota
	LevelRead
	LevelWrite
)

const (
	ModuleUnmapped    Module = ""
	ModuleProjects    Module = "projects"
	ModuleLocations   Module = "locations"
	ModuleFinance     Module = "finance"
	ModuleBilling     Module = "billing"
	ModuleSales       Module = "sales"
	ModuleProcurement Module = "procurement"
	ModuleInventory   Module = "inventory"
	ModuleHR          Module = "hr"
	ModuleMaster      Module = "master"
	ModuleReports     Module = "reports"
	ModuleUsers       Module = "users"
)

const (
	RoleAdmin          = "admin"
	RoleProjectManager = "manajer_proyek"
	RoleFinance        = "keuangan"
	RoleProcurement    = "pengadaan"
	RoleSales          = "penjualan"
	RoleHR             = "sdm"
	RoleObserver       = "pengamat"
)

const apiPrefix = "/api"

type Role struct {
	Code         string
	Name         string
	Description  string
	Unrestricted bool
	Grants       map[Module]Level
}

var moduleNames = map[Module]string{
	ModuleProjects:    "Proyek",
	ModuleLocations:   "Analisis Lokasi",
	ModuleFinance:     "Keuangan",
	ModuleBilling:     "Tagihan",
	ModuleSales:       "Penjualan",
	ModuleProcurement: "Pengadaan",
	ModuleInventory:   "Inventaris",
	ModuleHR:          "SDM",
	ModuleMaster:      "Data Master",
	ModuleReports:     "Laporan",
	ModuleUsers:       "Pengguna",
}

var routeModules = map[string]Module{
	"projects":    ModuleProjects,
	"locations":   ModuleLocations,
	"site":        ModuleLocations,
	"scoring":     ModuleLocations,
	"regulation":  ModuleLocations,
	"news":        ModuleLocations,
	"finance":     ModuleFinance,
	"billing":     ModuleBilling,
	"sales":       ModuleSales,
	"procurement": ModuleProcurement,
	"inventory":   ModuleInventory,
	"assets":      ModuleInventory,
	"equipment":   ModuleInventory,
	"hr":          ModuleHR,
	"master":      ModuleMaster,
	"reports":     ModuleReports,
	"users":       ModuleUsers,
}

var openSegments = map[string]bool{
	"auth": true,
	"pay":  true,
}

var roles = []Role{
	{
		Code:         RoleAdmin,
		Name:         "Admin",
		Description:  "Semua modul termasuk kelola pengguna",
		Unrestricted: true,
		Grants:       everyModule(LevelWrite),
	},
	{
		Code:        RoleProjectManager,
		Name:        "Manajer Proyek",
		Description: "Proyek, analisis lokasi, dan inventaris; melihat pengadaan dan keuangan",
		Grants: map[Module]Level{
			ModuleProjects:    LevelWrite,
			ModuleLocations:   LevelWrite,
			ModuleInventory:   LevelWrite,
			ModuleProcurement: LevelRead,
			ModuleFinance:     LevelRead,
			ModuleMaster:      LevelRead,
			ModuleReports:     LevelWrite,
		},
	},
	{
		Code:        RoleFinance,
		Name:        "Keuangan",
		Description: "Keuangan dan tagihan; melihat proyek, penjualan, dan pengadaan",
		Grants: map[Module]Level{
			ModuleFinance:     LevelWrite,
			ModuleBilling:     LevelWrite,
			ModuleProjects:    LevelRead,
			ModuleSales:       LevelRead,
			ModuleProcurement: LevelRead,
			ModuleMaster:      LevelRead,
			ModuleReports:     LevelWrite,
		},
	},
	{
		Code:        RoleProcurement,
		Name:        "Pengadaan",
		Description: "Pengadaan dan inventaris",
		Grants: map[Module]Level{
			ModuleProcurement: LevelWrite,
			ModuleInventory:   LevelWrite,
			ModuleProjects:    LevelRead,
			ModuleMaster:      LevelRead,
			ModuleReports:     LevelRead,
		},
	},
	{
		Code:        RoleSales,
		Name:        "Penjualan",
		Description: "Penjualan; melihat proyek dan piutang di tagihan",
		Grants: map[Module]Level{
			ModuleSales:    LevelWrite,
			ModuleProjects: LevelRead,
			ModuleBilling:  LevelRead,
			ModuleMaster:   LevelRead,
			ModuleReports:  LevelRead,
		},
	},
	{
		Code:        RoleHR,
		Name:        "SDM",
		Description: "Karyawan, absensi, dan penggajian",
		Grants: map[Module]Level{
			ModuleHR:       LevelWrite,
			ModuleProjects: LevelRead,
			ModuleMaster:   LevelRead,
			ModuleReports:  LevelRead,
		},
	},
	{
		Code:        RoleObserver,
		Name:        "Pengamat",
		Description: "Melihat semua modul tanpa mengubah data, kecuali kelola pengguna",
		Grants:      withoutModule(everyModule(LevelRead), ModuleUsers),
	},
}

func everyModule(level Level) map[Module]Level {
	grants := make(map[Module]Level, len(moduleNames))
	for module := range moduleNames {
		grants[module] = level
	}
	return grants
}

func withoutModule(grants map[Module]Level, module Module) map[Module]Level {
	delete(grants, module)
	return grants
}

func Roles() []Role {
	return append([]Role(nil), roles...)
}

func RoleCodes() []string {
	codes := make([]string, 0, len(roles))
	for _, role := range roles {
		codes = append(codes, role.Code)
	}
	return codes
}

func FindRole(code string) (Role, bool) {
	for _, role := range roles {
		if role.Code == code {
			return role, true
		}
	}
	return Role{}, false
}

func Modules() []Module {
	modules := make([]Module, 0, len(moduleNames))
	for module := range moduleNames {
		modules = append(modules, module)
	}
	sort.Slice(modules, func(i, j int) bool { return modules[i] < modules[j] })
	return modules
}

func ModuleName(module Module) string {
	if name, ok := moduleNames[module]; ok {
		return name
	}
	return "ini"
}

func ModuleForPath(path string) (Module, bool) {
	relative, underAPI := strings.CutPrefix(path, apiPrefix+"/")
	segment := firstSegment(relative)
	if !underAPI || segment == "" || openSegments[segment] {
		return ModuleUnmapped, false
	}
	return routeModules[segment], true
}

func firstSegment(path string) string {
	segment, _, _ := strings.Cut(path, "/")
	return segment
}

func LevelForMethod(method string) Level {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return LevelRead
	default:
		return LevelWrite
	}
}

func Allows(roleCode string, module Module, needed Level) bool {
	role, ok := FindRole(roleCode)
	if !ok {
		return false
	}
	if role.Unrestricted {
		return true
	}
	return role.Grants[module] >= needed
}

func (l Level) String() string {
	switch l {
	case LevelWrite:
		return "write"
	case LevelRead:
		return "read"
	default:
		return "none"
	}
}

func Permissions(roleCode string) map[string]string {
	role, ok := FindRole(roleCode)
	if !ok {
		return map[string]string{}
	}
	permissions := make(map[string]string, len(role.Grants))
	for module, level := range role.Grants {
		if level > LevelNone {
			permissions[string(module)] = level.String()
		}
	}
	return permissions
}

func DeniedMessage(module Module, needed Level) string {
	if module == ModuleUnmapped {
		return "Peran Anda belum diberi akses ke fitur ini"
	}
	if needed == LevelWrite {
		return "Peran Anda tidak punya akses mengubah data modul " + ModuleName(module)
	}
	return "Peran Anda tidak punya akses melihat modul " + ModuleName(module)
}
