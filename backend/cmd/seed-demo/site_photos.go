package main

import (
	"embed"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
)

//go:embed photos/*.jpg
var demoPhotoFiles embed.FS

const demoPhotoNote = "Ilustrasi: "

type sitePhotoSeed struct {
	project string
	file    string
	taken   int
	caption string
}

func demoSitePhotos() []sitePhotoSeed {
	return []sitePhotoSeed{
		{projectGriyaAsri, "gab_1.jpg", -120, "galian dan pondasi rumah Blok A"},
		{projectGriyaAsri, "gab_2.jpg", -6, "dinding bata Blok A sampai ring balok"},
		{projectHospital, "rsia_1.jpg", -75, "bekisting dan pengecoran kolom lantai 2"},
		{projectHospital, "rsia_2.jpg", -4, "struktur beton sampai lantai 4, dinding lantai dasar dimulai"},
		{projectRuko, "rsm_1.jpg", -30, "pengukuran batas lahan"},
		{projectRuko, "rsm_2.jpg", -10, "uji sondir tanah"},
		{projectCluster, "cld_1.jpg", -150, "pengecatan fasad dan paving jalan cluster"},
		{projectCluster, "cld_2.jpg", -100, "cluster selesai, siap serah terima"},
	}
}

func (s *seeder) seedSitePhotos() error {
	for _, seed := range demoSitePhotos() {
		projectID, err := s.known("project", seed.project)
		if err != nil {
			return err
		}
		content, err := demoPhotoFiles.ReadFile("photos/" + seed.file)
		if err != nil {
			return err
		}
		caption := demoPhotoNote + seed.caption
		taken := s.day(seed.taken)
		_, _, err = s.ensure("site_photo", map[string]any{"project_id": projectID, "caption": caption}, func() (uuid.UUID, error) {
			response, err := s.services.projects.CreateSitePhoto(s.ctx, projectID, s.owner, project.SitePhotoUpload{
				Caption: caption, TakenOn: &taken, Content: content,
			})
			return response.ID, err
		})
		if err != nil {
			return err
		}
	}
	return nil
}
