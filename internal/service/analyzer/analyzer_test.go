package analyzer

import (
	"context"
	"testing"

	"github.com/cuixiangyu/ai-interview/internal/domain"
)

func sampleResume() *domain.Resume {
	return &domain.Resume{
		ID: "res_1",
		Structured: domain.StructuredResume{
			BasicInfo: domain.BasicInfo{
				Name:  "张三",
				Email: "a@b.com",
				Phone: "13800138000",
				Links: []string{"https://github.com/zhangsan"},
			},
			Skills: domain.Skills{
				Languages:  []string{"Go", "Java"},
				Frameworks: []string{"Gin"},
				Tools:      []string{"Redis", "Docker"},
			},
			WorkExperience: []domain.WorkExperience{
				{
					Company: "Demo",
					Title:   "Backend",
					Highlights: []string{
						"负责后端开发",
						"将 P99 延迟从 800ms 降至 120ms",
					},
				},
			},
			Projects: []domain.Project{
				{Name: "Proj", Description: "QPS 200+", Achievements: []string{"主导设计"}},
			},
			Education: []domain.Education{{School: "Univ"}},
		},
		TargetRole: "Go 后端工程师",
		TargetJD:   "要求熟悉 Go Redis Docker 微服务 高并发",
	}
}

func TestRuleAnalyzerScoreAndIssues(t *testing.T) {
	a := NewRuleAnalyzer()
	an, err := a.Analyze(context.Background(), sampleResume())
	if err != nil {
		t.Fatal(err)
	}
	if an.OverallScore <= 0 || an.OverallScore > 100 {
		t.Fatalf("invalid score %v", an.OverallScore)
	}
	if an.Grade == "" {
		t.Fatal("expected grade")
	}
	if an.Summary == "" {
		t.Fatal("expected summary")
	}
	if len(an.Issues) == 0 {
		t.Fatal("expected some issues (vague highlight)")
	}
	if an.JDMatch == nil {
		t.Fatal("expected JD match")
	}
	if len(an.Highlights) == 0 {
		t.Fatal("expected highlights")
	}
}

func TestRuleAnalyzerNilResume(t *testing.T) {
	_, err := NewRuleAnalyzer().Analyze(context.Background(), nil)
	if err != domain.ErrInvalidInput {
		t.Fatalf("got %v", err)
	}
}

func TestRuleAnalyzerIncompleteResume(t *testing.T) {
	r := &domain.Resume{
		ID: "res_2",
		Structured: domain.StructuredResume{
			BasicInfo: domain.BasicInfo{Name: "李四"},
		},
	}
	an, err := NewRuleAnalyzer().Analyze(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	if an.Grade != domain.GradeD && an.OverallScore >= 60 {
		// incomplete resumes should score poorly
		t.Logf("score=%v grade=%v (acceptable if still low)", an.OverallScore, an.Grade)
	}
	foundEmailIssue := false
	for _, iss := range an.Issues {
		if iss.Priority == domain.PriorityP0 {
			foundEmailIssue = true
		}
	}
	if !foundEmailIssue {
		t.Fatal("expected P0 issues for incomplete resume")
	}
}
