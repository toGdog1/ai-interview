package domain

import "time"

// SuggestionStatus tracks user action on a suggestion.
type SuggestionStatus string

const (
	StatusPending  SuggestionStatus = "pending"
	StatusAccepted SuggestionStatus = "accepted"
	StatusIgnored  SuggestionStatus = "ignored"
	StatusEdited   SuggestionStatus = "edited"
)

// SuggestionType categorizes the kind of optimization advice.
type SuggestionType string

const (
	TypeRewrite    SuggestionType = "rewrite"
	TypeStructure  SuggestionType = "structure"
	TypeSkill      SuggestionType = "skill"
	TypeDelete     SuggestionType = "delete"
	TypeSupplement SuggestionType = "supplement"
	TypeFormat     SuggestionType = "format"
)

// Suggestion is a concrete, actionable optimization item.
type Suggestion struct {
	ID         string           `json:"id"`
	AnalysisID string           `json:"analysisId"`
	Module     string           `json:"module"`
	Type       SuggestionType   `json:"type"`
	Priority   Priority         `json:"priority"`
	Problem    string           `json:"problem"`
	Reason     string           `json:"reason"`
	Advice     string           `json:"suggestion"`
	Before     string           `json:"before"`
	After      string           `json:"after"`
	Status     SuggestionStatus `json:"status"`
	CreatedAt  time.Time        `json:"createdAt"`
}
