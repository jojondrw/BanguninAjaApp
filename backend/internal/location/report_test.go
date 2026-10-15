package location

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/regulation"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/reporting/export"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/geo"
)

const reportSamplesVariable = "SITE_REPORT_SAMPLES_DIR"

var reportGeneratedAt = time.Date(2026, 10, 15, 7, 5, 0, 0, time.UTC).In(reportZone)

var reportDimensions = []DimensionLabel{
	{ID: uuid.MustParse("59162a34-7edd-4dfa-950e-3324721da9e2"), Code: "fisik_lingkungan", Name: "Fisik & Lingkungan", Description: "Risiko bencana gabungan (InaRISK) dan kemiringan lahan", SortOrder: 1},
	{ID: uuid.MustParse("18af97ba-7592-4fb2-bc72-41c826aa4f70"), Code: "infrastruktur", Name: "Infrastruktur & Aksesibilitas", Description: "Jarak ke jalan, simpul transit, dan fasilitas kesehatan", SortOrder: 2},
	{ID: uuid.MustParse("bbf521c4-bd06-4e69-bafb-bcf2a2d64c11"), Code: "demografi_sosial", Name: "Demografi & Sosial", Description: "Jumlah penduduk di sekitar lokasi (WorldPop)", SortOrder: 3},
	{ID: uuid.MustParse("3a3c1f50-8d43-4c41-9f0e-7e2b1f0c9d21"), Code: "pasar_kompetisi", Name: "Pasar & Kompetisi", Description: "Kepadatan pesaing sejenis dan aktivitas komersial sekitar", SortOrder: 4},
	{ID: uuid.MustParse("c6a0a5d2-3b8e-4f7a-a1d4-5e9f0b2c7e83"), Code: "finansial_proyek", Name: "Finansial Proyek", Description: "Harga tanah (ZNT) dan indeks kemahalan konstruksi", SortOrder: 5},
}

var housingWeights = []int{10, 10, 53, 10, 17}

func stored(text string) *string {
	return &text
}

func reportLocation(name string, score int, values []int, flags *string) (SavedLocation, []DimensionScore) {
	location := SavedLocation{
		Base:            entity.Base{ID: uuid.New()},
		Name:            name,
		Point:           geo.Point{Lon: 106.761635, Lat: -6.192872},
		LandPricePerSqm: 8580000,
		Score:           score,
		FloodIndex:      0.776,
		EarthquakeIndex: 0.472,
		RiskFlags:       flags,
		SavedAt:         time.Date(2026, 10, 5, 4, 27, 59, 0, time.UTC),
	}
	scores := make([]DimensionScore, 0, len(values))
	for index, value := range values {
		scores = append(scores, DimensionScore{SavedLocationID: location.ID, DimensionID: reportDimensions[index].ID, Value: value})
	}
	return location, scores
}

func fullSiteReport() siteReport {
	flags := stored(`[{"code":"rawan_gempabumi","severity":"medium","message":"Indeks bahaya gempa bumi sedang (0,47)."},` +
		`{"code":"rawan_banjir","severity":"high","message":"Indeks bahaya banjir tinggi (0,78)."},` +
		`{"code":"rawan_likuefaksi","severity":"medium","message":"Indeks bahaya likuefaksi sedang (0,44)."}]`)
	location, scores := reportLocation("Kandidat Jakarta Barat", 78, []int{21, 88, 95, 64, 41}, flags)
	location.AreaSqm = 9800
	location.Note = "Akses dari Jl. Panjang; perlu cek genangan saat musim hujan."
	weights := make(map[uuid.UUID]int, len(reportDimensions))
	for index, dimension := range reportDimensions {
		weights[dimension.ID] = housingWeights[index]
	}
	return siteReport{
		Location:    location,
		Scores:      scores,
		Dimensions:  reportDimensions,
		Weights:     weights,
		ProfileName: "Perumahan",
		RegionName:  "Kota Administrasi Jakarta Barat",
		Project:     &ProjectLabel{Code: "DEMO-01", Name: "Contoh Kembangan Residence"},
		Regulation: &regulation.Response{
			ZoneName: "Kawasan Budidaya", ZoneCode: "SIM", SubZoneName: "Permukiman/Komersial", KDB: 60, KLB: 2.4, IsSimulated: true,
			RegulationNote: "Data RDTR belum tersedia untuk lokasi ini; nilai bersifat simulasi dan tidak digunakan dalam scoring.",
		},
		GeneratedAt: reportGeneratedAt,
	}
}

func comparedReport() comparisonReport {
	first, firstScores := reportLocation("Kandidat Jakarta Barat", 78, []int{21, 88, 95, 64, 41}, stored(`[{"code":"rawan_banjir","severity":"high","message":"Indeks bahaya banjir tinggi (0,78)."}]`))
	second, secondScores := reportLocation("Lahan Cibubur", 78, []int{70, 61, 82, 58, 49}, stored(`[]`))
	third, thirdScores := reportLocation("Lahan Cibubur", 64, []int{66, 52, 70}, nil)
	third.LandPricePerSqm = 0
	scores := append(append(firstScores, secondScores...), thirdScores...)
	return comparisonReport{
		Comparison:  Comparison{Base: entity.Base{ID: uuid.New()}, Name: "Kandidat lahan perumahan Jabodetabek"},
		Locations:   []SavedLocation{first, second, third},
		Scores:      scores,
		Dimensions:  reportDimensions,
		GeneratedAt: reportGeneratedAt,
	}
}

func sectionTitled(t *testing.T, document export.Document, title string) export.Section {
	t.Helper()
	for _, section := range document.Sections {
		if section.Title == title {
			return section
		}
	}
	t.Fatalf("missing section %q", title)
	return export.Section{}
}

func fieldValue(t *testing.T, section export.Section, label string) string {
	t.Helper()
	for _, field := range section.Fields {
		if field.Label == label {
			return field.Value
		}
	}
	t.Fatalf("section %q has no field %q", section.Title, label)
	return ""
}

func sectionTitles(document export.Document) string {
	titles := make([]string, 0, len(document.Sections))
	for _, section := range document.Sections {
		titles = append(titles, section.Title)
	}
	return strings.Join(titles, "|")
}

func TestSiteDocumentCoversTheStoredEvaluation(t *testing.T) {
	document := newSiteDocument(fullSiteReport())

	if document.Title != "Laporan Analisis Lokasi" || document.Filters != "Kandidat Jakarta Barat · Perumahan" {
		t.Fatalf("got heading %q / %q", document.Title, document.Filters)
	}
	if got := sectionTitles(document); got != "Identitas lokasi|Skor kelayakan|Penanda risiko|Fakta kunci|Konteks regulasi|Cara membaca skor" {
		t.Fatalf("got sections %s", got)
	}

	identity := sectionTitled(t, document, "Identitas lokasi")
	expectations := map[string]string{
		"Lintang (latitude)": "-6,192872",
		"Bujur (longitude)":  "106,761635",
		"Profil bangunan":    "Perumahan",
		"Wilayah":            "Kota Administrasi Jakarta Barat",
		"Proyek terkait":     "DEMO-01 · Contoh Kembangan Residence",
		"Luas lahan":         "9.800 m²",
		"Disimpan":           "05/10/2026 11:27 WIB",
		"Catatan":            "Akses dari Jl. Panjang; perlu cek genangan saat musim hujan.",
	}
	for label, want := range expectations {
		if got := fieldValue(t, identity, label); got != want {
			t.Errorf("%s: got %q, want %q", label, got, want)
		}
	}

	score := sectionTitled(t, document, "Skor kelayakan")
	if got := fieldValue(t, score, "Skor keseluruhan"); got != "78 dari 100" {
		t.Fatalf("got overall %q", got)
	}
	if got := fieldValue(t, score, "Penanda risiko"); got != "3 (1 tinggi, 2 sedang)" {
		t.Fatalf("got risk summary %q", got)
	}
	if len(score.Rows) != len(reportDimensions) {
		t.Fatalf("got %d dimension rows", len(score.Rows))
	}
	demography := score.Rows[2]
	if demography[0] != "Demografi & Sosial" || demography[1] != "Jumlah penduduk di sekitar lokasi (WorldPop)" ||
		demography[2] != int64(53) || demography[3] != int64(95) {
		t.Fatalf("got demography row %v", demography)
	}
	if score.Total[2] != int64(100) || score.Total[3] != int64(78) {
		t.Fatalf("got total %v", score.Total)
	}
	if len(score.Paragraphs) != 0 {
		t.Fatalf("weights are known, got note %v", score.Paragraphs)
	}

	risks := sectionTitled(t, document, "Penanda risiko")
	if len(risks.Rows) != 3 || risks.Rows[0][0] != "rawan_banjir" || risks.Rows[0][1] != "Tinggi" || risks.Rows[1][1] != "Sedang" {
		t.Fatalf("risk flags must list high severity first, got %v", risks.Rows)
	}

	facts := sectionTitled(t, document, "Fakta kunci")
	if fieldValue(t, facts, "Indeks bahaya banjir") != "0,78" || fieldValue(t, facts, "Indeks bahaya gempa bumi") != "0,47" ||
		fieldValue(t, facts, "Harga tanah ZNT per m²") != "Rp 8.580.000" {
		t.Fatalf("got facts %v", facts.Fields)
	}

	rule := sectionTitled(t, document, "Konteks regulasi")
	if fieldValue(t, rule, "Zona") != "Kawasan Budidaya – Permukiman/Komersial" || fieldValue(t, rule, "KDB") != "60%" ||
		fieldValue(t, rule, "KLB") != "2,4" || !strings.HasPrefix(fieldValue(t, rule, "Status data"), "Simulasi") {
		t.Fatalf("got regulation %v", rule.Fields)
	}

	guide := strings.Join(sectionTitled(t, document, "Cara membaca skor").Paragraphs, " ")
	for _, fragment := range []string{"jumlah berbobot dari lima skor dimensi", "1.489 titik sampel", "tidak mengubah skor", "05/10/2026 11:27 WIB"} {
		if !strings.Contains(guide, fragment) {
			t.Errorf("reading guide is missing %q", fragment)
		}
	}
}

func TestSiteDocumentSaysWhatIsNotStored(t *testing.T) {
	report := fullSiteReport()
	report.Location.RiskFlags = nil
	report.Location.Note = ""
	report.Location.AreaSqm = 0
	report.Location.LandPricePerSqm = 0
	report.Weights, report.ProfileName, report.RegionName, report.Project, report.Regulation = nil, "", "", nil, nil
	report.Scores = report.Scores[:3]

	document := newSiteDocument(report)
	if got := sectionTitles(document); got != "Identitas lokasi|Skor kelayakan|Penanda risiko|Fakta kunci|Cara membaca skor" {
		t.Fatalf("regulation must be left out when it is unavailable, got %s", got)
	}
	identity := sectionTitled(t, document, "Identitas lokasi")
	if fieldValue(t, identity, "Profil bangunan") != notRecorded || fieldValue(t, identity, "Proyek terkait") != "Tanpa proyek" {
		t.Fatalf("got identity %v", identity.Fields)
	}
	for _, field := range identity.Fields {
		if field.Label == "Catatan" || field.Label == "Luas lahan" {
			t.Fatalf("empty %s must be left out", field.Label)
		}
	}
	score := sectionTitled(t, document, "Skor kelayakan")
	if score.Rows[0][2] != missingValue || score.Rows[4][3] != missingValue || score.Total[2] != nil {
		t.Fatalf("missing weights and scores must show as %q, got %v total %v", missingValue, score.Rows, score.Total)
	}
	if len(score.Paragraphs) != 1 || score.Paragraphs[0] != missingWeightText {
		t.Fatalf("got score notes %v", score.Paragraphs)
	}
	risks := sectionTitled(t, document, "Penanda risiko")
	if len(risks.Columns) != 0 || risks.Paragraphs[0] != riskNotRecordedText {
		t.Fatalf("unrecorded flags must be explained, got %+v", risks)
	}
	if fieldValue(t, sectionTitled(t, document, "Fakta kunci"), "Harga tanah ZNT per m²") != notAvailable {
		t.Fatalf("a zero land price means it is not available")
	}

	report.Location.RiskFlags = stored(`[]`)
	empty := sectionTitled(t, newSiteDocument(report), "Penanda risiko")
	if len(empty.Paragraphs) != 1 || empty.Paragraphs[0] != "Tidak ada penanda risiko." {
		t.Fatalf("got %v", empty.Paragraphs)
	}
}

func TestComparisonDocumentPutsOneColumnPerLocation(t *testing.T) {
	document := newComparisonDocument(comparedReport())

	if document.Filters != "Kandidat lahan perumahan Jabodetabek · 3 lokasi" {
		t.Fatalf("got filters %q", document.Filters)
	}
	summary := document.Sections[0]
	if summary.Title != "Ringkasan" || fieldValue(t, summary, "Skor tertinggi") != "Kandidat Jakarta Barat dan Lahan Cibubur (1) dengan skor 78 dari 100" {
		t.Fatalf("got summary %v", summary.Fields)
	}

	table := sectionTitled(t, document, "Perbandingan per lokasi")
	headers := make([]string, 0, len(table.Columns))
	for _, column := range table.Columns {
		headers = append(headers, column.Header)
	}
	if got := strings.Join(headers, "|"); got != "Aspek|Kandidat Jakarta Barat|Lahan Cibubur (1)|Lahan Cibubur (2)" {
		t.Fatalf("got columns %s", got)
	}
	if len(table.Rows) != 1+len(reportDimensions)+2 {
		t.Fatalf("got %d rows", len(table.Rows))
	}
	rows := map[string][]any{}
	for _, row := range table.Rows {
		rows[row[0].(string)] = row[1:]
	}
	checks := map[string][]any{
		"Skor keseluruhan (0–100)": {int64(78), int64(78), int64(64)},
		"Pasar & Kompetisi":        {int64(64), int64(58), missingValue},
		"Penanda risiko (jumlah)":  {int64(1), int64(0), notRecorded},
		"Harga tanah ZNT (Rp/m²)":  {int64(8580000), int64(8580000), notAvailable},
	}
	for label, want := range checks {
		got := rows[label]
		for index := range want {
			if got[index] != want[index] {
				t.Errorf("%s: got %v, want %v", label, got, want)
				break
			}
		}
	}
	if !strings.Contains(strings.Join(table.Paragraphs, " "), comparisonRiskGapText) {
		t.Fatalf("unrecorded risk flags must be explained")
	}
}

func TestComparisonDocumentWithoutLocations(t *testing.T) {
	report := comparedReport()
	report.Locations, report.Scores = nil, nil
	document := newComparisonDocument(report)
	if got := sectionTitles(document); got != "Ringkasan|Cara membaca skor" {
		t.Fatalf("got sections %s", got)
	}
	if fieldValue(t, document.Sections[0], "Skor tertinggi") != "Belum ada lokasi" {
		t.Fatalf("got %v", document.Sections[0].Fields)
	}
}

func TestReportFileNamesUseSlugAndDate(t *testing.T) {
	cases := []struct {
		got, want string
	}{
		{siteFileName("Kandidat RSIA Harjamukti, Cirebon", reportGeneratedAt), "analisis-lokasi-kandidat-rsia-harjamukti-cirebon-2026-10-15.pdf"},
		{siteFileName("  ¿¿  ", reportGeneratedAt), "analisis-lokasi-2026-10-15.pdf"},
		{comparisonFileName("Perbandingan: Jawa Barat (Tahap 2)", reportGeneratedAt), "perbandingan-lokasi-perbandingan-jawa-barat-tahap-2-2026-10-15.pdf"},
		{siteFileName(strings.Repeat("Lahan ", 30), reportGeneratedAt), "analisis-lokasi-" + strings.TrimSuffix(strings.Repeat("lahan-", 10), "-") + "-2026-10-15.pdf"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("got %q, want %q", tc.got, tc.want)
		}
	}
	lateEvening := time.Date(2026, 10, 14, 20, 0, 0, 0, time.UTC)
	service := &reportService{now: func() time.Time { return lateEvening }}
	if got := service.generatedAt().Format(fileDateLayout); got != "2026-10-15" {
		t.Fatalf("file date must follow WIB, got %s", got)
	}
}

func sampleComparisonReport() comparisonReport {
	report := comparedReport()
	fourth, scores := reportLocation("Kandidat RSIA Harjamukti, Cirebon (lahan eks gudang di tepi jalan provinsi)", 81,
		[]int{74, 69, 90, 77, 58}, stored(`[{"code":"lereng_curam","severity":"medium","message":"Kemiringan lahan sekitar 16,2°, perlu kajian geoteknik."}]`))
	fourth.LandPricePerSqm = 2750000
	report.Locations = append(report.Locations, fourth)
	report.Scores = append(report.Scores, scores...)
	return report
}

func TestReportSamplesRenderAsPDF(t *testing.T) {
	samples := map[string]export.Document{
		"site-report-sample.pdf":     newSiteDocument(fullSiteReport()),
		"site-comparison-sample.pdf": newComparisonDocument(sampleComparisonReport()),
	}
	directory := os.Getenv(reportSamplesVariable)
	for name, document := range samples {
		var content bytes.Buffer
		if err := export.WritePDF(&content, document); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !bytes.HasPrefix(content.Bytes(), []byte("%PDF-")) {
			t.Fatalf("%s is not a PDF", name)
		}
		if bytes.IndexByte(content.Bytes(), '?') >= 0 {
			t.Fatalf("%s has characters that could not be encoded", name)
		}
		if directory == "" {
			continue
		}
		if err := os.WriteFile(filepath.Join(directory, name), content.Bytes(), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}
