package repository

import (
	"context"

	"github.com/cuixiangyu/ai-interview/internal/domain"
)

// ResumeRepository persists resume aggregates.
type ResumeRepository interface {
	Save(ctx context.Context, resume *domain.Resume) error
	FindByID(ctx context.Context, id string) (*domain.Resume, error)
	Update(ctx context.Context, resume *domain.Resume) error
	Delete(ctx context.Context, id string) error
}

// AnalysisRepository persists analysis reports.
type AnalysisRepository interface {
	Save(ctx context.Context, analysis *domain.Analysis) error
	FindByID(ctx context.Context, id string) (*domain.Analysis, error)
	FindByResumeID(ctx context.Context, resumeID string) (*domain.Analysis, error)
}

// SuggestionRepository persists optimization suggestions.
type SuggestionRepository interface {
	SaveBatch(ctx context.Context, suggestions []domain.Suggestion) error
	FindByAnalysisID(ctx context.Context, analysisID string) ([]domain.Suggestion, error)
	FindByID(ctx context.Context, id string) (*domain.Suggestion, error)
	Update(ctx context.Context, suggestion *domain.Suggestion) error
}

// VersionRepository persists generated resume versions.
type VersionRepository interface {
	Save(ctx context.Context, version *domain.ResumeVersion) error
	FindByID(ctx context.Context, id string) (*domain.ResumeVersion, error)
	FindByResumeID(ctx context.Context, resumeID string) ([]domain.ResumeVersion, error)
}
