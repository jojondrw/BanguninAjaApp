package export

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	styleDefault = iota
	styleHeader
	styleMoney
	styleDecimal
	styleDate
	stylePercent
	styleTitle
	styleTotalText
	styleTotalMoney
	styleTotalDecimal
	styleTotalPercent
)

const (
	sheetNameLimit   = 31
	headerRowNumber  = 5
	firstDataRow     = headerRowNumber + 1
	minimumWidth     = 8
	maximumWidth     = 60
	widthPadding     = 2
	numericWidthSlot = 14
	lettersInAlpha   = 26
)

var excelEpoch = time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)

var sheetNameReplacer = strings.NewReplacer("[", "(", "]", ")", ":", "-", "*", "-", "?", "", "/", "-", `\`, "-")

const xmlHeader = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n"

const xlsxRootRelationships = xmlHeader + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
	`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>` +
	`<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>` +
	`</Relationships>`

const xlsxStyles = xmlHeader + `<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
	`<numFmts count="2"><numFmt numFmtId="164" formatCode="dd/mm/yyyy"/><numFmt numFmtId="165" formatCode="0&quot;%&quot;"/></numFmts>` +
	`<fonts count="3"><font><sz val="11"/><name val="Calibri"/></font><font><b/><sz val="11"/><name val="Calibri"/></font><font><b/><sz val="14"/><name val="Calibri"/></font></fonts>` +
	`<fills count="3"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill>` +
	`<fill><patternFill patternType="solid"><fgColor rgb="FFE2E8F0"/><bgColor indexed="64"/></patternFill></fill></fills>` +
	`<borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>` +
	`<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>` +
	`<cellXfs count="11">` +
	`<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>` +
	`<xf numFmtId="0" fontId="1" fillId="2" borderId="0" xfId="0" applyFont="1" applyFill="1"/>` +
	`<xf numFmtId="3" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/>` +
	`<xf numFmtId="4" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/>` +
	`<xf numFmtId="164" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/>` +
	`<xf numFmtId="165" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/>` +
	`<xf numFmtId="0" fontId="2" fillId="0" borderId="0" xfId="0" applyFont="1"/>` +
	`<xf numFmtId="0" fontId="1" fillId="0" borderId="0" xfId="0" applyFont="1"/>` +
	`<xf numFmtId="3" fontId="1" fillId="0" borderId="0" xfId="0" applyNumberFormat="1" applyFont="1"/>` +
	`<xf numFmtId="4" fontId="1" fillId="0" borderId="0" xfId="0" applyNumberFormat="1" applyFont="1"/>` +
	`<xf numFmtId="165" fontId="1" fillId="0" borderId="0" xfId="0" applyNumberFormat="1" applyFont="1"/>` +
	`</cellXfs>` +
	`<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>` +
	`</styleSheet>`

type xlsxPart struct {
	name    string
	content string
}

func WriteXLSX(output io.Writer, document Document) error {
	archive := zip.NewWriter(output)
	for _, part := range xlsxParts(document) {
		file, err := archive.Create(part.name)
		if err != nil {
			return err
		}
		if _, err := io.WriteString(file, part.content); err != nil {
			return err
		}
	}
	return archive.Close()
}

func xlsxParts(document Document) []xlsxPart {
	sections := document.Sections
	if len(sections) == 0 {
		sections = []Section{{Title: document.Title}}
	}
	names := sheetNames(sections)
	parts := []xlsxPart{
		{"[Content_Types].xml", contentTypes(len(sections))},
		{"_rels/.rels", xlsxRootRelationships},
		{"docProps/core.xml", coreProperties(document)},
		{"xl/workbook.xml", workbook(names)},
		{"xl/_rels/workbook.xml.rels", workbookRelationships(len(sections))},
		{"xl/styles.xml", xlsxStyles},
	}
	for index, section := range sections {
		parts = append(parts, xlsxPart{fmt.Sprintf("xl/worksheets/sheet%d.xml", index+1), worksheet(document, section)})
	}
	return parts
}

func contentTypes(sheets int) string {
	var builder strings.Builder
	builder.WriteString(xmlHeader + `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	builder.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	builder.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	builder.WriteString(`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`)
	builder.WriteString(`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>`)
	builder.WriteString(`<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>`)
	for index := 1; index <= sheets; index++ {
		fmt.Fprintf(&builder, `<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, index)
	}
	builder.WriteString(`</Types>`)
	return builder.String()
}

func coreProperties(document Document) string {
	created := document.GeneratedAt.UTC().Format(time.RFC3339)
	return xmlHeader + `<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" ` +
		`xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" ` +
		`xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">` +
		`<dc:title>` + escapeXML(document.Title) + `</dc:title><dc:creator>BanguninAja</dc:creator>` +
		`<dcterms:created xsi:type="dcterms:W3CDTF">` + created + `</dcterms:created>` +
		`</cp:coreProperties>`
}

func workbook(names []string) string {
	var builder strings.Builder
	builder.WriteString(xmlHeader + `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" ` +
		`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>`)
	for index, name := range names {
		fmt.Fprintf(&builder, `<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, escapeXML(name), index+1, index+1)
	}
	builder.WriteString(`</sheets></workbook>`)
	return builder.String()
}

func workbookRelationships(sheets int) string {
	var builder strings.Builder
	builder.WriteString(xmlHeader + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for index := 1; index <= sheets; index++ {
		fmt.Fprintf(&builder, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, index, index)
	}
	fmt.Fprintf(&builder, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`, sheets+1)
	builder.WriteString(`</Relationships>`)
	return builder.String()
}

func sheetNames(sections []Section) []string {
	names := make([]string, len(sections))
	used := map[string]bool{}
	for index, section := range sections {
		base := sanitizeSheetName(section.Title, index)
		name := base
		for suffix := 2; used[strings.ToLower(name)]; suffix++ {
			tail := fmt.Sprintf(" %d", suffix)
			name = truncateRunes(base, sheetNameLimit-len(tail)) + tail
		}
		used[strings.ToLower(name)] = true
		names[index] = name
	}
	return names
}

func sanitizeSheetName(title string, index int) string {
	name := strings.Trim(strings.TrimSpace(sheetNameReplacer.Replace(title)), "'")
	if name == "" {
		name = fmt.Sprintf("Bagian %d", index+1)
	}
	return truncateRunes(name, sheetNameLimit)
}

func truncateRunes(text string, limit int) string {
	if utf8.RuneCountInString(text) <= limit {
		return text
	}
	return string([]rune(text)[:limit])
}

func worksheet(document Document, section Section) string {
	var rows strings.Builder
	writeTextRow(&rows, 1, document.Title, styleTitle)
	writeTextRow(&rows, 2, document.Filters, styleDefault)
	writeTextRow(&rows, 3, document.GeneratedLine(), styleDefault)
	writeTextRow(&rows, 4, section.Title, styleTotalText)
	writeHeaderRow(&rows, section.Columns)
	for index, row := range section.Rows {
		writeValueRow(&rows, firstDataRow+index, section.Columns, row, false)
	}
	if len(section.Total) > 0 {
		writeValueRow(&rows, firstDataRow+len(section.Rows), section.Columns, section.Total, true)
	}

	var builder strings.Builder
	builder.WriteString(xmlHeader + `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" ` +
		`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	fmt.Fprintf(&builder, `<sheetViews><sheetView workbookViewId="0"><pane ySplit="%d" topLeftCell="A%d" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>`,
		headerRowNumber, firstDataRow)
	builder.WriteString(columnWidths(section))
	builder.WriteString(`<sheetData>`)
	builder.WriteString(rows.String())
	builder.WriteString(`</sheetData></worksheet>`)
	return builder.String()
}

func columnWidths(section Section) string {
	if len(section.Columns) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString(`<cols>`)
	for index, column := range section.Columns {
		fmt.Fprintf(&builder, `<col min="%d" max="%d" width="%d" customWidth="1"/>`, index+1, index+1, columnWidth(section, index, column))
	}
	builder.WriteString(`</cols>`)
	return builder.String()
}

func columnWidth(section Section, index int, column Column) int {
	width := utf8.RuneCountInString(column.Header)
	if column.IsNumeric() || column.Kind == Date {
		width = max(width, numericWidthSlot)
	}
	for _, row := range section.Rows {
		if index < len(row) {
			width = max(width, utf8.RuneCountInString(DisplayText(column.Kind, row[index])))
		}
	}
	return min(max(width+widthPadding, minimumWidth), maximumWidth)
}

func writeTextRow(builder *strings.Builder, number int, text string, style int) {
	fmt.Fprintf(builder, `<row r="%d">`, number)
	if text != "" {
		writeStringCell(builder, cellReference(0, number), text, style)
	}
	builder.WriteString(`</row>`)
}

func writeHeaderRow(builder *strings.Builder, columns []Column) {
	fmt.Fprintf(builder, `<row r="%d">`, headerRowNumber)
	for index, column := range columns {
		writeStringCell(builder, cellReference(index, headerRowNumber), column.Header, styleHeader)
	}
	builder.WriteString(`</row>`)
}

func writeValueRow(builder *strings.Builder, number int, columns []Column, values []any, isTotal bool) {
	fmt.Fprintf(builder, `<row r="%d">`, number)
	for index, value := range values {
		kind := Text
		if index < len(columns) {
			kind = columns[index].Kind
		}
		writeValueCell(builder, cellReference(index, number), kind, value, isTotal)
	}
	builder.WriteString(`</row>`)
}

func writeValueCell(builder *strings.Builder, reference string, kind Kind, value any, isTotal bool) {
	switch typed := value.(type) {
	case nil:
		return
	case string:
		writeStringCell(builder, reference, typed, textStyle(isTotal))
	case time.Time:
		serial := typed.Sub(excelEpoch).Hours() / float64(24)
		writeNumberCell(builder, reference, strconv.FormatFloat(serial, 'f', -1, 64), styleDate)
	case int64:
		writeNumberCell(builder, reference, strconv.FormatInt(typed, 10), numberStyle(kind, isTotal))
	case int:
		writeNumberCell(builder, reference, strconv.Itoa(typed), numberStyle(kind, isTotal))
	case float64:
		writeNumberCell(builder, reference, strconv.FormatFloat(typed, 'f', -1, 64), numberStyle(kind, isTotal))
	}
}

func textStyle(isTotal bool) int {
	if isTotal {
		return styleTotalText
	}
	return styleDefault
}

func numberStyle(kind Kind, isTotal bool) int {
	switch {
	case kind == Percent && isTotal:
		return styleTotalPercent
	case kind == Percent:
		return stylePercent
	case kind == Decimal && isTotal:
		return styleTotalDecimal
	case kind == Decimal:
		return styleDecimal
	case isTotal:
		return styleTotalMoney
	default:
		return styleMoney
	}
}

func writeStringCell(builder *strings.Builder, reference, text string, style int) {
	fmt.Fprintf(builder, `<c r="%s" t="inlineStr" s="%d"><is><t xml:space="preserve">%s</t></is></c>`, reference, style, escapeXML(text))
}

func writeNumberCell(builder *strings.Builder, reference, number string, style int) {
	fmt.Fprintf(builder, `<c r="%s" s="%d"><v>%s</v></c>`, reference, style, number)
}

func cellReference(column, row int) string {
	return columnLetters(column) + strconv.Itoa(row)
}

func columnLetters(index int) string {
	letters := ""
	for index >= 0 {
		letters = string(rune('A'+index%lettersInAlpha)) + letters
		index = index/lettersInAlpha - 1
	}
	return letters
}

func escapeXML(text string) string {
	var buffer bytes.Buffer
	_ = xml.EscapeText(&buffer, []byte(stripControlCharacters(text)))
	return buffer.String()
}

func stripControlCharacters(text string) string {
	return strings.Map(func(character rune) rune {
		if character < ' ' && character != '\t' && character != '\n' && character != '\r' {
			return -1
		}
		return character
	}, text)
}
