package resume_test

import (
	"context"
	"strings"
	"testing"

	"github.com/cuixiangyu/ai-interview/internal/ai"
	"github.com/cuixiangyu/ai-interview/internal/domain"
	"github.com/cuixiangyu/ai-interview/internal/repository/memory"
	"github.com/cuixiangyu/ai-interview/internal/service/analyzer"
	"github.com/cuixiangyu/ai-interview/internal/service/exporter"
	"github.com/cuixiangyu/ai-interview/internal/service/generator"
	"github.com/cuixiangyu/ai-interview/internal/service/parser"
	"github.com/cuixiangyu/ai-interview/internal/service/resume"
	"github.com/cuixiangyu/ai-interview/internal/service/suggestion"
)

func newTestService() *resume.Service {
	store := memory.NewStore()
	p := parser.New(parser.NewRuleStructureBuilder(), 10*1024*1024, parser.DefaultExtractors()...)
	return resume.New(resume.Deps{
		Resumes:     store,
		Analyses:    memory.AnalysisAdapter{Store: store},
		Suggestions: memory.SuggestionAdapter{Store: store},
		Versions:    memory.VersionAdapter{Store: store},
		Parser:      p,
		Analyzer:    analyzer.NewRuleAnalyzer(),
		Suggester:   suggestion.NewRuleSuggester(),
		Generator:   generator.Default(),
		Exporter:    exporter.Default(),
		AI:          ai.RuleProvider{},
	})
}

const resumeText = `王五
wangwu@example.com
13900139000

专业技能
Go Redis MySQL Docker

工作经历
某公司
后端工程师
2022-01 - 2024-01
- 负责后端开发与维护

项目经历
订单系统
使用 Go 与 Redis 构建高并发订单服务
- 支撑日活用户增长

教育背景
某某大学
本科
2018-09 - 2022-06
`

func TestUploadAnalyzeGenerateExport(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	result, err := svc.UploadAndAnalyze(ctx, resume.UploadRequest{
		Text:       resumeText,
		TargetRole: "Go 后端",
		TargetJD:   "Go Redis Docker 高并发",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Resume.ID == "" || result.Analysis.ID == "" {
		t.Fatal("missing ids")
	}
	if len(result.Suggestions) == 0 {
		t.Fatal("expected suggestions")
	}

	// accept first suggestion
	_, err = svc.UpdateSuggestionStatus(ctx, result.Suggestions[0].ID, domain.StatusAccepted, "")
	if err != nil {
		t.Fatal(err)
	}

	ver, err := svc.Generate(ctx, resume.GenerateRequest{
		ResumeID: result.Resume.ID,
		Template: domain.TemplateTech,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ver.RenderedMD, "王五") {
		t.Fatalf("rendered missing name: %s", ver.RenderedMD)
	}

	filename, ct, data, err := svc.Export(ctx, ver.ID, domain.ExportMarkdown)
	if err != nil {
		t.Fatal(err)
	}
	if filename == "" || !strings.Contains(ct, "markdown") || len(data) == 0 {
		t.Fatalf("export failed: %s %s %d", filename, ct, len(data))
	}

	orig, opt, err := svc.Compare(ctx, result.Resume.ID, ver.ID)
	if err != nil || orig == "" || opt == "" {
		t.Fatalf("compare failed: %v", err)
	}
}

func TestUpdateStructuredAndDelete(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	result, err := svc.UploadAndAnalyze(ctx, resume.UploadRequest{Text: resumeText})
	if err != nil {
		t.Fatal(err)
	}
	sr := result.Resume.Structured
	sr.BasicInfo.Location = "上海"
	updated, err := svc.UpdateStructured(ctx, result.Resume.ID, sr)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Structured.BasicInfo.Location != "上海" {
		t.Fatal("location not updated")
	}
	if err := svc.DeleteResume(ctx, result.Resume.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetResume(ctx, result.Resume.ID); err != domain.ErrNotFound {
		t.Fatalf("want not found, got %v", err)
	}
}
