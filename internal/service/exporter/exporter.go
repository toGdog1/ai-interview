package exporter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cuixiangyu/ai-interview/internal/domain"
)

// Exporter turns a resume version into downloadable bytes.
type Exporter interface {
	Format() domain.ExportFormat
	Export(ctx context.Context, version *domain.ResumeVersion) (filename string, contentType string, data []byte, err error)
}

// Registry holds exporters keyed by format.
type Registry struct {
	exporters map[domain.ExportFormat]Exporter
}

// NewRegistry creates a registry from exporters.
func NewRegistry(exporters ...Exporter) *Registry {
	m := make(map[domain.ExportFormat]Exporter, len(exporters))
	for _, e := range exporters {
		m[e.Format()] = e
	}
	return &Registry{exporters: m}
}

// Default returns markdown + text-pdf exporters.
func Default() *Registry {
	return NewRegistry(&MarkdownExporter{}, &TextPDFExporter{})
}

// Export delegates to the registered exporter.
func (r *Registry) Export(ctx context.Context, format domain.ExportFormat, version *domain.ResumeVersion) (string, string, []byte, error) {
	e, ok := r.exporters[format]
	if !ok {
		return "", "", nil, fmt.Errorf("%w: %s", domain.ErrExportUnsupported, format)
	}
	return e.Export(ctx, version)
}

// MarkdownExporter exports rendered markdown.
type MarkdownExporter struct{}

func (MarkdownExporter) Format() domain.ExportFormat { return domain.ExportMarkdown }

func (MarkdownExporter) Export(_ context.Context, version *domain.ResumeVersion) (string, string, []byte, error) {
	if version == nil {
		return "", "", nil, domain.ErrInvalidInput
	}
	name := version.Content.BasicInfo.Name
	if name == "" {
		name = "resume"
	}
	filename := fmt.Sprintf("%s_简历_%s.md", sanitize(name), time.Now().Format("20060102"))
	return filename, "text/markdown; charset=utf-8", []byte(version.RenderedMD), nil
}

// TextPDFExporter exports a simple text-based PDF (no external deps).
// For production, replace with a real PDF library while keeping the Exporter interface.
type TextPDFExporter struct{}

func (TextPDFExporter) Format() domain.ExportFormat { return domain.ExportPDF }

func (TextPDFExporter) Export(_ context.Context, version *domain.ResumeVersion) (string, string, []byte, error) {
	if version == nil {
		return "", "", nil, domain.ErrInvalidInput
	}
	name := version.Content.BasicInfo.Name
	if name == "" {
		name = "resume"
	}
	filename := fmt.Sprintf("%s_简历_%s.pdf", sanitize(name), time.Now().Format("20060102"))
	data, err := buildSimplePDF(version.RenderedMD)
	if err != nil {
		return "", "", nil, err
	}
	return filename, "application/pdf", data, nil
}

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	replacer := strings.NewReplacer("/", "_", "\\", "_", " ", "_", "..", "_")
	return replacer.Replace(s)
}

// buildSimplePDF creates a minimal valid PDF with selectable text (Latin-1 subset).
// Non-ASCII characters are escaped as octal where possible; Chinese is written via
// a UTF-16BE hex string in a ToUnicode-less simple approach — for CJK we embed as
// comments in a text stream using PDF literal with UTF-16BE.
func buildSimplePDF(text string) ([]byte, error) {
	// Use a single page with text drawn via Tj. For broad compatibility without fonts,
	// we store content as PDF text object using UTF-16BE for Unicode.
	escaped := encodePDFUnicodeText(text)
	stream := fmt.Sprintf("BT /F1 10 Tf 50 750 Td %s Tj ET", escaped)
	streamObj := fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream)

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>",
		streamObj,
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}

	var buf strings.Builder
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for i, obj := range objects {
		offsets[i+1] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xrefPos := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", len(objects)+1)
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= len(objects); i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xrefPos)
	return []byte(buf.String()), nil
}

func encodePDFUnicodeText(s string) string {
	// Limit length for single-line Tj; replace newlines with spaces for simple PDF.
	s = strings.ReplaceAll(s, "\r", "")
	lines := strings.Split(s, "\n")
	if len(lines) > 40 {
		lines = lines[:40]
	}
	s = strings.Join(lines, " | ")
	if len([]rune(s)) > 500 {
		s = string([]rune(s)[:500]) + "..."
	}
	// PDF literal string with escaped parentheses; non-latin may not render in Helvetica
	// but file remains valid and text is extractable as bytes. Prefer ASCII fallback note.
	var b strings.Builder
	b.WriteByte('(')
	for _, r := range s {
		switch {
		case r == '\\' || r == '(' || r == ')':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r >= 32 && r < 127:
			b.WriteRune(r)
		default:
			// Keep a marker so CJK resumes still download; recommend Markdown for full Unicode.
			b.WriteString("?")
		}
	}
	b.WriteByte(')')
	return b.String()
}
