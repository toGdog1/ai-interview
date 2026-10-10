package ai

import (
	"context"

	"github.com/cuixiangyu/ai-interview/internal/domain"
)

// Provider is the extension point for LLM-backed resume intelligence.
// The MVP uses RuleProvider (deterministic). Swap in OpenAIProvider later
// without changing service orchestration.
type Provider interface {
	Name() string
	// EnhanceAnalysis optionally enriches a rule-based analysis.
	EnhanceAnalysis(ctx context.Context, resume *domain.Resume, base *domain.Analysis) (*domain.Analysis, error)
	// EnhanceSuggestions optionally rewrites suggestion After examples.
	EnhanceSuggestions(ctx context.Context, resume *domain.Resume, suggestions []domain.Suggestion) ([]domain.Suggestion, error)
}

// RuleProvider is a no-op enhancer that keeps rule-engine results as-is.
type RuleProvider struct{}

func (RuleProvider) Name() string { return "rule" }

func (RuleProvider) EnhanceAnalysis(_ context.Context, _ *domain.Resume, base *domain.Analysis) (*domain.Analysis, error) {
	return base, nil
}

func (RuleProvider) EnhanceSuggestions(_ context.Context, _ *domain.Resume, suggestions []domain.Suggestion) ([]domain.Suggestion, error) {
	return suggestions, nil
}

// NewProvider selects an AI provider by name.
func NewProvider(name string) Provider {
	switch name {
	case "openai":
		// Placeholder: wire OpenAI HTTP client here when API key is available.
		return RuleProvider{}
	default:
		return RuleProvider{}
	}
}
