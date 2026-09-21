package geo

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
)

const (
	sridWGS84         = 4326
	ewkbPointByteSize = 25
	typePoint         = 1
	sridFlag          = 0x20000000
)

type Point struct {
	Lon float64
	Lat float64
}

func (Point) GormDataType() string {
	return fmt.Sprintf("geometry(Point,%d)", sridWGS84)
}

func (p Point) Value() (driver.Value, error) {
	return fmt.Sprintf("SRID=%d;POINT(%g %g)", sridWGS84, p.Lon, p.Lat), nil
}

func (p *Point) Scan(value any) error {
	switch typed := value.(type) {
	case nil:
		return nil
	case []byte:
		return p.parse(string(typed))
	case string:
		return p.parse(typed)
	default:
		return fmt.Errorf("point: cannot read %T", value)
	}
}

func (p *Point) parse(raw string) error {
	decoded, err := hex.DecodeString(raw)
	if err != nil {
		return fmt.Errorf("point: value is not ewkb hex: %w", err)
	}
	if len(decoded) < ewkbPointByteSize {
		return fmt.Errorf("point: ewkb too short, got %d bytes", len(decoded))
	}

	order := binary.ByteOrder(binary.BigEndian)
	if decoded[0] == 1 {
		order = binary.LittleEndian
	}

	geometryType := order.Uint32(decoded[1:5])
	if geometryType&^uint32(sridFlag) != typePoint {
		return fmt.Errorf("point: geometry is not a point")
	}

	offset := 5
	if geometryType&sridFlag != 0 {
		offset += 4
	}

	p.Lon = math.Float64frombits(order.Uint64(decoded[offset : offset+8]))
	p.Lat = math.Float64frombits(order.Uint64(decoded[offset+8 : offset+16]))

	return nil
}
