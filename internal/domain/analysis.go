package domain

import "time"

// Priority indicates how urgent an issue or suggestion is.
type Priority string

const (
	PriorityP0 Priority = "P0"
	PriorityP1 Priority = "P1"
	PriorityP2 Priority = "P2"
)

// Grade maps overall score to letter grade.
type Grade string

const (
	GradeA Grade = "A"
	GradeB Grade = "B"
	GradeC Grade = "C"
	GradeD Grade = "D"
)

// DimensionScores holds the five scoring dimensions from the PRD.
type DimensionScores struct {
	Completeness float64 `json:"completeness"` // 内容完整性 20%
	TechExpress  float64 `json:"techExpress"`  // 技术表达 25%
	ProjectQual  float64 `json:"projectQual"`  // 项目质量 25%
	JDMatch      float64 `json:"jdMatch"`      // 岗位匹配 20%
	Readability  float64 `json:"readability"`  // 格式可读性 10%
}

// Issue describes a concrete problem found in the resume.
type Issue struct {
	ID          string   `json:"id"`
	Priority    Priority `json:"priority"`
	Module      string   `json:"module"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Location    string   `json:"location"`
}

// JDMatchReport summarizes keyword coverage against a job description.
type JDMatchReport struct {
	CoveredKeywords []string `json:"coveredKeywords"`
	MissingKeywords []string `json:"missingKeywords"`
	Suggestions     []string `json:"suggestions"`
	MatchScore      float64  `json:"matchScore"`
}

// Analysis is the AI analysis report for a resume.
type Analysis struct {
	ID              string          `json:"id"`
	ResumeID        string          `json:"resumeId"`
	OverallScore    float64         `json:"overallScore"`
	Grade           Grade           `json:"grade"`
	Summary         string          `json:"summary"`
	DimensionScores DimensionScores `json:"dimensionScores"`
	Issues          []Issue         `json:"issues"`
	Highlights      []string        `json:"highlights"`
	JDMatch         *JDMatchReport  `json:"jdMatch,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
}

// ScoreToGrade maps a 0-100 score to a letter grade.
func ScoreToGrade(score float64) Grade {
	switch {
	case score >= 90:
		return GradeA
	case score >= 75:
		return GradeB
	case score >= 60:
		return GradeC
	default:
		return GradeD
	}
}

// WeightedOverall computes the overall score using PRD weights.
func (d DimensionScores) WeightedOverall() float64 {
	return d.Completeness*0.20 +
		d.TechExpress*0.25 +
		d.ProjectQual*0.25 +
		d.JDMatch*0.20 +
		d.Readability*0.10
}
