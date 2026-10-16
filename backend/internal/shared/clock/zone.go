package clock

import (
	"fmt"
	"os"
	"time"
	_ "time/tzdata"
)

const (
	zoneEnv     = "APP_TIMEZONE"
	defaultZone = "Asia/Jakarta"
)

func UseAppZone() (string, error) {
	name := os.Getenv(zoneEnv)
	if name == "" {
		name = defaultZone
	}

	location, err := time.LoadLocation(name)
	if err != nil {
		return "", fmt.Errorf("load %s %q: %w", zoneEnv, name, err)
	}

	time.Local = location
	return name, nil
}
