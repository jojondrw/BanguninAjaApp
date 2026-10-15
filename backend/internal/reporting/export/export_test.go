package export

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

var generatedAt = time.Date(2026, 10, 3, 9, 30, 0, 0, time.UTC)

func sampleDocument(rows int) Document {
	section := Section{
		Title: "Transaksi kas",
		Columns: []Column{
			{Header: "Tanggal", Kind: Date},
			{Header: "Keterangan", Kind: Text},
			{Header: "Jumlah (Rp)", Kind: Money},
			{Header: "Volume", Kind: Decimal},
			{Header: "Serapan", Kind: Percent},
		},
	}
	for index := 1; index <= rows; index++ {
		section.Rows = append(section.Rows, []any{
			time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), fmt.Sprintf("UJI-%04d Pembelian (semen) \"Tiga Roda\"", index),
			int64(index * 1500000), 12.5, int64(45),
		})
	}
	section.Total = []any{nil, "Total", int64(rows * 1500000), nil, nil}
	return Document{
		Title:       "Laporan Arus Kas – Oktober 2026",
		Filters:     "Proyek: UJI Cibubur · Periode: 01/10/2026 – 31/10/2026",
		GeneratedAt: generatedAt,
		Sections:    []Section{section, {Title: "Ringkasan: bulan/periode", Columns: []Column{{Header: "Bulan", Kind: Text}}}},
	}
}

func TestCSVStartsWithBOMAndUsesSemicolons(t *testing.T) {
	var output bytes.Buffer
	if err := WriteCSV(&output, sampleDocument(2)); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	if !bytes.HasPrefix(output.Bytes(), []byte("\xEF\xBB\xBF")) {
		t.Fatalf("csv must start with a UTF-8 BOM")
	}
	reader := csv.NewReader(bytes.NewReader(output.Bytes()[3:]))
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("read csv back: %v", err)
	}
	if records[0][0] != "Laporan Arus Kas – Oktober 2026" {
		t.Fatalf("got title %q", records[0][0])
	}
	if records[3][0] != "Transaksi kas" {
		t.Fatalf("got section title %q", records[3])
	}
	header := strings.Join(records[4], "|")
	if header != "Tanggal|Keterangan|Jumlah (Rp)|Volume|Serapan" {
		t.Fatalf("got header %q", header)
	}
	first := records[5]
	if first[0] != "2026-10-01" || first[1] != `UJI-0001 Pembelian (semen) "Tiga Roda"` || first[2] != "1500000" || first[3] != "12,5" || first[4] != "45" {
		t.Fatalf("got row %q", first)
	}
	if total := records[7]; total[1] != "Total" || total[2] != "3000000" {
		t.Fatalf("got total %q", total)
	}
}

func TestXLSXIsAZipWithTheRequiredParts(t *testing.T) {
	var output bytes.Buffer
	if err := WriteXLSX(&output, sampleDocument(3)); err != nil {
		t.Fatalf("write xlsx: %v", err)
	}
	archive, err := zip.NewReader(bytes.NewReader(output.Bytes()), int64(output.Len()))
	if err != nil {
		t.Fatalf("xlsx is not a zip: %v", err)
	}
	parts := map[string]string{}
	for _, file := range archive.File {
		parts[file.Name] = readZipFile(t, file)
	}
	for _, name := range []string{
		"[Content_Types].xml", "_rels/.rels", "xl/workbook.xml", "xl/_rels/workbook.xml.rels",
		"xl/styles.xml", "xl/worksheets/sheet1.xml", "xl/worksheets/sheet2.xml",
	} {
		content, ok := parts[name]
		if !ok {
			t.Fatalf("missing part %s", name)
		}
		assertWellFormedXML(t, name, content)
	}
	workbook := parts["xl/workbook.xml"]
	if !strings.Contains(workbook, `name="Transaksi kas"`) || !strings.Contains(workbook, `name="Ringkasan- bulan-periode"`) {
		t.Fatalf("sheet names not sanitized: %s", workbook)
	}
	sheet := parts["xl/worksheets/sheet1.xml"]
	for _, fragment := range []string{
		`<c r="A5" t="inlineStr" s="1">`,
		`<c r="C6" s="2"><v>1500000</v></c>`,
		`<c r="A6" s="4"><v>46296</v></c>`,
		`<c r="D6" s="3"><v>12.5</v></c>`,
		`<c r="C9" s="8"><v>4500000</v></c>`,
		`<col min="2" max="2"`,
		`state="frozen"`,
	} {
		if !strings.Contains(sheet, fragment) {
			t.Errorf("sheet1 is missing %s", fragment)
		}
	}
}

func readZipFile(t *testing.T, file *zip.File) string {
	t.Helper()
	reader, err := file.Open()
	if err != nil {
		t.Fatalf("open %s: %v", file.Name, err)
	}
	defer func() { _ = reader.Close() }()
	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read %s: %v", file.Name, err)
	}
	return string(content)
}

func assertWellFormedXML(t *testing.T, name, content string) {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(content))
	for {
		_, err := decoder.Token()
		if err == io.EOF {
			return
		}
		if err != nil {
			t.Fatalf("%s is not well-formed XML: %v", name, err)
		}
	}
}

func TestPDFSinglePageForShortTable(t *testing.T) {
	data := writePDF(t, sampleDocument(3))
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatalf("pdf must start with %%PDF-")
	}
	if pages := pageCount(t, data); pages != 1 {
		t.Fatalf("got %d pages", pages)
	}
	assertValidXref(t, data)
	if !bytes.Contains(data, []byte("(Halaman 1 dari 1)")) {
		t.Fatalf("missing page number")
	}
}

func TestPDFBreaksLongTablesAcrossPages(t *testing.T) {
	const rows = 300
	data := writePDF(t, sampleDocument(rows))
	pages := pageCount(t, data)
	available := a4LongSide - 2*pageMargin - tableBottom
	minimumPages := int(float64(rows)*rowHeight/available) + 1
	if pages < minimumPages || pages > minimumPages+1 {
		t.Fatalf("got %d pages, want %d or %d", pages, minimumPages, minimumPages+1)
	}
	assertValidXref(t, data)
	for index := 1; index <= rows; index++ {
		marker := fmt.Sprintf("(UJI-%04d ", index)
		if count := bytes.Count(data, []byte(marker)); count != 1 {
			t.Fatalf("row %d drawn %d times", index, count)
		}
	}
	if !bytes.Contains(data, []byte(fmt.Sprintf("(Halaman %d dari %d)", pages, pages))) {
		t.Fatalf("missing last page number")
	}
	if !bytes.Contains(data, []byte("(Transaksi kas \\(lanjutan\\))")) {
		t.Fatalf("continued pages must repeat the section title")
	}
}

func TestPDFEncodesTextAsWinAnsi(t *testing.T) {
	data := writePDF(t, sampleDocument(1))
	if !bytes.Contains(data, []byte("(Laporan Arus Kas \x96 Oktober 2026)")) {
		t.Fatalf("en dash must be encoded as WinAnsi 0x96")
	}
	if !bytes.Contains(data, []byte("/BaseFont /Helvetica /Encoding /WinAnsiEncoding")) {
		t.Fatalf("missing Helvetica WinAnsi font")
	}
}

func writePDF(t *testing.T, document Document) []byte {
	t.Helper()
	var output bytes.Buffer
	if err := WritePDF(&output, document); err != nil {
		t.Fatalf("write pdf: %v", err)
	}
	return output.Bytes()
}

func pageCount(t *testing.T, data []byte) int {
	t.Helper()
	objects := bytes.Count(data, []byte("/Type /Page /Parent"))
	match := regexp.MustCompile(`/Type /Pages /Kids \[[^\]]*\] /Count (\d+)`).FindSubmatch(data)
	if match == nil {
		t.Fatalf("missing page tree")
	}
	declared, _ := strconv.Atoi(string(match[1]))
	if declared != objects {
		t.Fatalf("page tree declares %d pages but has %d page objects", declared, objects)
	}
	return objects
}

func assertValidXref(t *testing.T, data []byte) {
	t.Helper()
	match := regexp.MustCompile(`startxref\n(\d+)\n%%EOF\n$`).FindSubmatch(data)
	if match == nil {
		t.Fatalf("missing startxref trailer")
	}
	start, _ := strconv.Atoi(string(match[1]))
	if !bytes.HasPrefix(data[start:], []byte("xref\n0 ")) {
		t.Fatalf("startxref does not point at the xref table")
	}
	lines := strings.Split(string(data[start:]), "\n")
	size, _ := strconv.Atoi(strings.Fields(lines[1])[1])
	for object := 1; object < size; object++ {
		entry := lines[2+object]
		if len(entry)+1 != 20 {
			t.Fatalf("xref entry %d is %d bytes", object, len(entry)+1)
		}
		offset, _ := strconv.Atoi(entry[:10])
		if want := fmt.Sprintf("%d 0 obj\n", object); !bytes.HasPrefix(data[offset:], []byte(want)) {
			t.Fatalf("xref entry %d points at %q", object, data[offset:offset+12])
		}
	}
}
