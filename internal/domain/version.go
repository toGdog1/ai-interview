package domain

import "time"

// TemplateID identifies a resume rendering template.
type TemplateID string

const (
	TemplateATS    TemplateID = "ats"
	TemplateTech   TemplateID = "tech"
	TemplateClassic TemplateID = "classic"
)

// ExportFormat is the download format for a generated resume.
type ExportFormat string

const (
	ExportMarkdown ExportFormat = "markdown"
	ExportPDF      ExportFormat = "pdf"
	ExportDOCX     ExportFormat = "docx"
)

// ResumeVersion is a generated optimized resume snapshot.
type ResumeVersion struct {
	ID            string           `json:"id"`
	ResumeID      string           `json:"resumeId"`
	Template      TemplateID       `json:"template"`
	Content       StructuredResume `json:"content"`
	RenderedMD    string           `json:"renderedMd"`
	SuggestionIDs []string         `json:"suggestionIds"`
	CreatedAt     time.Time        `json:"createdAt"`
}
