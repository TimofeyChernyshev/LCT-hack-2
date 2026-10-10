package resume

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	streamRegex = regexp.MustCompile(`(?s)stream\r?\n(.*?)\r?\nendstream`)
	tjRegex     = regexp.MustCompile(`\((.*?)\)\s*Tj`)
	tjArrayItem = regexp.MustCompile(`\((.*?)\)`)
)

// ExtractTextFromPDF extracts readable text from PDF binary data or plain text
func ExtractTextFromPDF(data []byte) (string, error) {
	if len(data) == 0 {
		return "", errors.New("empty file content")
	}

	// 1. If not a PDF header, treat as plain UTF-8 text
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		if utf8.Valid(data) {
			return string(data), nil
		}
	}

	var sb strings.Builder

	// 2. Extract streams
	matches := streamRegex.FindAllSubmatch(data, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		rawStream := match[1]

		// Try decompressing Flate/zlib stream
		decompressed := decompressStream(rawStream)
		streamText := extractTextFromStream(decompressed)
		if len(streamText) > 0 {
			sb.WriteString(streamText)
			sb.WriteString("\n")
		}
	}

	res := strings.TrimSpace(sb.String())

	// 3. Fallback: if no text found via operators, search literal strings directly in PDF body
	if len(res) < 20 {
		fallbackText := extractLiteralStringsFallback(data)
		if len(fallbackText) > len(res) {
			res = fallbackText
		}
	}

	// Clean up multi-line whitespace
	lines := strings.Split(res, "\n")
	var cleanLines []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if len(trimmed) > 0 {
			cleanLines = append(cleanLines, trimmed)
		}
	}

	return strings.Join(cleanLines, "\n"), nil
}

func decompressStream(data []byte) []byte {
	// Try zlib decompression
	zr, err := zlib.NewReader(bytes.NewReader(data))
	if err == nil {
		defer zr.Close()
		decompressed, err := io.ReadAll(zr)
		if err == nil && len(decompressed) > 0 {
			return decompressed
		}
	}

	// If not zlib compressed, return as is
	return data
}

func extractTextFromStream(stream []byte) string {
	var sb strings.Builder

	// 1. Process array text operators: [(...) 120 (...)] TJ
	tjMatches := tjRegex.FindAllSubmatch(stream, -1)
	for _, m := range tjMatches {
		if len(m) > 1 {
			decoded := decodePDFString(m[1])
			if len(decoded) > 0 {
				sb.WriteString(decoded)
				sb.WriteString("\n")
			}
		}
	}

	// 2. Process TJ array operators
	tjBlocks := extractTJBlocks(stream)
	for _, block := range tjBlocks {
		items := tjArrayItem.FindAllSubmatch(block, -1)
		for _, item := range items {
			if len(item) > 1 {
				decoded := decodePDFString(item[1])
				sb.WriteString(decoded)
			}
		}
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String())
}

func extractTJBlocks(stream []byte) [][]byte {
	var blocks [][]byte
	startIdx := 0
	for {
		openBracket := bytes.Index(stream[startIdx:], []byte("["))
		if openBracket == -1 {
			break
		}
		pos := startIdx + openBracket
		closeBracket := bytes.Index(stream[pos:], []byte("]"))
		if closeBracket == -1 {
			break
		}
		end := pos + closeBracket + 1

		// Verify followed by TJ
		after := bytes.TrimSpace(stream[end:])
		if bytes.HasPrefix(after, []byte("TJ")) {
			blocks = append(blocks, stream[pos:end])
		}
		startIdx = end
	}
	return blocks
}

func decodePDFString(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}

	// 1. Handle UTF-16BE (Byte Order Mark: \xFE\xFF)
	if len(raw) >= 2 && raw[0] == 0xFE && raw[1] == 0xFF {
		var runes []rune
		for i := 2; i < len(raw)-1; i += 2 {
			r := rune(raw[i])<<8 | rune(raw[i+1])
			runes = append(runes, r)
		}
		return string(runes)
	}

	// 2. Handle escape sequences
	var sb strings.Builder
	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if ch == '\\' && i+1 < len(raw) {
			next := raw[i+1]
			switch next {
			case 'n':
				sb.WriteByte('\n')
				i++
			case 'r':
				sb.WriteByte('\r')
				i++
			case 't':
				sb.WriteByte('\t')
				i++
			case '(', ')', '\\':
				sb.WriteByte(next)
				i++
			default:
				// Octal escape \ddd
				if next >= '0' && next <= '7' {
					octLen := 1
					for octLen < 3 && i+1+octLen < len(raw) && raw[i+1+octLen] >= '0' && raw[i+1+octLen] <= '7' {
						octLen++
					}
					octStr := string(raw[i+1 : i+1+octLen])
					if val, err := strconv.ParseInt(octStr, 8, 32); err == nil {
						sb.WriteByte(byte(val))
					}
					i += octLen
				} else {
					sb.WriteByte(next)
					i++
				}
			}
		} else {
			sb.WriteByte(ch)
		}
	}

	strBytes := []byte(sb.String())
	if utf8.Valid(strBytes) {
		return string(strBytes)
	}

	// Fallback: Windows-1251 decode for Russian characters
	return decodeCP1251(strBytes)
}

func decodeCP1251(data []byte) string {
	var sb strings.Builder
	for _, b := range data {
		if b < 128 {
			sb.WriteByte(b)
		} else if b >= 192 && b <= 255 {
			// А..Я, а..я in CP1251: 192..255 -> 0x0410..0x044F
			r := rune(0x0410 + int(b) - 192)
			sb.WriteRune(r)
		} else if b == 168 { // Ё
			sb.WriteRune('Ё')
		} else if b == 184 { // ё
			sb.WriteRune('ё')
		} else {
			sb.WriteByte('?')
		}
	}
	return sb.String()
}

func extractLiteralStringsFallback(data []byte) string {
	var sb strings.Builder
	inParen := false
	var current strings.Builder

	for i := 0; i < len(data); i++ {
		ch := data[i]
		if ch == '(' && (i == 0 || data[i-1] != '\\') {
			inParen = true
			current.Reset()
		} else if ch == ')' && (i == 0 || data[i-1] != '\\') && inParen {
			inParen = false
			txt := strings.TrimSpace(current.String())
			if len(txt) > 2 && !strings.HasPrefix(txt, "/") {
				sb.WriteString(txt)
				sb.WriteString(" ")
			}
			current.Reset()
		} else if inParen {
			current.WriteByte(ch)
		}
	}

	return strings.TrimSpace(sb.String())
}
