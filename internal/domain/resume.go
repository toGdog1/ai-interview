package domain

import "time"

// FileFormat represents supported resume upload formats.
type FileFormat string

const (
	FormatTXT  FileFormat = "txt"
	FormatMD   FileFormat = "md"
	FormatPDF  FileFormat = "pdf"
	FormatDOCX FileFormat = "docx"
)

// BasicInfo holds personal contact information extracted from a resume.
type BasicInfo struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Phone    string   `json:"phone"`
	Location string   `json:"location"`
	Links    []string `json:"links"`
}

// Skills groups technical skills by category.
type Skills struct {
	Languages  []string `json:"languages"`
	Frameworks []string `json:"frameworks"`
	Tools      []string `json:"tools"`
	Other      []string `json:"other"`
}

// WorkExperience represents one employment entry.
type WorkExperience struct {
	Company    string   `json:"company"`
	Title      string   `json:"title"`
	Duration   string   `json:"duration"`
	Highlights []string `json:"highlights"`
}

// Project represents an independent or open-source project.
type Project struct {
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	TechStack    []string `json:"techStack"`
	Description  string   `json:"description"`
	Achievements []string `json:"achievements"`
}

// Education represents schooling background.
type Education struct {
	School   string `json:"school"`
	Degree   string `json:"degree"`
	Major    string `json:"major"`
	Duration string `json:"duration"`
}

// StructuredResume is the canonical parsed representation of a resume.
type StructuredResume struct {
	BasicInfo      BasicInfo        `json:"basicInfo"`
	Summary        string           `json:"summary"`
	Skills         Skills           `json:"skills"`
	WorkExperience []WorkExperience `json:"workExperience"`
	Projects       []Project        `json:"projects"`
	Education      []Education      `json:"education"`
	Certifications []string         `json:"certifications"`
	Others         []string         `json:"others"`
}

// Resume is the persisted resume aggregate.
type Resume struct {
	ID           string           `json:"id"`
	UserID       string           `json:"userId,omitempty"`
	RawContent   string           `json:"rawContent,omitempty"`
	Format       FileFormat       `json:"format"`
	Structured   StructuredResume `json:"structured"`
	TargetRole   string           `json:"targetRole,omitempty"`
	TargetJD     string           `json:"targetJd,omitempty"`
	YearsOfExp   int              `json:"yearsOfExp,omitempty"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
}
