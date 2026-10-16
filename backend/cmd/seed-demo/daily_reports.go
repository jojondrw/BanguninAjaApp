package main

import (
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
)

const dailyReportDays = 14

type dailyReportPlan struct {
	project   string
	crew      int
	equipment string
	works     []string
}

type dayWeather struct {
	morning   string
	afternoon string
	issue     string
}

// Oktober sudah masuk musim hujan: pagi biasanya cerah atau berawan, hujan
// turun siang atau sore. Pola tetap supaya data demo sama setiap dibuat ulang.
var demoWeather = []dayWeather{
	{"sunny", "cloudy", ""},
	{"cloudy", "rain", "Hujan pukul 14.00, pekerjaan luar berhenti sekitar 2 jam"},
	{"sunny", "sunny", ""},
	{"cloudy", "heavy_rain", "Hujan lebat sejak 13.30, pengecoran ditunda ke esok hari"},
	{"drizzle", "cloudy", ""},
	{"sunny", "drizzle", ""},
	{"cloudy", "rain", "Akses truk ke lokasi licin, pengiriman material mundur satu jam"},
}

func demoDailyReportPlans() []dailyReportPlan {
	return []dailyReportPlan{
		{
			project:   projectHospital,
			crew:      68,
			equipment: "1 tower crane, 1 concrete pump, 2 truk mixer",
			works: []string{
				"Pembesian dan bekisting kolom lantai 4 as A sampai C",
				"Pengecoran pelat lantai 4 zona timur",
				"Instalasi pipa gas medis lantai 2",
				"Pemasangan tray kabel dan jalur listrik lantai 1",
			},
		},
		{
			project:   projectGriyaAsri,
			crew:      41,
			equipment: "1 molen beton, 1 truk material, scaffolding 40 set",
			works: []string{
				"Pemasangan dinding bata ringan blok A nomor 1 sampai 8",
				"Pengecoran sloof blok B nomor 1 sampai 6",
				"Pemasangan rangka atap baja ringan blok A",
				"Pembersihan saluran drainase utama sebelum musim hujan",
			},
		},
	}
}

func (s *seeder) seedDailyReports() error {
	for _, plan := range demoDailyReportPlans() {
		projectID, known := s.ids["project:"+plan.project]
		if !known {
			continue
		}
		written := 0
		for offset := 1; written < dailyReportDays; offset++ {
			day := s.day(-offset)
			if day.Weekday() == time.Sunday {
				continue
			}
			weather := demoWeather[offset%len(demoWeather)]
			workers := plan.crew + (offset*7)%11 - 5
			if weather.afternoon == "heavy_rain" {
				workers -= 12
			}
			request := project.DailyReportRequest{
				ReportDate:       &day,
				WeatherMorning:   weather.morning,
				WeatherAfternoon: weather.afternoon,
				Workers:          &workers,
				Equipment:        plan.equipment,
				WorkDone:         plan.works[offset%len(plan.works)],
				Issues:           weather.issue,
			}
			_, _, err := s.ensure("daily_report", map[string]any{"project_id": projectID, "report_date": day}, func() (uuid.UUID, error) {
				response, err := s.services.projects.CreateDailyReport(s.ctx, projectID, s.owner, request)
				return response.ID, err
			})
			if err != nil {
				return err
			}
			written++
		}
	}
	return nil
}
