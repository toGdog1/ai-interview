package generator

import (
	"context"
	"strings"
	"testing"

	"github.com/cuixiangyu/ai-interview/internal/domain"
)

func TestGenerateATSAndTech(t *testing.T) {
	g := Default()
	resume := &domain.Resume{
		ID: "r1",
		Structured: domain.StructuredResume{
			BasicInfo: domain.BasicInfo{Name: "张三", Email: "a@b.com"},
			Skills:    domain.Skills{Languages: []string{"Go"}, Tools: []string{"Redis"}},
			WorkExperience: []domain.WorkExperience{
				{Company: "Demo", Title: "BE", Highlights: []string{"负责后端开发"}},
			},
			Projects: []domain.Project{{Name: "P1", Description: "desc"}},
		},
	}
	sugs := []domain.Suggestion{
		{
			ID:     "s1",
			Type:   domain.TypeRewrite,
			Status: domain.StatusAccepted,
			Before: "负责后端开发",
			After:  "将 P99 延迟从 800ms 降至 120ms",
			Module: "work",
		},
	}

	for _, tpl := range []domain.TemplateID{domain.TemplateATS, domain.TemplateTech} {
		ver, err := g.Generate(context.Background(), resume, sugs, tpl)
		if err != nil {
			t.Fatalf("template %s: %v", tpl, err)
		}
		if ver.RenderedMD == "" {
			t.Fatalf("empty render for %s", tpl)
		}
		if !strings.Contains(ver.RenderedMD, "张三") {
			t.Fatalf("missing name in %s", tpl)
		}
		if len(ver.SuggestionIDs) != 1 {
			t.Fatalf("suggestion ids=%v", ver.SuggestionIDs)
		}
	}
}

func TestGenerateUnknownTemplate(t *testing.T) {
	_, err := Default().Generate(context.Background(), &domain.Resume{ID: "r"}, nil, "unknown")
	if err == nil {
		t.Fatal("expected error")
	}
}
