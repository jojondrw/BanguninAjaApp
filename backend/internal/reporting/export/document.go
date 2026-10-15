package export

import (
	"strconv"
	"strings"
	"time"
)

type Kind string

const (
	Text    Kind = "text"
	Money   Kind = "money"
	Integer Kind = "integer"
	Decimal Kind = "decimal"
	Percent Kind = "percent"
	Date    Kind = "date"
)

const (
	dateLayout      = "2006-01-02"
	displayDate     = "02/01/2006"
	displayDateTime = "02/01/2006 15:04"
	thousandsGroup  = 3
)

type Column struct {
	Header string `json:"header"`
	Kind   Kind   `json:"kind"`
}

type Section struct {
	Title   string   `json:"title"`
	Columns []Column `json:"columns"`
	Rows    [][]any  `json:"rows"`
	Total   []any    `json:"total,omitempty"`
}

type Document struct {
	Title       string
	Filters     string
	GeneratedAt time.Time
	Sections    []Section
}

func (d Document) GeneratedLine() string {
	return "Dibuat " + d.GeneratedAt.Format(displayDateTime)
}

func (c Column) IsNumeric() bool {
	return c.Kind == Money || c.Kind == Integer || c.Kind == Decimal || c.Kind == Percent
}

func DisplayText(kind Kind, value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case time.Time:
		return typed.Format(displayDate)
	case int64:
		return displayInteger(kind, typed)
	case int:
		return displayInteger(kind, int64(typed))
	case float64:
		return displayDecimal(kind, typed)
	default:
		return ""
	}
}

func MachineText(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case time.Time:
		return typed.Format(dateLayout)
	case int64:
		return strconv.FormatInt(typed, 10)
	case int:
		return strconv.Itoa(typed)
	case float64:
		return strings.Replace(strconv.FormatFloat(typed, 'f', -1, 64), ".", ",", 1)
	default:
		return ""
	}
}

func displayInteger(kind Kind, value int64) string {
	text := groupThousands(value)
	if kind == Percent {
		return text + "%"
	}
	return text
}

func displayDecimal(kind Kind, value float64) string {
	whole := int64(value)
	fraction := value - float64(whole)
	text := groupThousands(whole)
	if fraction != 0 {
		decimals := strconv.FormatFloat(absolute(fraction), 'f', 2, 64)
		text += "," + strings.TrimRight(decimals[2:], "0")
	}
	if kind == Percent {
		return text + "%"
	}
	return text
}

func absolute(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func groupThousands(value int64) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	digits := strconv.FormatInt(value, 10)
	var builder strings.Builder
	for index, digit := range digits {
		if index > 0 && (len(digits)-index)%thousandsGroup == 0 {
			builder.WriteByte('.')
		}
		builder.WriteRune(digit)
	}
	return sign + builder.String()
}
