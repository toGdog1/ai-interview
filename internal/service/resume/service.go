package resume

import (
	"context"
	"fmt"
	"time"

	"github.com/cuixiangyu/ai-interview/internal/ai"
	"github.com/cuixiangyu/ai-interview/internal/domain"
	"github.com/cuixiangyu/ai-interview/internal/repository"
	"github.com/cuixiangyu/ai-interview/internal/service/analyzer"
	"github.com/cuixiangyu/ai-interview/internal/service/exporter"
	"github.com/cuixiangyu/ai-interview/internal/service/generator"
	"github.com/cuixiangyu/ai-interview/internal/service/parser"
	"github.com/cuixiangyu/ai-interview/internal/service/suggestion"
	"github.com/cuixiangyu/ai-interview/pkg/idgen"
)

// UploadRequest is the input for creating a resume.
type UploadRequest struct {
	Content    []byte
	Format     domain.FileFormat
	Text       string
	TargetRole string
	TargetJD   string
	YearsOfExp int
	UserID     string
}

// PipelineResult is returned after upload + analyze + suggest.
type PipelineResult struct {
	Resume      *domain.Resume       `json:"resume"`
	Analysis    *domain.Analysis     `json:"analysis"`
	Suggestions []domain.Suggestion  `json:"suggestions"`
}

// GenerateRequest controls optimized resume generation.
type GenerateRequest struct {
	ResumeID      string
	Template      domain.TemplateID
	SuggestionIDs []string // empty = all accepted
}

// Service orchestrates the resume optimization pipeline.
type Service struct {
	resumes     repository.ResumeRepository
	analyses    repository.AnalysisRepository
	suggestions repository.SuggestionRepository
	versions    repository.VersionRepository
	parser      *parser.Parser
	analyzer    analyzer.Analyzer
	suggester   suggestion.Suggester
	generator   generator.Generator
	exporter      *exporter.Registry
	ai          ai.Provider
}

// Deps wires collaborators for Service construction.
type Deps struct {
	Resumes     repository.ResumeRepository
	Analyses    repository.AnalysisRepository
	Suggestions repository.SuggestionRepository
	Versions    repository.VersionRepository
	Parser      *parser.Parser
	Analyzer    analyzer.Analyzer
	Suggester   suggestion.Suggester
	Generator   generator.Generator
	Exporter    *exporter.Registry
	AI          ai.Provider
}

// New creates a resume Service.
func New(d Deps) *Service {
	if d.AI == nil {
		d.AI = ai.RuleProvider{}
	}
	return &Service{
		resumes:     d.Resumes,
		analyses:    d.Analyses,
		suggestions: d.Suggestions,
		versions:    d.Versions,
		parser:      d.Parser,
		analyzer:    d.Analyzer,
		suggester:   d.Suggester,
		generator:   d.Generator,
		exporter:      d.Exporter,
		ai:          d.AI,
	}
}

// UploadAndAnalyze parses, stores, analyzes, and generates suggestions.
func (s *Service) UploadAndAnalyze(ctx context.Context, req UploadRequest) (*PipelineResult, error) {
	raw, structured, err := s.parser.Parse(ctx, parser.ParseInput{
		Content: req.Content,
		Format:  req.Format,
		Text:    req.Text,
	})
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	format := req.Format
	if format == "" && req.Text != "" {
		format = domain.FormatTXT
	}
	resume := &domain.Resume{
		ID:         idgen.New("res"),
		UserID:     req.UserID,
		RawContent: raw,
		Format:     format,
		Structured: structured,
		TargetRole: req.TargetRole,
		TargetJD:   req.TargetJD,
		YearsOfExp: req.YearsOfExp,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.resumes.Save(ctx, resume); err != nil {
		return nil, err
	}

	analysis, err := s.analyzer.Analyze(ctx, resume)
	if err != nil {
		return nil, err
	}
	analysis, err = s.ai.EnhanceAnalysis(ctx, resume, analysis)
	if err != nil {
		return nil, err
	}
	if err := s.analyses.Save(ctx, analysis); err != nil {
		return nil, err
	}

	suggestions, err := s.suggester.Suggest(ctx, resume, analysis)
	if err != nil {
		return nil, err
	}
	suggestions, err = s.ai.EnhanceSuggestions(ctx, resume, suggestions)
	if err != nil {
		return nil, err
	}
	if err := s.suggestions.SaveBatch(ctx, suggestions); err != nil {
		return nil, err
	}

	return &PipelineResult{
		Resume:      resume,
		Analysis:    analysis,
		Suggestions: suggestions,
	}, nil
}

// GetResume returns a stored resume.
func (s *Service) GetResume(ctx context.Context, id string) (*domain.Resume, error) {
	return s.resumes.FindByID(ctx, id)
}

// UpdateStructured allows users to correct parsed fields before re-analysis.
func (s *Service) UpdateStructured(ctx context.Context, id string, structured domain.StructuredResume) (*domain.Resume, error) {
	resume, err := s.resumes.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resume.Structured = structured
	resume.UpdatedAt = time.Now().UTC()
	if err := s.resumes.Update(ctx, resume); err != nil {
		return nil, err
	}
	return resume, nil
}

// GetAnalysis returns analysis by id.
func (s *Service) GetAnalysis(ctx context.Context, id string) (*domain.Analysis, error) {
	return s.analyses.FindByID(ctx, id)
}

// GetAnalysisByResume returns analysis for a resume.
func (s *Service) GetAnalysisByResume(ctx context.Context, resumeID string) (*domain.Analysis, error) {
	return s.analyses.FindByResumeID(ctx, resumeID)
}

// ListSuggestions returns suggestions for an analysis.
func (s *Service) ListSuggestions(ctx context.Context, analysisID string) ([]domain.Suggestion, error) {
	return s.suggestions.FindByAnalysisID(ctx, analysisID)
}

// UpdateSuggestionStatus accepts, ignores, or edits a suggestion.
func (s *Service) UpdateSuggestionStatus(ctx context.Context, id string, status domain.SuggestionStatus, after string) (*domain.Suggestion, error) {
	sug, err := s.suggestions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	switch status {
	case domain.StatusAccepted, domain.StatusIgnored, domain.StatusEdited, domain.StatusPending:
	default:
		return nil, domain.ErrInvalidStatus
	}
	sug.Status = status
	if status == domain.StatusEdited && after != "" {
		sug.After = after
	}
	if err := s.suggestions.Update(ctx, sug); err != nil {
		return nil, err
	}
	return sug, nil
}

// Generate builds an optimized resume version from accepted suggestions.
func (s *Service) Generate(ctx context.Context, req GenerateRequest) (*domain.ResumeVersion, error) {
	resume, err := s.resumes.FindByID(ctx, req.ResumeID)
	if err != nil {
		return nil, err
	}
	analysis, err := s.analyses.FindByResumeID(ctx, resume.ID)
	if err != nil {
		return nil, err
	}
	all, err := s.suggestions.FindByAnalysisID(ctx, analysis.ID)
	if err != nil {
		return nil, err
	}

	selected := filterSuggestions(all, req.SuggestionIDs)
	// If caller didn't accept any yet but passed IDs, mark them accepted in-memory for generation.
	if len(req.SuggestionIDs) > 0 {
		for i := range selected {
			if selected[i].Status == domain.StatusPending {
				selected[i].Status = domain.StatusAccepted
			}
		}
	} else {
		// Default: treat all as accepted for one-click generate (MVP UX).
		for i := range selected {
			if selected[i].Status == domain.StatusPending {
				selected[i].Status = domain.StatusAccepted
			}
		}
	}

	version, err := s.generator.Generate(ctx, resume, selected, req.Template)
	if err != nil {
		return nil, err
	}
	if err := s.versions.Save(ctx, version); err != nil {
		return nil, err
	}
	return version, nil
}

// Export downloads a version in the given format.
func (s *Service) Export(ctx context.Context, versionID string, format domain.ExportFormat) (filename, contentType string, data []byte, err error) {
	version, err := s.versions.FindByID(ctx, versionID)
	if err != nil {
		return "", "", nil, err
	}
	return s.exporter.Export(ctx, format, version)
}

// GetVersion returns a generated version.
func (s *Service) GetVersion(ctx context.Context, id string) (*domain.ResumeVersion, error) {
	return s.versions.FindByID(ctx, id)
}

// DeleteResume removes a resume (privacy).
func (s *Service) DeleteResume(ctx context.Context, id string) error {
	return s.resumes.Delete(ctx, id)
}

// SuggestionSummary returns acceptance stats.
func (s *Service) SuggestionSummary(ctx context.Context, analysisID string) (accepted int, estimatedGain float64, err error) {
	list, err := s.suggestions.FindByAnalysisID(ctx, analysisID)
	if err != nil {
		return 0, 0, err
	}
	for _, sug := range list {
		if sug.Status == domain.StatusAccepted || sug.Status == domain.StatusEdited {
			accepted++
		}
	}
	return accepted, suggestion.EstimateScoreGain(list), nil
}

func filterSuggestions(all []domain.Suggestion, ids []string) []domain.Suggestion {
	if len(ids) == 0 {
		return all
	}
	want := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		want[id] = struct{}{}
	}
	var out []domain.Suggestion
	for _, s := range all {
		if _, ok := want[s.ID]; ok {
			out = append(out, s)
		}
	}
	return out
}

// Compare returns a simple original vs optimized markdown pair.
func (s *Service) Compare(ctx context.Context, resumeID, versionID string) (originalMD, optimizedMD string, err error) {
	resume, err := s.resumes.FindByID(ctx, resumeID)
	if err != nil {
		return "", "", err
	}
	version, err := s.versions.FindByID(ctx, versionID)
	if err != nil {
		return "", "", err
	}
	originalMD = resume.RawContent
	if originalMD == "" {
		originalMD = fmt.Sprintf("%s\n%s", resume.Structured.BasicInfo.Name, resume.Structured.Summary)
	}
	return originalMD, version.RenderedMD, nil
}
