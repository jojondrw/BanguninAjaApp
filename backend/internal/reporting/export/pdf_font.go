package export

import "strings"

type pdfFont struct {
	resource string
	widths   [95]int
}

const (
	firstMeasuredByte = 32
	lastMeasuredByte  = 126
	defaultGlyphWidth = 556
	glyphUnits        = 1000.0
	ellipsisByte      = 0x85
	replacementByte   = '?'
	latinOneStart     = 0xA0
	latinOneEnd       = 0xFF
)

var helvetica = pdfFont{resource: "F1", widths: [95]int{
	278, 278, 355, 556, 556, 889, 667, 191, 333, 333, 389, 584, 278, 333, 278, 278,
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 278, 278, 584, 584, 584, 556,
	1015, 667, 667, 722, 722, 667, 611, 778, 722, 278, 500, 667, 556, 833, 722, 778,
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 278, 278, 278, 469, 556,
	333, 556, 556, 500, 556, 556, 278, 556, 556, 222, 222, 500, 222, 833, 556, 556,
	556, 556, 333, 500, 278, 556, 500, 722, 500, 500, 500, 334, 260, 334, 584,
}}

var helveticaBold = pdfFont{resource: "F2", widths: [95]int{
	278, 333, 474, 556, 556, 889, 722, 238, 333, 333, 389, 584, 278, 333, 278, 278,
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 333, 333, 584, 584, 584, 611,
	975, 722, 722, 722, 722, 667, 611, 778, 722, 278, 556, 722, 611, 833, 722, 778,
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 333, 278, 333, 584, 556,
	333, 556, 611, 556, 611, 556, 333, 611, 611, 278, 278, 556, 278, 889, 611, 611,
	611, 611, 389, 556, 333, 611, 556, 778, 556, 556, 500, 389, 280, 389, 584,
}}

var winAnsiExtras = map[rune]byte{
	'€': 0x80, '…': 0x85, '‘': 0x91, '’': 0x92, '“': 0x93, '”': 0x94,
	'•': 0x95, '–': 0x96, '—': 0x97, '−': '-',
}

var extraGlyphWidths = map[byte]int{0x85: 1000, 0x91: 222, 0x92: 222, 0x93: 333, 0x94: 333, 0x95: 350, 0x97: 1000}

func encodeWinAnsi(text string) []byte {
	encoded := make([]byte, 0, len(text))
	for _, character := range text {
		encoded = append(encoded, winAnsiByte(character))
	}
	return encoded
}

func winAnsiByte(character rune) byte {
	switch {
	case character == '\t' || character == '\n' || character == '\r':
		return ' '
	case character >= firstMeasuredByte && character <= lastMeasuredByte:
		return byte(character)
	case character >= latinOneStart && character <= latinOneEnd:
		return byte(character)
	}
	if mapped, ok := winAnsiExtras[character]; ok {
		return mapped
	}
	return replacementByte
}

func (f pdfFont) byteWidth(value byte) int {
	if value >= firstMeasuredByte && value <= lastMeasuredByte {
		return f.widths[value-firstMeasuredByte]
	}
	if width, ok := extraGlyphWidths[value]; ok {
		return width
	}
	return defaultGlyphWidth
}

func (f pdfFont) width(encoded []byte, size float64) float64 {
	total := 0
	for _, value := range encoded {
		total += f.byteWidth(value)
	}
	return float64(total) * size / glyphUnits
}

func (f pdfFont) wrap(text string, size, limit float64) [][]byte {
	var lines [][]byte
	for _, paragraph := range strings.Split(text, "\n") {
		lines = append(lines, f.wrapParagraph(paragraph, size, limit)...)
	}
	return lines
}

func (f pdfFont) wrapParagraph(text string, size, limit float64) [][]byte {
	var lines [][]byte
	var line []byte
	for _, word := range strings.Fields(text) {
		for _, piece := range f.splitWord(encodeWinAnsi(word), size, limit) {
			line, lines = f.place(line, lines, piece, size, limit)
		}
	}
	return append(lines, line)
}

func (f pdfFont) place(line []byte, lines [][]byte, piece []byte, size, limit float64) ([]byte, [][]byte) {
	if len(line) == 0 {
		return piece, lines
	}
	joined := make([]byte, 0, len(line)+1+len(piece))
	joined = append(append(append(joined, line...), ' '), piece...)
	if f.width(joined, size) <= limit {
		return joined, lines
	}
	return piece, append(lines, line)
}

func (f pdfFont) splitWord(word []byte, size, limit float64) [][]byte {
	var pieces [][]byte
	for len(word) > 1 && f.width(word, size) > limit {
		cut := f.fittingPrefix(word, size, limit)
		pieces = append(pieces, word[:cut])
		word = word[cut:]
	}
	return append(pieces, word)
}

func (f pdfFont) fittingPrefix(word []byte, size, limit float64) int {
	cut := 1
	for cut < len(word) && f.width(word[:cut+1], size) <= limit {
		cut++
	}
	return cut
}

func (f pdfFont) longestWord(text string, size float64) float64 {
	longest := 0.0
	for _, word := range strings.Fields(text) {
		longest = max(longest, f.width(encodeWinAnsi(word), size))
	}
	return longest
}

func (f pdfFont) fit(text string, size, limit float64) []byte {
	encoded := encodeWinAnsi(text)
	if f.width(encoded, size) <= limit {
		return encoded
	}
	ellipsis := float64(f.byteWidth(ellipsisByte)) * size / glyphUnits
	for len(encoded) > 0 && f.width(encoded, size)+ellipsis > limit {
		encoded = encoded[:len(encoded)-1]
	}
	return append(encoded, ellipsisByte)
}
