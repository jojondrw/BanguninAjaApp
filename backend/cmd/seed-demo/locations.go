package main

import (
	"log/slog"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/location"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/site"
)

const (
	evaluatedNote = "Dinilai otomatis lewat layanan penilaian lokasi saat data demo dibuat."
	sampleNote    = "Nilai contoh: layanan penilaian lokasi tidak tersedia saat data demo dibuat, jadi skor, indeks risiko, dan harga tanah ini bukan hasil analisis nyata."

	comparisonName = "Perbandingan kandidat lahan perumahan Jawa Barat"
)

var scoredDimensions = []string{"fisik_lingkungan", "infrastruktur", "demografi_sosial", "pasar_kompetisi", "finansial_proyek"}

type siteSeed struct {
	name      string
	region    string
	profile   string
	project   string
	latitude  float64
	longitude float64
	area      float64
	landPrice int64
	flood     float64
	quake     float64
	scores    []int
}

func demoSites() []siteSeed {
	return []siteSeed{
		{"Kandidat perumahan Gedebage, Bandung", "3273", "housing", projectGriyaAsri, -6.9536, 107.6915, 38_000, 4_250_000, 0.42, 0.31, []int{64, 78, 81, 74, 70}},
		{"Kandidat perumahan Cimahi Utara", "3273", "housing", "", -6.8540, 107.5425, 21_500, 3_600_000, 0.18, 0.36, []int{79, 71, 75, 68, 73}},
		{"Kandidat perumahan Sawangan, Depok", "3276", "housing", "", -6.4027, 106.7688, 26_000, 5_900_000, 0.22, 0.24, []int{76, 82, 84, 61, 58}},
		{"Kandidat RSIA Harjamukti, Cirebon", "3274", "hospital", projectHospital, -6.7465, 108.5562, 9_800, 2_750_000, 0.27, 0.19, []int{72, 80, 77, 83, 69}},
		{"Kandidat ruko Panakkukang, Makassar", "7371", "mall", projectRuko, -5.1561, 119.4466, 4_200, 7_800_000, 0.35, 0.12, []int{68, 85, 79, 72, 64}},
		{"Kandidat pusat belanja Surabaya Barat", "3578", "mall", "", -7.2887, 112.6762, 15_500, 9_400_000, 0.29, 0.21, []int{71, 77, 80, 59, 62}},
		{"Kandidat taman hiburan Marina, Semarang", "3374", "entertainment", "", -6.9528, 110.3897, 52_000, 3_100_000, 0.64, 0.17, []int{48, 70, 66, 63, 57}},
	}
}

func comparisonSites() []string {
	return []string{"Kandidat perumahan Gedebage, Bandung", "Kandidat perumahan Cimahi Utara", "Kandidat perumahan Sawangan, Depok"}
}

func demoSiteNames() []string {
	names := make([]string, 0, len(demoSites()))
	for _, seed := range demoSites() {
		names = append(names, seed.name)
	}
	return names
}

func (s *seeder) seedLocations() error {
	for _, seed := range demoSites() {
		if err := s.seedSite(seed); err != nil {
			return err
		}
	}
	return s.seedComparison()
}

func (s *seeder) seedSite(seed siteSeed) error {
	id, _, err := s.ensure("saved_location", map[string]any{"user_id": s.owner, "name": seed.name}, func() (uuid.UUID, error) {
		if s.services.sites != nil {
			if id, ok := s.evaluateSite(seed); ok {
				return id, nil
			}
		}
		return s.sampleSite(seed)
	})
	if err != nil {
		return err
	}
	s.remember("saved_location", seed.name, id)
	return nil
}

func (s *seeder) evaluateSite(seed siteSeed) (uuid.UUID, bool) {
	profileID, projectID, err := s.siteReferences(seed)
	if err != nil {
		return uuid.Nil, false
	}
	latitude, longitude := seed.latitude, seed.longitude
	response, err := s.services.sites.Evaluate(s.ctx, s.owner, site.EvaluateRequest{
		ProjectID: projectID, Latitude: &latitude, Longitude: &longitude,
		BuildingProfileID: profileID, Name: seed.name,
	})
	if err != nil {
		slog.Warn("score service failed, falling back to sample values", slog.String("site", seed.name), slog.String("error", err.Error()))
		s.services.sites = nil
		return uuid.Nil, false
	}
	regionID, err := s.reference("region", seed.region)
	if err != nil {
		return uuid.Nil, false
	}
	details := locationDetails{note: evaluatedNote, regionID: &regionID, areaSqm: seed.area}
	if err := s.patcher.describeLocation(s.ctx, response.SavedLocationID, details); err != nil {
		return uuid.Nil, false
	}
	s.tally.add("saved_location (evaluated)")
	return response.SavedLocationID, true
}

func (s *seeder) siteReferences(seed siteSeed) (uuid.UUID, *uuid.UUID, error) {
	profileID, err := s.reference("building_profile", seed.profile)
	if err != nil {
		return uuid.Nil, nil, err
	}
	projectID, err := s.knownPointer("project", seed.project)
	return profileID, projectID, err
}

func (s *seeder) sampleSite(seed siteSeed) (uuid.UUID, error) {
	profileID, projectID, err := s.siteReferences(seed)
	if err != nil {
		return uuid.Nil, err
	}
	regionID, err := s.reference("region", seed.region)
	if err != nil {
		return uuid.Nil, err
	}
	scores, overall, err := s.sampleScores(seed.scores)
	if err != nil {
		return uuid.Nil, err
	}
	latitude, longitude := seed.latitude, seed.longitude
	response, err := s.services.locations.CreateSavedLocation(s.ctx, s.owner, location.SavedLocationRequest{
		Name: seed.name, RegionID: &regionID, BuildingProfileID: &profileID, ProjectID: projectID,
		Latitude: &latitude, Longitude: &longitude, AreaSqm: seed.area, LandPricePerSqm: seed.landPrice,
		Score: overall, FloodIndex: seed.flood, EarthquakeIndex: seed.quake, Note: sampleNote, DimensionScores: scores,
	})
	if err != nil {
		return uuid.Nil, err
	}
	s.tally.add("saved_location (sample)")
	return response.ID, s.patcher.markSimulated(s.ctx, response.ID)
}

func (s *seeder) sampleScores(values []int) ([]location.DimensionScoreRequest, int, error) {
	scores := make([]location.DimensionScoreRequest, 0, len(values))
	total := 0
	for index, value := range values {
		dimensionID, err := s.reference("dimension", scoredDimensions[index])
		if err != nil {
			return nil, 0, err
		}
		scores = append(scores, location.DimensionScoreRequest{DimensionID: dimensionID, Value: value})
		total += value
	}
	return scores, total / len(values), nil
}

func (s *seeder) seedComparison() error {
	ids := make([]uuid.UUID, 0, len(comparisonSites()))
	for _, name := range comparisonSites() {
		id, err := s.known("saved_location", name)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}
	_, _, err := s.ensure("comparison", map[string]any{"user_id": s.owner, "name": comparisonName}, func() (uuid.UUID, error) {
		response, err := s.services.locations.CreateComparison(s.ctx, s.owner, location.ComparisonRequest{Name: comparisonName, SavedLocationIDs: ids})
		return response.ID, err
	})
	return err
}
