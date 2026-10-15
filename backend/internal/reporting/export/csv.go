package export

import (
	"encoding/csv"
	"io"
)

const (
	CSVSeparator = ';'
	utf8BOM      = string(rune(0xFEFF))
)

func WriteCSV(output io.Writer, document Document) error {
	if _, err := io.WriteString(output, utf8BOM); err != nil {
		return err
	}
	writer := csv.NewWriter(output)
	writer.Comma = CSVSeparator
	writer.UseCRLF = true

	records := [][]string{{document.Title}, {document.Filters}, {document.GeneratedLine()}}
	for _, section := range document.Sections {
		records = append(records, []string{})
		records = append(records, sectionRecords(section)...)
	}
	if err := writer.WriteAll(records); err != nil {
		return err
	}
	return writer.Error()
}

func sectionRecords(section Section) [][]string {
	records := [][]string{{section.Title}, headerRecord(section.Columns)}
	for _, row := range section.Rows {
		records = append(records, valueRecord(row))
	}
	if len(section.Total) > 0 {
		records = append(records, valueRecord(section.Total))
	}
	return records
}

func headerRecord(columns []Column) []string {
	headers := make([]string, len(columns))
	for index, column := range columns {
		headers[index] = column.Header
	}
	return headers
}

func valueRecord(values []any) []string {
	record := make([]string, len(values))
	for index, value := range values {
		record[index] = MachineText(value)
	}
	return record
}
