package parser

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/cuixiangyu/ai-interview/internal/domain"
)

// PlainTextExtractor handles .txt files.
type PlainTextExtractor struct{}

func (PlainTextExtractor) Format() domain.FileFormat { return domain.FormatTXT }

func (PlainTextExtractor) Extract(_ context.Context, data []byte) (string, error) {
	if !utf8.Valid(data) {
		return "", fmt.Errorf("content is not valid UTF-8")
	}
	return string(data), nil
}

// MarkdownExtractor handles .md files (treated as UTF-8 text).
type MarkdownExtractor struct{}

func (MarkdownExtractor) Format() domain.FileFormat { return domain.FormatMD }

func (MarkdownExtractor) Extract(_ context.Context, data []byte) (string, error) {
	if !utf8.Valid(data) {
		return "", fmt.Errorf("content is not valid UTF-8")
	}
	return string(data), nil
}

// PDFExtractor is a stub that rejects binary PDFs and guides users to paste text.
// Replace with a real PDF library (e.g. pdfcpu / ledongthuc/pdf) when ready.
type PDFExtractor struct{}

func (PDFExtractor) Format() domain.FileFormat { return domain.FormatPDF }

func (PDFExtractor) Extract(_ context.Context, data []byte) (string, error) {
	if len(data) >= 4 && string(data[:4]) == "%PDF" {
		return "", fmt.Errorf("PDF text extraction not enabled; please paste resume text or upload .txt/.md")
	}
	// Allow treating non-binary content labeled as pdf as text for testing.
	if utf8.Valid(data) {
		return string(data), nil
	}
	return "", fmt.Errorf("invalid PDF content")
}

// DOCXExtractor is a stub; swap for mammoth/docxlib later.
type DOCXExtractor struct{}

func (DOCXExtractor) Format() domain.FileFormat { return domain.FormatDOCX }

func (DOCXExtractor) Extract(_ context.Context, data []byte) (string, error) {
	if len(data) >= 2 && data[0] == 0x50 && data[1] == 0x4B {
		return "", fmt.Errorf("DOCX extraction not enabled; please paste resume text or upload .txt/.md")
	}
	if utf8.Valid(data) {
		return string(data), nil
	}
	return "", fmt.Errorf("invalid DOCX content")
}

// DefaultExtractors returns the built-in extractors for all supported formats.
func DefaultExtractors() []TextExtractor {
	return []TextExtractor{
		PlainTextExtractor{},
		MarkdownExtractor{},
		PDFExtractor{},
		DOCXExtractor{},
	}
}

// NormalizeWhitespace collapses excessive blank lines for cleaner parsing.
func NormalizeWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	blank := 0
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if strings.TrimSpace(trimmed) == "" {
			blank++
			if blank <= 1 {
				out = append(out, "")
			}
			continue
		}
		blank = 0
		out = append(out, trimmed)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
