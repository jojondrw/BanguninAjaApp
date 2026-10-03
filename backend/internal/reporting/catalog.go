package reporting

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	typeCashFlow    = "arus_kas"
	typeBudget      = "anggaran"
	typeProjects    = "proyek"
	typeAging       = "piutang_utang"
	typeStock       = "stok"
	typePayroll     = "penggajian"
	typeSales       = "penjualan"
	fileNamePrefix  = "laporan-"
	monthPeriod     = "2006-01"
	isoDate         = "2006-01-02"
	compactDate     = "02/01/2006"
	filterSeparator = " · "
)

type reportKind struct {
	key         string
	title       string
	usesPeriod  bool
	monthlyOnly bool
}

var reportKinds = map[string]reportKind{
	typeCashFlow: {key: typeCashFlow, title: "Arus Kas", usesPeriod: true, monthlyOnly: true},
	typeBudget:   {key: typeBudget, title: "Anggaran dan Realisasi", usesPeriod: true},
	typeProjects: {key: typeProjects, title: "Proyek", usesPeriod: true},
	typeAging:    {key: typeAging, title: "Umur Piutang dan Utang", usesPeriod: true},
	typeStock:    {key: typeStock, title: "Stok Material"},
	typePayroll:  {key: typePayroll, title: "Penggajian", usesPeriod: true, monthlyOnly: true},
	typeSales:    {key: typeSales, title: "Penjualan", usesPeriod: true},
}

var reportKindAliases = map[string]string{
	"cash_flow":            typeCashFlow,
	"cashflow":             typeCashFlow,
	"budget":               typeBudget,
	"budget_realization":   typeBudget,
	"project":              typeProjects,
	"projects":             typeProjects,
	"receivable_payable":   typeAging,
	"receivables_payables": typeAging,
	"aging":                typeAging,
	"stock":                typeStock,
	"inventory":            typeStock,
	"payroll":              typePayroll,
	"sales":                typeSales,
}

var monthNames = [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

type reportSpec struct {
	kind        reportKind
	projectID   *uuid.UUID
	periodStart *time.Time
	periodEnd   *time.Time
}

func resolveKind(value string) (reportKind, bool) {
	key := strings.ToLower(strings.TrimSpace(value))
	if alias, ok := reportKindAliases[key]; ok {
		key = alias
	}
	kind, ok := reportKinds[key]
	return kind, ok
}

func newReportSpec(typeName string, projectID *uuid.UUID, start, end *time.Time, today time.Time) (reportSpec, error) {
	kind, ok := resolveKind(typeName)
	if !ok {
		return reportSpec{}, errReportTypeInvalid
	}
	if !kind.usesPeriod && (start != nil || end != nil) {
		return reportSpec{}, errReportPeriodUnsupported
	}
	if !dateRangeValid(start, end) {
		return reportSpec{}, errInvalidDateRange
	}
	spec := reportSpec{kind: kind, projectID: projectID, periodStart: dateOnly(start), periodEnd: dateOnly(end)}
	if kind.monthlyOnly {
		spec.periodStart, spec.periodEnd = monthBounds(spec.periodStart, spec.periodEnd, today)
	}
	return spec, nil
}

func monthBounds(start, end *time.Time, today time.Time) (*time.Time, *time.Time) {
	switch {
	case start == nil && end == nil:
		first := firstOfMonth(today)
		last := lastOfMonth(today)
		return &first, &last
	case start == nil:
		first := firstOfMonth(*end)
		return &first, end
	case end == nil:
		last := lastOfMonth(*start)
		return start, &last
	default:
		return start, end
	}
}

func dateOnly(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	year, month, day := value.Date()
	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &date
}

func firstOfMonth(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func lastOfMonth(value time.Time) time.Time {
	return firstOfMonth(value).AddDate(0, 1, -1)
}

func monthLabel(value time.Time) string {
	return fmt.Sprintf("%s %d", monthNames[value.Month()-1], value.Year())
}

func periodLabel(period string) string {
	parsed, err := time.Parse(monthPeriod, period)
	if err != nil {
		return period
	}
	return monthLabel(parsed)
}

func (s reportSpec) title() string {
	return "Laporan " + s.kind.title
}

func (s reportSpec) periodText() string {
	switch {
	case s.periodStart != nil && s.periodEnd != nil:
		return "Periode " + s.periodStart.Format(compactDate) + " – " + s.periodEnd.Format(compactDate)
	case s.periodStart != nil:
		return "Mulai " + s.periodStart.Format(compactDate)
	case s.periodEnd != nil:
		return "Sampai " + s.periodEnd.Format(compactDate)
	case s.kind.usesPeriod:
		return "Semua periode"
	default:
		return ""
	}
}

func (s reportSpec) fileName(projectCode, extension string, today time.Time) string {
	parts := []string{fileNamePrefix + strings.ReplaceAll(s.kind.key, "_", "-")}
	if slug := slugify(projectCode); slug != "" {
		parts = append(parts, slug)
	}
	parts = append(parts, s.periodSlug(today))
	return strings.Join(parts, "-") + "." + extension
}

func (s reportSpec) periodSlug(today time.Time) string {
	start, end := s.periodStart, s.periodEnd
	switch {
	case start != nil && end != nil && coversMonth(*start, *end):
		return start.Format(monthPeriod)
	case start != nil && end != nil && coversYear(*start, *end):
		return start.Format("2006")
	case start != nil && end != nil:
		return start.Format(isoDate) + "_" + end.Format(isoDate)
	case start != nil:
		return "sejak-" + start.Format(isoDate)
	case end != nil:
		return "sampai-" + end.Format(isoDate)
	default:
		return today.Format(isoDate)
	}
}

func coversMonth(start, end time.Time) bool {
	return start.Equal(firstOfMonth(start)) && end.Equal(lastOfMonth(start))
}

func coversYear(start, end time.Time) bool {
	return start.Month() == time.January && start.Day() == 1 && end.Year() == start.Year() && end.Month() == time.December && end.Day() == 31
}

func slugify(text string) string {
	var builder strings.Builder
	lastDash := true
	for _, character := range strings.ToLower(text) {
		isWordCharacter := (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9')
		if isWordCharacter {
			builder.WriteRune(character)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
	}
	return strings.TrimSuffix(builder.String(), "-")
}
