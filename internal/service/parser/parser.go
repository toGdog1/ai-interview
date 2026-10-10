package parser

import (
	"context"
	"fmt"
	"strings"

	"github.com/cuixiangyu/ai-interview/internal/domain"
)

// TextExtractor extracts plain text from a specific file format.
// Implement this interface to add PDF/DOCX support without changing the Parser.
type TextExtractor interface {
	Format() domain.FileFormat
	Extract(ctx context.Context, data []byte) (string, error)
}

// StructureBuilder turns raw text into a StructuredResume.
// Rule-based and LLM-based implementations can be swapped.
type StructureBuilder interface {
	Build(ctx context.Context, text string) (domain.StructuredResume, error)
}

// Parser orchestrates format detection, text extraction, and structuring.
type Parser struct {
	extractors map[domain.FileFormat]TextExtractor
	builder    StructureBuilder
	maxBytes   int64
}

// New creates a Parser with the given extractors and structure builder.
func New(builder StructureBuilder, maxBytes int64, extractors ...TextExtractor) *Parser {
	m := make(map[domain.FileFormat]TextExtractor, len(extractors))
	for _, e := range extractors {
		m[e.Format()] = e
	}
	if maxBytes <= 0 {
		maxBytes = 10 * 1024 * 1024
	}
	return &Parser{extractors: m, builder: builder, maxBytes: maxBytes}
}

// ParseInput is the input for parsing a resume upload or paste.
type ParseInput struct {
	Content []byte
	Format  domain.FileFormat
	Text    string // when pasting plain text directly
}

// Parse extracts and structures resume content.
func (p *Parser) Parse(ctx context.Context, in ParseInput) (string, domain.StructuredResume, error) {
	var text string

	if strings.TrimSpace(in.Text) != "" {
		text = strings.TrimSpace(in.Text)
	} else {
		if int64(len(in.Content)) > p.maxBytes {
			return "", domain.StructuredResume{}, domain.ErrFileTooLarge
		}
		if len(in.Content) == 0 {
			return "", domain.StructuredResume{}, domain.ErrEmptyContent
		}
		format := in.Format
		if format == "" {
			format = DetectFormat(in.Content)
		}
		ext, ok := p.extractors[format]
		if !ok {
			return "", domain.StructuredResume{}, fmt.Errorf("%w: %s", domain.ErrUnsupportedFormat, format)
		}
		extracted, err := ext.Extract(ctx, in.Content)
		if err != nil {
			return "", domain.StructuredResume{}, fmt.Errorf("%w: %v", domain.ErrParseFailed, err)
		}
		text = strings.TrimSpace(extracted)
	}

	if text == "" {
		return "", domain.StructuredResume{}, domain.ErrEmptyContent
	}

	structured, err := p.builder.Build(ctx, text)
	if err != nil {
		return "", domain.StructuredResume{}, fmt.Errorf("%w: %v", domain.ErrParseFailed, err)
	}
	return text, structured, nil
}

// DetectFormat heuristically detects file format from content / magic bytes.
func DetectFormat(data []byte) domain.FileFormat {
	if len(data) >= 4 && string(data[:4]) == "%PDF" {
		return domain.FormatPDF
	}
	if len(data) >= 2 && data[0] == 0x50 && data[1] == 0x4B {
		return domain.FormatDOCX
	}
	sample := string(data)
	if len(sample) > 200 {
		sample = sample[:200]
	}
	if strings.Contains(sample, "# ") || strings.Contains(sample, "## ") {
		return domain.FormatMD
	}
	return domain.FormatTXT
}
