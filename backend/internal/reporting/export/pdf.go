package export

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	a4ShortSide = 595.28
	a4LongSide  = 841.89

	pageMargin       = 36.0
	footerBaseline   = 20.0
	tableBottom      = 40.0
	titleSize        = 14.0
	filterSize       = 9.0
	sectionTitleSize = 11.0
	cellSize         = 8.0
	footerSize       = 7.0
	titleLeading     = 20.0
	filterLeading    = 13.0
	sectionLeading   = 18.0
	headerHeight     = 16.0
	rowHeight        = 13.0
	sectionGap       = 14.0
	cellPadding      = 4.0
	minimumColumn    = 34.0
	measuredRows     = 300
	textBaselineLift = 4.0
	fixedShareLimit  = 0.7
	textColumnFloor  = 60.0
	cellLeading      = 10.0
	wrappedColumnCap = 140.0
	bodySize         = 9.0
	bodyLeading      = 13.0
	fieldPadding     = 2.0
	fieldLabelFloor  = 90.0
	fieldLabelCap    = 170.0
	fieldLabelGap    = 12.0
	paragraphGap     = 4.0

	pdfObjectsBeforePages = 5
	continuedSuffix       = " (lanjutan)"
	emptySectionText      = "Tidak ada data untuk filter ini."
)

type pdfPage struct {
	content bytes.Buffer
}

type tablePlan struct {
	widths     []float64
	header     [][][]byte
	headerBand float64
}

type pdfLayout struct {
	width  float64
	height float64
	pages  []*pdfPage
	cursor float64
}

func WritePDF(output io.Writer, document Document) error {
	layout := newPDFLayout(document)
	layout.drawHeading(document)
	for _, section := range document.Sections {
		layout.drawSection(section)
	}
	_, err := output.Write(layout.render(document))
	return err
}

func newPDFLayout(document Document) *pdfLayout {
	layout := &pdfLayout{width: a4ShortSide, height: a4LongSide}
	if needsLandscape(document) {
		layout.width, layout.height = a4LongSide, a4ShortSide
	}
	layout.addPage()
	return layout
}

func needsLandscape(document Document) bool {
	available := a4ShortSide - 2*pageMargin
	for _, section := range document.Sections {
		if sumWidths(desiredWidths(section)) > available {
			return true
		}
	}
	return false
}

func (l *pdfLayout) addPage() {
	l.pages = append(l.pages, &pdfPage{})
	l.cursor = l.height - pageMargin
}

func (l *pdfLayout) page() *pdfPage {
	return l.pages[len(l.pages)-1]
}

func (l *pdfLayout) usableWidth() float64 {
	return l.width - 2*pageMargin
}

func (l *pdfLayout) drawHeading(document Document) {
	l.cursor -= titleSize
	l.text(helveticaBold, titleSize, pageMargin, l.cursor, helveticaBold.fit(document.Title, titleSize, l.usableWidth()))
	l.cursor -= titleLeading - titleSize + filterSize
	if document.Filters != "" {
		l.text(helvetica, filterSize, pageMargin, l.cursor, helvetica.fit(document.Filters, filterSize, l.usableWidth()))
		l.cursor -= filterLeading
	}
	l.text(helvetica, filterSize, pageMargin, l.cursor, encodeWinAnsi(document.GeneratedLine()))
	l.cursor -= sectionGap
}

func (l *pdfLayout) drawSection(section Section) {
	if !hasContent(section) {
		return
	}
	plan := l.planTable(section)
	if l.cursor-sectionLeading-openingHeight(section, plan) < tableBottom {
		l.addPage()
	}
	l.drawTitle(section.Title)
	l.drawFields(section)
	l.drawTable(section, plan)
	l.drawParagraphs(section)
	l.cursor -= sectionGap
}

func hasContent(section Section) bool {
	return len(section.Columns) > 0 || len(section.Fields) > 0 || len(section.Paragraphs) > 0
}

func openingHeight(section Section, plan tablePlan) float64 {
	if len(section.Fields) > 0 || len(section.Columns) == 0 {
		return fieldPadding + bodyLeading
	}
	return plan.headerBand + rowHeight
}

func (l *pdfLayout) planTable(section Section) tablePlan {
	if len(section.Columns) == 0 {
		return tablePlan{}
	}
	widths := fitWidths(section.Columns, desiredWidths(section), l.usableWidth())
	header := make([][][]byte, len(section.Columns))
	for index, column := range section.Columns {
		header[index] = cellLines(helveticaBold, column.Header, column.Wrap, widths[index])
	}
	return tablePlan{widths: widths, header: header, headerBand: bandHeight(headerHeight, header)}
}

func (l *pdfLayout) drawTitle(title string) {
	l.cursor -= sectionTitleSize
	l.text(helveticaBold, sectionTitleSize, pageMargin, l.cursor, helveticaBold.fit(title, sectionTitleSize, l.usableWidth()))
	l.cursor -= sectionLeading - sectionTitleSize
}

func (l *pdfLayout) drawTable(section Section, plan tablePlan) {
	if len(section.Columns) == 0 {
		return
	}
	if l.cursor-plan.headerBand-rowHeight < tableBottom {
		l.addPage()
		l.drawTitle(section.Title + continuedSuffix)
	}
	l.drawHeader(section.Columns, plan)
	for _, row := range section.Rows {
		cells := rowCells(section.Columns, plan.widths, row, helvetica)
		l.ensureRowSpace(section, plan, bandHeight(rowHeight, cells))
		l.drawRow(section.Columns, plan.widths, cells, helvetica)
	}
	if len(section.Total) > 0 {
		cells := rowCells(section.Columns, plan.widths, section.Total, helveticaBold)
		l.ensureRowSpace(section, plan, bandHeight(rowHeight, cells))
		l.line(pageMargin, l.cursor, pageMargin+sumWidths(plan.widths), l.cursor, "0.2 G 0.8 w")
		l.drawRow(section.Columns, plan.widths, cells, helveticaBold)
	}
	if len(section.Rows) == 0 && len(section.Total) == 0 {
		l.cursor -= rowHeight
		l.text(helvetica, cellSize, pageMargin+cellPadding, l.cursor+textBaselineLift, encodeWinAnsi(emptySectionText))
	}
}

func (l *pdfLayout) ensureRowSpace(section Section, plan tablePlan, height float64) {
	if l.cursor-height >= tableBottom {
		return
	}
	l.addPage()
	l.drawTitle(section.Title + continuedSuffix)
	l.drawHeader(section.Columns, plan)
}

func (l *pdfLayout) drawHeader(columns []Column, plan tablePlan) {
	top := l.cursor
	l.cursor -= plan.headerBand
	fmt.Fprintf(&l.page().content, "0.886 0.910 0.941 rg %s %s %s %s re f 0 g\n",
		number(pageMargin), number(l.cursor), number(sumWidths(plan.widths)), number(top-l.cursor))
	l.drawCells(columns, plan.widths, plan.header, helveticaBold, top-headerHeight+textBaselineLift)
}

func (l *pdfLayout) drawRow(columns []Column, widths []float64, cells [][][]byte, font pdfFont) {
	top := l.cursor
	l.cursor -= bandHeight(rowHeight, cells)
	l.drawCells(columns, widths, cells, font, top-rowHeight+textBaselineLift)
	l.line(pageMargin, l.cursor, pageMargin+sumWidths(widths), l.cursor, "0.85 G 0.4 w")
}

func (l *pdfLayout) drawCells(columns []Column, widths []float64, cells [][][]byte, font pdfFont, baseline float64) {
	x := pageMargin
	for index, column := range columns {
		l.drawCellLines(font, cells[index], column.IsNumeric(), x, widths[index], baseline)
		x += widths[index]
	}
}

func (l *pdfLayout) drawCellLines(font pdfFont, lines [][]byte, alignRight bool, x, width, baseline float64) {
	for index, line := range lines {
		left := x + cellPadding
		if alignRight {
			left = x + width - cellPadding - font.width(line, cellSize)
		}
		l.text(font, cellSize, left, baseline-float64(index)*cellLeading, line)
	}
}

func (l *pdfLayout) drawFields(section Section) {
	if len(section.Fields) == 0 {
		return
	}
	labelWidth := fieldLabelWidth(section.Fields)
	for _, field := range section.Fields {
		labels := helveticaBold.wrap(field.Label, bodySize, labelWidth-fieldLabelGap)
		values := helvetica.wrap(field.Value, bodySize, l.usableWidth()-labelWidth)
		l.cursor -= fieldPadding
		for index := range max(len(labels), len(values)) {
			l.ensureTextSpace(section.Title, bodyLeading)
			l.cursor -= bodyLeading
			l.bodyLine(helveticaBold, labels, index, pageMargin)
			l.bodyLine(helvetica, values, index, pageMargin+labelWidth)
		}
		l.cursor -= fieldPadding
		l.line(pageMargin, l.cursor, pageMargin+l.usableWidth(), l.cursor, "0.88 G 0.4 w")
	}
}

func (l *pdfLayout) drawParagraphs(section Section) {
	for _, paragraph := range section.Paragraphs {
		l.cursor -= paragraphGap
		for _, line := range helvetica.wrap(paragraph, bodySize, l.usableWidth()) {
			l.ensureTextSpace(section.Title, bodyLeading)
			l.cursor -= bodyLeading
			l.text(helvetica, bodySize, pageMargin, l.cursor+textBaselineLift, line)
		}
	}
}

func (l *pdfLayout) bodyLine(font pdfFont, lines [][]byte, index int, x float64) {
	if index < len(lines) {
		l.text(font, bodySize, x, l.cursor+textBaselineLift, lines[index])
	}
}

func (l *pdfLayout) ensureTextSpace(title string, height float64) {
	if l.cursor-height >= tableBottom {
		return
	}
	l.addPage()
	l.drawTitle(title + continuedSuffix)
}

func fieldLabelWidth(fields []Field) float64 {
	widest := 0.0
	for _, field := range fields {
		widest = max(widest, helveticaBold.width(encodeWinAnsi(field.Label), bodySize))
	}
	return min(max(widest+fieldLabelGap, fieldLabelFloor), fieldLabelCap)
}

func rowCells(columns []Column, widths []float64, values []any, font pdfFont) [][][]byte {
	cells := make([][][]byte, len(columns))
	for index, column := range columns {
		if index < len(values) {
			cells[index] = cellLines(font, DisplayText(column.Kind, values[index]), column.Wrap, widths[index])
		}
	}
	return cells
}

func cellLines(font pdfFont, text string, wrap bool, width float64) [][]byte {
	limit := width - 2*cellPadding
	if wrap {
		return font.wrap(text, cellSize, limit)
	}
	return [][]byte{font.fit(text, cellSize, limit)}
}

func bandHeight(base float64, cells [][][]byte) float64 {
	lines := 1
	for _, cell := range cells {
		lines = max(lines, len(cell))
	}
	return base + float64(lines-1)*cellLeading
}

func (l *pdfLayout) text(font pdfFont, size, x, y float64, encoded []byte) {
	content := &l.page().content
	fmt.Fprintf(content, "BT /%s %s Tf 1 0 0 1 %s %s Tm (", font.resource, number(size), number(x), number(y))
	content.Write(escapePDFString(encoded))
	content.WriteString(") Tj ET\n")
}

func (l *pdfLayout) line(x1, y1, x2, y2 float64, stroke string) {
	fmt.Fprintf(&l.page().content, "%s %s %s m %s %s l S 0 G\n", stroke, number(x1), number(y1), number(x2), number(y2))
}

func (l *pdfLayout) drawFooter(page *pdfPage, title string, index, total int) {
	label := encodeWinAnsi(fmt.Sprintf("Halaman %d dari %d", index, total))
	fmt.Fprintf(&page.content, "BT /F1 %s Tf 1 0 0 1 %s %s Tm (", number(footerSize), number(pageMargin), number(footerBaseline))
	page.content.Write(escapePDFString(helvetica.fit(title, footerSize, l.usableWidth()/2)))
	page.content.WriteString(") Tj ET\n")
	right := l.width - pageMargin - helvetica.width(label, footerSize)
	fmt.Fprintf(&page.content, "BT /F1 %s Tf 1 0 0 1 %s %s Tm (", number(footerSize), number(right), number(footerBaseline))
	page.content.Write(escapePDFString(label))
	page.content.WriteString(") Tj ET\n")
}

func (l *pdfLayout) render(document Document) []byte {
	for index, page := range l.pages {
		l.drawFooter(page, document.Title, index+1, len(l.pages))
	}
	writer := &pdfWriter{}
	writer.buffer.WriteString("%PDF-1.4\n%\xE2\xE3\xCF\xD3\n")
	writer.object("<< /Type /Catalog /Pages 2 0 R >>")
	writer.object(l.pagesDictionary())
	writer.object("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>")
	writer.object("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>")
	writer.object("<< /Title (" + string(escapePDFString(encodeWinAnsi(document.Title))) + ") /Producer (BanguninAja) >>")
	for index, page := range l.pages {
		contentID := pageObjectID(index) + 1
		writer.object(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %s %s] /Resources << /Font << /F1 3 0 R /F2 4 0 R >> >> /Contents %d 0 R >>",
			number(l.width), number(l.height), contentID))
		writer.stream(page.content.Bytes())
	}
	return writer.finish()
}

func (l *pdfLayout) pagesDictionary() string {
	kids := make([]string, len(l.pages))
	for index := range l.pages {
		kids[index] = fmt.Sprintf("%d 0 R", pageObjectID(index))
	}
	return fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), len(l.pages))
}

func pageObjectID(index int) int {
	return pdfObjectsBeforePages + 1 + 2*index
}

type pdfWriter struct {
	buffer  bytes.Buffer
	offsets []int
}

func (w *pdfWriter) object(body string) {
	w.begin()
	w.buffer.WriteString(body)
	w.buffer.WriteString("\nendobj\n")
}

func (w *pdfWriter) stream(content []byte) {
	w.begin()
	fmt.Fprintf(&w.buffer, "<< /Length %d >>\nstream\n", len(content))
	w.buffer.Write(content)
	w.buffer.WriteString("\nendstream\nendobj\n")
}

func (w *pdfWriter) begin() {
	w.offsets = append(w.offsets, w.buffer.Len())
	fmt.Fprintf(&w.buffer, "%d 0 obj\n", len(w.offsets))
}

func (w *pdfWriter) finish() []byte {
	start := w.buffer.Len()
	fmt.Fprintf(&w.buffer, "xref\n0 %d\n0000000000 65535 f \n", len(w.offsets)+1)
	for _, offset := range w.offsets {
		fmt.Fprintf(&w.buffer, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&w.buffer, "trailer\n<< /Size %d /Root 1 0 R /Info 5 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(w.offsets)+1, start)
	return w.buffer.Bytes()
}

func desiredWidths(section Section) []float64 {
	widths := make([]float64, len(section.Columns))
	for index, column := range section.Columns {
		width := measure(helveticaBold, column.Header, column.Wrap)
		for rowIndex, row := range section.Rows {
			if rowIndex >= measuredRows {
				break
			}
			if index < len(row) {
				width = max(width, measure(helvetica, DisplayText(column.Kind, row[index]), column.Wrap))
			}
		}
		if index < len(section.Total) {
			width = max(width, measure(helveticaBold, DisplayText(column.Kind, section.Total[index]), column.Wrap))
		}
		widths[index] = max(width+2*cellPadding, minimumColumn)
	}
	return widths
}

func measure(font pdfFont, text string, wrap bool) float64 {
	full := font.width(encodeWinAnsi(text), cellSize)
	if !wrap {
		return full
	}
	return max(font.longestWord(text, cellSize), min(full, wrappedColumnCap))
}

func fitWidths(columns []Column, desired []float64, available float64) []float64 {
	fixed := 0.0
	for index, column := range columns {
		if column.Kind != Text {
			fixed += desired[index]
		}
	}
	textSpace := available - fixed
	if sumWidths(desired) <= available || fixed == sumWidths(desired) || textSpace < available*(1-fixedShareLimit) {
		return scaleWidths(desired, available/sumWidths(desired))
	}
	return shrinkTextColumns(columns, desired, textSpace)
}

func shrinkTextColumns(columns []Column, desired []float64, textSpace float64) []float64 {
	fitted := append([]float64(nil), desired...)
	floors, surplus := 0.0, 0.0
	for index, column := range columns {
		if column.Kind == Text {
			floors += min(desired[index], textColumnFloor)
			surplus += desired[index] - min(desired[index], textColumnFloor)
		}
	}
	for index, column := range columns {
		if column.Kind != Text {
			continue
		}
		floor := min(desired[index], textColumnFloor)
		if floors >= textSpace {
			fitted[index] = floor * textSpace / floors
			continue
		}
		fitted[index] = floor + (desired[index]-floor)*(textSpace-floors)/surplus
	}
	return fitted
}

func scaleWidths(widths []float64, scale float64) []float64 {
	scaled := make([]float64, len(widths))
	for index, width := range widths {
		scaled[index] = width * scale
	}
	return scaled
}

func sumWidths(widths []float64) float64 {
	total := 0.0
	for _, width := range widths {
		total += width
	}
	return total
}

func number(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func escapePDFString(encoded []byte) []byte {
	escaped := make([]byte, 0, len(encoded))
	for _, value := range encoded {
		if value == '(' || value == ')' || value == '\\' {
			escaped = append(escaped, '\\')
		}
		escaped = append(escaped, value)
	}
	return escaped
}
