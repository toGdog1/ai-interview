package suggestion

import (
	"context"
	"testing"

	"github.com/cuixiangyu/ai-interview/internal/domain"
)

func TestRuleSuggester(t *testing.T) {
	resume := &domain.Resume{
		ID: "r1",
		Structured: domain.StructuredResume{
			Skills: domain.Skills{Languages: []string{"Go"}},
			WorkExperience: []domain.WorkExperience{
				{Highlights: []string{"负责后端开发"}},
			},
		},
	}
	analysis := &domain.Analysis{
		ID: "a1",
		Issues: []domain.Issue{
			{
				ID:          "i1",
				Priority:    domain.PriorityP0,
				Module:      "projects",
				Description: "项目/工作描述缺少量化指标",
				Impact:      "影响通过率",
			},
			{
				ID:          "i2",
				Priority:    domain.PriorityP0,
				Module:      "basic",
				Description: "缺少邮箱联系方式",
				Impact:      "无法联系",
			},
		},
	}

	sugs, err := NewRuleSuggester().Suggest(context.Background(), resume, analysis)
	if err != nil {
		t.Fatal(err)
	}
	if len(sugs) < 2 {
		t.Fatalf("expected at least 2 suggestions, got %d", len(sugs))
	}
	for _, s := range sugs {
		if s.Advice == "" || s.Status != domain.StatusPending {
			t.Fatalf("invalid suggestion: %+v", s)
		}
	}
}

func TestEstimateScoreGain(t *testing.T) {
	sugs := []domain.Suggestion{
		{Priority: domain.PriorityP0, Status: domain.StatusAccepted},
		{Priority: domain.PriorityP1, Status: domain.StatusAccepted},
		{Priority: domain.PriorityP2, Status: domain.StatusIgnored},
	}
	gain := EstimateScoreGain(sugs)
	if gain != 6.5 {
		t.Fatalf("gain=%v", gain)
	}
}
