package generator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cuixiangyu/ai-interview/internal/domain"
	"github.com/cuixiangyu/ai-interview/pkg/idgen"
)

// Generator creates an optimized resume version from accepted suggestions.
type Generator interface {
	Generate(ctx context.Context, resume *domain.Resume, suggestions []domain.Suggestion, template domain.TemplateID) (*domain.ResumeVersion, error)
}

// TemplateRenderer renders a StructuredResume into markdown for a template.
type TemplateRenderer interface {
	ID() domain.TemplateID
	Render(sr domain.StructuredResume) string
}

// Service generates optimized resume versions.
type Service struct {
	renderers map[domain.TemplateID]TemplateRenderer
}

// New creates a Generator with the given template renderers.
func New(renderers ...TemplateRenderer) *Service {
	m := make(map[domain.TemplateID]TemplateRenderer, len(renderers))
	for _, r := range renderers {
		m[r.ID()] = r
	}
	return &Service{renderers: m}
}

// Default creates a generator with ATS and Tech templates.
func Default() *Service {
	return New(&ATSRenderer{}, &TechRenderer{})
}

// Generate applies accepted suggestions and renders the chosen template.
func (s *Service) Generate(_ context.Context, resume *domain.Resume, suggestions []domain.Suggestion, template domain.TemplateID) (*domain.ResumeVersion, error) {
	if resume == nil {
		return nil, domain.ErrInvalidInput
	}
	if template == "" {
		template = domain.TemplateATS
	}
	renderer, ok := s.renderers[template]
	if !ok {
		return nil, fmt.Errorf("%w: template %s", domain.ErrInvalidInput, template)
	}

	optimized := applySuggestions(resume.Structured, suggestions)
	var acceptedIDs []string
	for _, sug := range suggestions {
		if sug.Status == domain.StatusAccepted || sug.Status == domain.StatusEdited {
			acceptedIDs = append(acceptedIDs, sug.ID)
		}
	}

	return &domain.ResumeVersion{
		ID:            idgen.New("ver"),
		ResumeID:      resume.ID,
		Template:      template,
		Content:       optimized,
		RenderedMD:    renderer.Render(optimized),
		SuggestionIDs: acceptedIDs,
		CreatedAt:     time.Now().UTC(),
	}, nil
}

func applySuggestions(sr domain.StructuredResume, suggestions []domain.Suggestion) domain.StructuredResume {
	out := cloneResume(sr)
	for _, sug := range suggestions {
		if sug.Status != domain.StatusAccepted && sug.Status != domain.StatusEdited {
			continue
		}
		switch sug.Type {
		case domain.TypeSupplement:
			if sug.Module == "basic" && strings.Contains(sug.After, "GitHub") {
				out.BasicInfo.Links = appendUnique(out.BasicInfo.Links, extractURL(sug.After))
			}
			if sug.Module == "basic" && (strings.Contains(sug.After, "邮箱") || strings.Contains(sug.After, "@")) {
				if email := extractEmail(sug.After); email != "" && out.BasicInfo.Email == "" {
					out.BasicInfo.Email = email
				}
			}
		case domain.TypeRewrite:
			rewriteFirstHighlight(&out, sug.Before, sug.After)
		case domain.TypeSkill:
			// Keep structured skills; rendering will use grouped format.
		}
	}
	// Never invent fake companies/projects — only rewrite existing text or fill missing contact.
	return out
}

func rewriteFirstHighlight(sr *domain.StructuredResume, before, after string) {
	if after == "" {
		return
	}
	for i := range sr.WorkExperience {
		for j, h := range sr.WorkExperience[i].Highlights {
			if before == "" || strings.Contains(h, before) || h == before {
				sr.WorkExperience[i].Highlights[j] = after
				return
			}
		}
	}
	for i := range sr.Projects {
		if before == "" || strings.Contains(sr.Projects[i].Description, before) {
			if len(sr.Projects[i].Achievements) > 0 {
				sr.Projects[i].Achievements[0] = after
			} else {
				sr.Projects[i].Achievements = []string{after}
			}
			return
		}
	}
	// If nothing matched, append as achievement on first project/work — only when after looks like guidance not placeholder
	if len(sr.WorkExperience) > 0 && !strings.Contains(after, "项目名称") {
		sr.WorkExperience[0].Highlights = append(sr.WorkExperience[0].Highlights, after)
	}
}

func cloneResume(sr domain.StructuredResume) domain.StructuredResume {
	out := sr
	out.BasicInfo.Links = append([]string{}, sr.BasicInfo.Links...)
	out.Skills.Languages = append([]string{}, sr.Skills.Languages...)
	out.Skills.Frameworks = append([]string{}, sr.Skills.Frameworks...)
	out.Skills.Tools = append([]string{}, sr.Skills.Tools...)
	out.Skills.Other = append([]string{}, sr.Skills.Other...)
	out.WorkExperience = append([]domain.WorkExperience{}, sr.WorkExperience...)
	for i := range out.WorkExperience {
		out.WorkExperience[i].Highlights = append([]string{}, sr.WorkExperience[i].Highlights...)
	}
	out.Projects = append([]domain.Project{}, sr.Projects...)
	for i := range out.Projects {
		out.Projects[i].TechStack = append([]string{}, sr.Projects[i].TechStack...)
		out.Projects[i].Achievements = append([]string{}, sr.Projects[i].Achievements...)
	}
	out.Education = append([]domain.Education{}, sr.Education...)
	out.Certifications = append([]string{}, sr.Certifications...)
	out.Others = append([]string{}, sr.Others...)
	return out
}

func appendUnique(list []string, v string) []string {
	if v == "" {
		return list
	}
	for _, x := range list {
		if strings.EqualFold(x, v) {
			return list
		}
	}
	return append(list, v)
}

func extractURL(s string) string {
	for _, part := range strings.Fields(s) {
		if strings.Contains(part, "http") || strings.Contains(part, "github.com") {
			return strings.TrimRight(part, "。.;,")
		}
	}
	return ""
}

func extractEmail(s string) string {
	for _, part := range strings.Fields(s) {
		if strings.Contains(part, "@") {
			return strings.TrimRight(part, "。.;,")
		}
	}
	return ""
}
