package memory

import (
	"context"
	"sync"

	"github.com/cuixiangyu/ai-interview/internal/domain"
)

// Store is a thread-safe in-memory implementation of all repositories.
// Suitable for MVP and tests; swap for SQL/Redis in production.
type Store struct {
	mu          sync.RWMutex
	resumes     map[string]*domain.Resume
	analyses    map[string]*domain.Analysis
	suggestions map[string]*domain.Suggestion
	byAnalysis  map[string][]string // analysisID -> suggestionIDs
	versions    map[string]*domain.ResumeVersion
	byResumeVer map[string][]string // resumeID -> versionIDs
}

// NewStore creates an empty in-memory store.
func NewStore() *Store {
	return &Store{
		resumes:     make(map[string]*domain.Resume),
		analyses:    make(map[string]*domain.Analysis),
		suggestions: make(map[string]*domain.Suggestion),
		byAnalysis:  make(map[string][]string),
		versions:    make(map[string]*domain.ResumeVersion),
		byResumeVer: make(map[string][]string),
	}
}

// --- ResumeRepository ---

func (s *Store) Save(_ context.Context, resume *domain.Resume) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *resume
	s.resumes[resume.ID] = &cp
	return nil
}

func (s *Store) FindByID(_ context.Context, id string) (*domain.Resume, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.resumes[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *r
	return &cp, nil
}

func (s *Store) Update(ctx context.Context, resume *domain.Resume) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.resumes[resume.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *resume
	s.resumes[resume.ID] = &cp
	return nil
}

func (s *Store) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.resumes[id]; !ok {
		return domain.ErrNotFound
	}
	delete(s.resumes, id)
	return nil
}

// --- AnalysisRepository ---

func (s *Store) SaveAnalysis(_ context.Context, analysis *domain.Analysis) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *analysis
	s.analyses[analysis.ID] = &cp
	return nil
}

func (s *Store) FindAnalysisByID(_ context.Context, id string) (*domain.Analysis, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.analyses[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *a
	return &cp, nil
}

func (s *Store) FindAnalysisByResumeID(_ context.Context, resumeID string) (*domain.Analysis, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.analyses {
		if a.ResumeID == resumeID {
			cp := *a
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

// AnalysisAdapter wraps Store to satisfy AnalysisRepository with shorter method names.
type AnalysisAdapter struct{ *Store }

func (a AnalysisAdapter) Save(ctx context.Context, analysis *domain.Analysis) error {
	return a.SaveAnalysis(ctx, analysis)
}

func (a AnalysisAdapter) FindByID(ctx context.Context, id string) (*domain.Analysis, error) {
	return a.FindAnalysisByID(ctx, id)
}

func (a AnalysisAdapter) FindByResumeID(ctx context.Context, resumeID string) (*domain.Analysis, error) {
	return a.FindAnalysisByResumeID(ctx, resumeID)
}

// --- SuggestionRepository ---

func (s *Store) SaveSuggestions(_ context.Context, suggestions []domain.Suggestion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range suggestions {
		cp := suggestions[i]
		s.suggestions[cp.ID] = &cp
		s.byAnalysis[cp.AnalysisID] = append(s.byAnalysis[cp.AnalysisID], cp.ID)
	}
	return nil
}

func (s *Store) FindSuggestionsByAnalysisID(_ context.Context, analysisID string) ([]domain.Suggestion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.byAnalysis[analysisID]
	out := make([]domain.Suggestion, 0, len(ids))
	for _, id := range ids {
		if sug, ok := s.suggestions[id]; ok {
			out = append(out, *sug)
		}
	}
	return out, nil
}

func (s *Store) FindSuggestionByID(_ context.Context, id string) (*domain.Suggestion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sug, ok := s.suggestions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *sug
	return &cp, nil
}

func (s *Store) UpdateSuggestion(_ context.Context, suggestion *domain.Suggestion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.suggestions[suggestion.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *suggestion
	s.suggestions[suggestion.ID] = &cp
	return nil
}

// SuggestionAdapter wraps Store for SuggestionRepository.
type SuggestionAdapter struct{ *Store }

func (a SuggestionAdapter) SaveBatch(ctx context.Context, suggestions []domain.Suggestion) error {
	return a.SaveSuggestions(ctx, suggestions)
}

func (a SuggestionAdapter) FindByAnalysisID(ctx context.Context, analysisID string) ([]domain.Suggestion, error) {
	return a.FindSuggestionsByAnalysisID(ctx, analysisID)
}

func (a SuggestionAdapter) FindByID(ctx context.Context, id string) (*domain.Suggestion, error) {
	return a.FindSuggestionByID(ctx, id)
}

func (a SuggestionAdapter) Update(ctx context.Context, suggestion *domain.Suggestion) error {
	return a.UpdateSuggestion(ctx, suggestion)
}

// --- VersionRepository ---

func (s *Store) SaveVersion(_ context.Context, version *domain.ResumeVersion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *version
	s.versions[version.ID] = &cp
	s.byResumeVer[version.ResumeID] = append(s.byResumeVer[version.ResumeID], version.ID)
	return nil
}

func (s *Store) FindVersionByID(_ context.Context, id string) (*domain.ResumeVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.versions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *v
	return &cp, nil
}

func (s *Store) FindVersionsByResumeID(_ context.Context, resumeID string) ([]domain.ResumeVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.byResumeVer[resumeID]
	out := make([]domain.ResumeVersion, 0, len(ids))
	for _, id := range ids {
		if v, ok := s.versions[id]; ok {
			out = append(out, *v)
		}
	}
	return out, nil
}

// VersionAdapter wraps Store for VersionRepository.
type VersionAdapter struct{ *Store }

func (a VersionAdapter) Save(ctx context.Context, version *domain.ResumeVersion) error {
	return a.SaveVersion(ctx, version)
}

func (a VersionAdapter) FindByID(ctx context.Context, id string) (*domain.ResumeVersion, error) {
	return a.FindVersionByID(ctx, id)
}

func (a VersionAdapter) FindByResumeID(ctx context.Context, resumeID string) ([]domain.ResumeVersion, error) {
	return a.FindVersionsByResumeID(ctx, resumeID)
}
