package location

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/regulation"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/reporting/export"
)

const (
	siteReportTitle       = "Laporan Analisis Lokasi"
	comparisonReportTitle = "Laporan Perbandingan Lokasi"
	siteFilePrefix        = "analisis-lokasi"
	comparisonFilePrefix  = "perbandingan-lokasi"
	reportFileExtension   = ".pdf"
	fileDateLayout        = "2006-01-02"
	timestampLayout       = "02/01/2006 15:04"
	reportZoneName        = "WIB"
	reportZoneOffset      = 7 * 60 * 60
	coordinateDigits      = 6
	fullScore             = 100
	samplePointCount      = 1489
	slugLimit             = 60
	filterSeparator       = " · "
	notRecorded           = "Tidak tercatat"
	notAvailable          = "Tidak tersedia"
	missingValue          = "-"
	unknownSeverityRank   = 9
)

const (
	riskNotRecordedText = "Penanda risiko tidak tercatat: lokasi ini dievaluasi sebelum penanda risiko ikut disimpan. " +
		"Evaluasi ulang lokasinya untuk mendapatkan penanda yang lengkap."
	noRiskText        = "Tidak ada penanda risiko."
	missingWeightText = "Bobot tidak tersedia karena profil bangunan lokasi ini tidak tercatat."
	hazardScaleText   = "Indeks bahaya berskala 0 sampai 1: makin rendah makin aman. Nilai 0 juga dipakai bila data " +
		"tidak tersedia di titik ini."
	regulationNoteText  = "Konteks regulasi bersifat deskriptif dan tidak ikut dihitung dalam skor."
	simulatedRuleText   = "Simulasi (data RDTR belum tersedia di titik ini)"
	recordedRuleText    = "RDTR"
	comparisonOrderText = "Kolom mengikuti urutan lokasi di perbandingan. Skor berskala 0 sampai 100, makin tinggi makin " +
		"baik; harga tanah makin murah makin baik."
	comparisonRiskGapText = "Tidak tercatat berarti lokasi itu dievaluasi sebelum penanda risiko ikut disimpan."
	emptyComparisonText   = "Semua lokasi di perbandingan ini sudah dihapus, jadi tidak ada yang bisa dibandingkan."
	comparisonSourceText  = "Laporan ini disusun dari hasil evaluasi yang tersimpan untuk tiap lokasi; skornya tidak " +
		"dihitung ulang. Semua waktu memakai WIB (UTC+7)."
)

var reportZone = time.FixedZone(reportZoneName, reportZoneOffset)

var severityLabels = map[string]string{"high": "Tinggi", "medium": "Sedang", "low": "Rendah"}

var severityRanks = map[string]int{"high": 0, "medium": 1, "low": 2}

type siteReport struct {
	Location    SavedLocation
	Scores      []DimensionScore
	Dimensions  []DimensionLabel
	Weights     map[uuid.UUID]int
	ProfileName string
	RegionName  string
	Project     *ProjectLabel
	Regulation  *regulation.Response
	GeneratedAt time.Time
}

type comparisonReport struct {
	Comparison  Comparison
	Locations   []SavedLocation
	Scores      []DimensionScore
	Dimensions  []DimensionLabel
	GeneratedAt time.Time
}

func newSiteDocument(report siteReport) export.Document {
	sections := []export.Section{
		identitySection(report),
		scoreSection(report),
		riskSection(decodeRiskFlags(report.Location.RiskFlags)),
		factSection(report.Location),
	}
	if report.Regulation != nil {
		sections = append(sections, regulationSection(*report.Regulation))
	}
	source := fmt.Sprintf("Laporan ini disusun dari hasil evaluasi yang tersimpan pada %s; skornya tidak dihitung ulang. "+
		"Semua waktu memakai WIB (UTC+7).", timestampText(report.Location.SavedAt))
	return export.Document{
		Title:       siteReportTitle,
		Filters:     joinPresent(filterSeparator, report.Location.Name, report.ProfileName),
		GeneratedAt: report.GeneratedAt,
		Sections:    append(sections, readingGuideSection(source)),
	}
}

func identitySection(report siteReport) export.Section {
	location := report.Location
	fields := []export.Field{
		{Label: "Nama lokasi", Value: location.Name},
		{Label: "Lintang (latitude)", Value: coordinateText(location.Point.Lat)},
		{Label: "Bujur (longitude)", Value: coordinateText(location.Point.Lon)},
		{Label: "Profil bangunan", Value: textOr(report.ProfileName, notRecorded)},
		{Label: "Wilayah", Value: textOr(report.RegionName, notRecorded)},
		{Label: "Proyek terkait", Value: projectText(report.Project)},
	}
	if location.AreaSqm > 0 {
		fields = append(fields, export.Field{Label: "Luas lahan", Value: export.DisplayText(export.Decimal, location.AreaSqm) + " m²"})
	}
	fields = append(fields, export.Field{Label: "Disimpan", Value: timestampText(location.SavedAt)})
	if note := strings.TrimSpace(location.Note); note != "" {
		fields = append(fields, export.Field{Label: "Catatan", Value: note})
	}
	return export.Section{Title: "Identitas lokasi", Fields: fields}
}

func scoreSourceText(simulated bool) string {
	if simulated {
		return "SIMULASI. Layanan penilaian tidak tersedia saat lokasi dievaluasi, jadi skor ini angka contoh tanpa bobot dan bukan hasil data GIS. Evaluasi ulang sebelum dipakai untuk keputusan"
	}
	return "Model penilaian dari data GIS nasional, berbobot sesuai profil bangunan"
}

func scoreSection(report siteReport) export.Section {
	values := scoresOf(report.Scores)[report.Location.ID]
	section := export.Section{
		Title: "Skor kelayakan",
		Fields: []export.Field{
			{Label: "Sumber skor", Value: scoreSourceText(report.Location.ScoreSimulated)},
			{Label: "Skor keseluruhan", Value: scoreText(report.Location.Score)},
			{Label: "Penanda risiko", Value: riskSummaryText(decodeRiskFlags(report.Location.RiskFlags))},
		},
		Columns: []export.Column{
			{Header: "Dimensi", Kind: export.Text, Wrap: true},
			{Header: "Yang diukur", Kind: export.Text, Wrap: true},
			{Header: "Bobot", Kind: export.Percent},
			{Header: "Skor (0–100)", Kind: export.Integer},
		},
		Total: []any{"Jumlah berbobot (skor keseluruhan)", nil, totalWeight(report.Weights), int64(report.Location.Score)},
	}
	for _, dimension := range report.Dimensions {
		section.Rows = append(section.Rows, []any{
			dimension.Name, dimension.Description, weightCell(report.Weights, dimension.ID), scoreCell(values, dimension.ID),
		})
	}
	if len(report.Weights) == 0 {
		section.Paragraphs = []string{missingWeightText}
	}
	return section
}

func riskSection(flags []RiskFlag) export.Section {
	section := export.Section{Title: "Penanda risiko"}
	switch {
	case flags == nil:
		section.Paragraphs = []string{riskNotRecordedText}
	case len(flags) == 0:
		section.Paragraphs = []string{noRiskText}
	default:
		section.Columns = []export.Column{
			{Header: "Kode", Kind: export.Text},
			{Header: "Tingkat", Kind: export.Text},
			{Header: "Pesan", Kind: export.Text, Wrap: true},
		}
		for _, flag := range sortedFlags(flags) {
			section.Rows = append(section.Rows, []any{flag.Code, severityLabel(flag.Severity), flag.Message})
		}
	}
	return section
}

func factSection(location SavedLocation) export.Section {
	return export.Section{
		Title: "Fakta kunci",
		Fields: []export.Field{
			{Label: "Indeks bahaya banjir", Value: export.DisplayText(export.Decimal, location.FloodIndex)},
			{Label: "Indeks bahaya gempa bumi", Value: export.DisplayText(export.Decimal, location.EarthquakeIndex)},
			{Label: "Harga tanah ZNT per m²", Value: landPriceText(location.LandPricePerSqm)},
		},
		Paragraphs: []string{hazardScaleText},
	}
}

func regulationSection(rule regulation.Response) export.Section {
	fields := []export.Field{
		{Label: "Zona", Value: zoneText(rule)},
		{Label: "KDB", Value: export.DisplayText(export.Percent, rule.KDB)},
		{Label: "KLB", Value: export.DisplayText(export.Decimal, rule.KLB)},
		{Label: "Status data", Value: ruleStatusText(rule.IsSimulated)},
	}
	if area := joinPresent(", ", rule.District, rule.Village); area != "" {
		fields = append(fields, export.Field{Label: "Kecamatan, kelurahan", Value: area})
	}
	paragraphs := []string{regulationNoteText}
	if note := strings.TrimSpace(rule.RegulationNote); note != "" {
		paragraphs = append([]string{note}, paragraphs...)
	}
	return export.Section{Title: "Konteks regulasi", Fields: fields, Paragraphs: paragraphs}
}

func readingGuideSection(source string) export.Section {
	return export.Section{
		Title: "Cara membaca skor",
		Paragraphs: []string{
			"Skor keseluruhan 0–100 adalah jumlah berbobot dari lima skor dimensi. Bobot tiap dimensi mengikuti " +
				"profil bangunan yang dipilih saat evaluasi, dan jumlah semua bobotnya 100%.",
			fmt.Sprintf("Setiap skor dimensi juga berskala 0–100 dan dihitung dengan membandingkan lokasi terhadap %s "+
				"titik sampel di seluruh Indonesia. Makin tinggi skornya, makin baik posisi lokasi dibanding titik sampel lain.",
				export.DisplayText(export.Integer, samplePointCount)),
			"Penanda risiko dinilai terpisah dan tidak mengubah skor. Lokasi dengan skor tinggi tetap perlu diperiksa " +
				"penanda risikonya sebelum diputuskan.",
			source,
		},
	}
}

func newComparisonDocument(report comparisonReport) export.Document {
	labels := locationLabels(report.Locations)
	sections := []export.Section{comparisonSummarySection(report, labels)}
	if len(report.Locations) > 0 {
		sections = append(sections, comparisonTableSection(report, labels))
	}
	return export.Document{
		Title:       comparisonReportTitle,
		Filters:     report.Comparison.Name + filterSeparator + fmt.Sprintf("%d lokasi", len(report.Locations)),
		GeneratedAt: report.GeneratedAt,
		Sections:    append(sections, readingGuideSection(comparisonSourceText)),
	}
}

func comparisonSummarySection(report comparisonReport, labels []string) export.Section {
	section := export.Section{
		Title: "Ringkasan",
		Fields: []export.Field{
			{Label: "Perbandingan", Value: report.Comparison.Name},
			{Label: "Skor tertinggi", Value: topScoreText(report.Locations, labels)},
			{Label: "Jumlah lokasi", Value: strconv.Itoa(len(report.Locations))},
		},
	}
	if len(report.Locations) == 0 {
		section.Paragraphs = []string{emptyComparisonText}
	}
	return section
}

func comparisonTableSection(report comparisonReport, labels []string) export.Section {
	columns := []export.Column{{Header: "Aspek", Kind: export.Text, Wrap: true}}
	for _, label := range labels {
		columns = append(columns, export.Column{Header: label, Kind: export.Integer, Wrap: true})
	}
	scores := scoresOf(report.Scores)
	rows := [][]any{aspectRow("Skor keseluruhan (0–100)", report.Locations, func(location SavedLocation) any {
		return int64(location.Score)
	})}
	for _, dimension := range report.Dimensions {
		rows = append(rows, aspectRow(dimension.Name, report.Locations, func(location SavedLocation) any {
			return scoreCell(scores[location.ID], dimension.ID)
		}))
	}
	rows = append(rows,
		aspectRow("Penanda risiko (jumlah)", report.Locations, riskCountCell),
		aspectRow("Harga tanah ZNT (Rp/m²)", report.Locations, landPriceCell),
	)
	section := export.Section{Title: "Perbandingan per lokasi", Columns: columns, Rows: rows, Paragraphs: []string{comparisonOrderText}}
	if hasUnrecordedRisk(report.Locations) {
		section.Paragraphs = append(section.Paragraphs, comparisonRiskGapText)
	}
	return section
}

func aspectRow(label string, locations []SavedLocation, value func(SavedLocation) any) []any {
	row := []any{label}
	for _, location := range locations {
		row = append(row, value(location))
	}
	return row
}

func riskCountCell(location SavedLocation) any {
	flags := decodeRiskFlags(location.RiskFlags)
	if flags == nil {
		return notRecorded
	}
	return int64(len(flags))
}

func landPriceCell(location SavedLocation) any {
	if location.LandPricePerSqm <= 0 {
		return notAvailable
	}
	return location.LandPricePerSqm
}

func hasUnrecordedRisk(locations []SavedLocation) bool {
	for _, location := range locations {
		if decodeRiskFlags(location.RiskFlags) == nil {
			return true
		}
	}
	return false
}

func topScoreText(locations []SavedLocation, labels []string) string {
	if len(locations) == 0 {
		return "Belum ada lokasi"
	}
	top := locations[0].Score
	for _, location := range locations {
		top = max(top, location.Score)
	}
	var leaders []string
	for index, location := range locations {
		if location.Score == top {
			leaders = append(leaders, labels[index])
		}
	}
	return joinWords(leaders) + " dengan skor " + scoreText(top)
}

func locationLabels(locations []SavedLocation) []string {
	totals := make(map[string]int, len(locations))
	for _, location := range locations {
		totals[location.Name]++
	}
	seen := make(map[string]int, len(locations))
	labels := make([]string, 0, len(locations))
	for _, location := range locations {
		if totals[location.Name] < 2 {
			labels = append(labels, location.Name)
			continue
		}
		seen[location.Name]++
		labels = append(labels, fmt.Sprintf("%s (%d)", location.Name, seen[location.Name]))
	}
	return labels
}

func orderLocations(items []ComparisonItem, locations []SavedLocation) []SavedLocation {
	byID := make(map[uuid.UUID]SavedLocation, len(locations))
	for _, location := range locations {
		byID[location.ID] = location
	}
	ordered := make([]SavedLocation, 0, len(items))
	for _, item := range items {
		if location, ok := byID[item.SavedLocationID]; ok {
			ordered = append(ordered, location)
		}
	}
	return ordered
}

func scoresOf(scores []DimensionScore) map[uuid.UUID]map[uuid.UUID]int {
	grouped := make(map[uuid.UUID]map[uuid.UUID]int)
	for _, score := range scores {
		if grouped[score.SavedLocationID] == nil {
			grouped[score.SavedLocationID] = make(map[uuid.UUID]int)
		}
		grouped[score.SavedLocationID][score.DimensionID] = score.Value
	}
	return grouped
}

func weightsByDimension(weights []ProfileWeight) map[uuid.UUID]int {
	byDimension := make(map[uuid.UUID]int, len(weights))
	for _, weight := range weights {
		byDimension[weight.DimensionID] = weight.Percent
	}
	return byDimension
}

func scoreCell(values map[uuid.UUID]int, dimensionID uuid.UUID) any {
	value, ok := values[dimensionID]
	if !ok {
		return missingValue
	}
	return int64(value)
}

func weightCell(weights map[uuid.UUID]int, dimensionID uuid.UUID) any {
	percent, ok := weights[dimensionID]
	if !ok {
		return missingValue
	}
	return int64(percent)
}

func totalWeight(weights map[uuid.UUID]int) any {
	if len(weights) == 0 {
		return nil
	}
	total := 0
	for _, percent := range weights {
		total += percent
	}
	return int64(total)
}

func sortedFlags(flags []RiskFlag) []RiskFlag {
	sorted := append([]RiskFlag(nil), flags...)
	sort.SliceStable(sorted, func(left, right int) bool {
		return severityRank(sorted[left].Severity) < severityRank(sorted[right].Severity)
	})
	return sorted
}

func severityRank(severity string) int {
	if rank, ok := severityRanks[severity]; ok {
		return rank
	}
	return unknownSeverityRank
}

func severityLabel(severity string) string {
	return textOr(severityLabels[severity], severity)
}

func riskSummaryText(flags []RiskFlag) string {
	if flags == nil {
		return notRecorded
	}
	if len(flags) == 0 {
		return "Tidak ada"
	}
	counts := make(map[string]int, len(flags))
	var severities []string
	for _, flag := range sortedFlags(flags) {
		if counts[flag.Severity] == 0 {
			severities = append(severities, flag.Severity)
		}
		counts[flag.Severity]++
	}
	parts := make([]string, 0, len(severities))
	for _, severity := range severities {
		parts = append(parts, fmt.Sprintf("%d %s", counts[severity], strings.ToLower(severityLabel(severity))))
	}
	return fmt.Sprintf("%d (%s)", len(flags), strings.Join(parts, ", "))
}

func scoreText(score int) string {
	return fmt.Sprintf("%d dari %d", score, fullScore)
}

func landPriceText(price int64) string {
	if price <= 0 {
		return notAvailable
	}
	return "Rp " + export.DisplayText(export.Money, price)
}

func coordinateText(value float64) string {
	return strings.Replace(strconv.FormatFloat(value, 'f', coordinateDigits, 64), ".", ",", 1)
}

func timestampText(value time.Time) string {
	return value.In(reportZone).Format(timestampLayout) + " " + reportZoneName
}

func projectText(project *ProjectLabel) string {
	if project == nil {
		return "Tanpa proyek"
	}
	return joinPresent(filterSeparator, project.Code, project.Name)
}

func zoneText(rule regulation.Response) string {
	zone := joinPresent(" – ", rule.ZoneName, rule.SubZoneName)
	code := textOr(rule.SubZoneCode, rule.ZoneCode)
	if rule.IsSimulated || code == "" {
		return textOr(zone, missingValue)
	}
	return zone + " (" + code + ")"
}

func ruleStatusText(isSimulated bool) string {
	if isSimulated {
		return simulatedRuleText
	}
	return recordedRuleText
}

func textOr(text, fallback string) string {
	if strings.TrimSpace(text) == "" {
		return fallback
	}
	return text
}

func joinPresent(separator string, parts ...string) string {
	present := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			present = append(present, trimmed)
		}
	}
	return strings.Join(present, separator)
}

func joinWords(words []string) string {
	if len(words) <= 1 {
		return strings.Join(words, "")
	}
	return strings.Join(words[:len(words)-1], ", ") + " dan " + words[len(words)-1]
}

func siteFileName(name string, at time.Time) string {
	return reportFileName(siteFilePrefix, name, at)
}

func comparisonFileName(name string, at time.Time) string {
	return reportFileName(comparisonFilePrefix, name, at)
}

func reportFileName(prefix, name string, at time.Time) string {
	parts := []string{prefix}
	if slug := slugify(name); slug != "" {
		parts = append(parts, slug)
	}
	parts = append(parts, at.Format(fileDateLayout))
	return strings.Join(parts, "-") + reportFileExtension
}

func slugify(text string) string {
	var builder strings.Builder
	lastDash := true
	for _, character := range strings.ToLower(text) {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') {
			builder.WriteRune(character)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
	}
	slug := builder.String()
	if len(slug) > slugLimit {
		slug = slug[:slugLimit]
	}
	return strings.Trim(slug, "-")
}
