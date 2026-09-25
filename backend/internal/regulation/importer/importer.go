package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"regexp"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/regulation"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/geo"
	"gorm.io/gorm"
)

type rawRecord struct {
	RTR    string   `json:"rtr"`
	IDRTR  string   `json:"id_rtr"`
	IDWlyh string   `json:"idwlyh"`

	Namzon string `json:"namzon"`
	Kodzon string `json:"kodzon"`
	Namszn string `json:"namszn"`
	Kodszn string `json:"kodszn"`

	Wadmkc string `json:"wadmkc"`
	Wadmkd string `json:"wadmkd"`

	KDB []json.RawMessage `json:"kdb"`
	KLB []json.RawMessage `json:"klb"`
	KDH []json.RawMessage `json:"kdh"`
	GSB json.RawMessage `json:"gsb"`
	Note string   `json:"nothpr"`

	Lat float64 `json:"_lat"`
	Lon float64 `json:"_lon"`
}

func Import(ctx context.Context, db *gorm.DB, dir string, selectedFile string) error {
	files, err := filepath.Glob(filepath.Join(dir, "rdtr_*.json"))
if err != nil {
	return fmt.Errorf("find RDTR files: %w", err)
}

if len(files) != 10 {
	return fmt.Errorf("expected 10 RDTR files, found %d", len(files))
}

if selectedFile != "" {
	var selectedPath string

	for _, path := range files {
		if filepath.Base(path) == selectedFile {
			selectedPath = path
			break
		}
	}

	if selectedPath == "" {
		return fmt.Errorf("RDTR file %q not found", selectedFile)
	}

	return importFile(ctx, db, selectedPath)
}

for _, path := range files {
	if err := importFile(ctx, db, path); err != nil {
		return fmt.Errorf("import %s: %w", filepath.Base(path), err)
	}
}

return nil
}

func importFile(ctx context.Context, db *gorm.DB, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("read JSON start: %w", err)
	}

	delim, ok := token.(json.Delim)
	if !ok || delim != '[' {
		return fmt.Errorf("expected JSON array")
	}

	count := 0

	for decoder.More() {
		var r rawRecord

		if err := decoder.Decode(&r); err != nil {
			return fmt.Errorf("decode record %d: %w", count, err)
		}

		kdb, err := parseFloat(r.KDB)
		if err != nil {
			return fmt.Errorf("record %d KDB: %w", count, err)
		}

		klb, err := parseFloat(r.KLB)
		if err != nil {
			return fmt.Errorf("record %d KLB: %w", count, err)
		}

		kdh, err := parseFloat(r.KDH)
		if err != nil {
			return fmt.Errorf("record %d KDH: %w", count, err)
		}

		item := regulation.Regulation{
			RTRID:          r.IDRTR,
			RegionCode:     r.IDWlyh,
			ZoneName:       r.Namzon,
			ZoneCode:       r.Kodzon,
			SubZoneName:    r.Namszn,
			SubZoneCode:    r.Kodszn,
			District:       r.Wadmkc,
			Village:        r.Wadmkd,
			KDB:            kdb,
			KLB:            klb,
			KDH:            kdh,
			GSB: string(r.GSB),
			RegulationNote: r.Note,
			Point: geo.Point{
				Lon: r.Lon,
				Lat: r.Lat,
			},
			IsSimulated: false,
		}

		if err := db.WithContext(ctx).Create(&item).Error; err != nil {
			return fmt.Errorf("record %d insert: %w", count, err)
		}

		count++

		if count%500 == 0 {
			fmt.Printf("%s: %d records imported\n", filepath.Base(path), count)
		}
	}

	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("read JSON end: %w", err)
	}

	fmt.Printf("imported %s: %d records\n", filepath.Base(path), count)

	return nil
}

func firstValue(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func parseFloat(values []json.RawMessage) (float64, error) {
	if len(values) == 0 {
		return 0, nil
	}

	return parseRawValue(values[0])
}

func parseRawValue(raw json.RawMessage) (float64, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		s = strings.TrimSpace(s)

		if s == "" || s == "-" || s == "- %" {
			return 0, nil
		}

		if strings.EqualFold(s, "Tidak ada bangunan") {
    		return 0, nil
		}

		s = strings.ReplaceAll(s, ",", ".")

		if result, err := strconv.ParseFloat(s, 64); err == nil {
			return result, nil
		}

		re := regexp.MustCompile(`[-+]?\d+(?:\.\d+)?`)
		match := re.FindString(s)
		if match == "" {
			return 0, fmt.Errorf("invalid value %q", s)
		}

		result, err := strconv.ParseFloat(match, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid value %q", s)
		}

		return result, nil
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err == nil {
		for _, value := range obj {
			var arr []json.RawMessage
			if err := json.Unmarshal(value, &arr); err == nil {
				for _, item := range arr {
					if result, err := parseRawValue(item); err == nil {
						return result, nil
					}
				}
			}

			if result, err := parseRawValue(value); err == nil {
				return result, nil
			}
		}
	}

	return 0, fmt.Errorf("unsupported value %s", string(raw))
}