package app

import (
	"github.com/cuixiangyu/ai-interview/internal/ai"
	"github.com/cuixiangyu/ai-interview/internal/config"
	"github.com/cuixiangyu/ai-interview/internal/handler"
	"github.com/cuixiangyu/ai-interview/internal/repository/memory"
	"github.com/cuixiangyu/ai-interview/internal/service/analyzer"
	"github.com/cuixiangyu/ai-interview/internal/service/exporter"
	"github.com/cuixiangyu/ai-interview/internal/service/generator"
	"github.com/cuixiangyu/ai-interview/internal/service/parser"
	"github.com/cuixiangyu/ai-interview/internal/service/resume"
	"github.com/cuixiangyu/ai-interview/internal/service/suggestion"
	"net/http"
)

// New builds the fully wired HTTP handler from config.
func New(cfg config.Config) http.Handler {
	store := memory.NewStore()

	p := parser.New(
		parser.NewRuleStructureBuilder(),
		cfg.MaxUploadBytes,
		parser.DefaultExtractors()...,
	)

	svc := resume.New(resume.Deps{
		Resumes:     store,
		Analyses:    memory.AnalysisAdapter{Store: store},
		Suggestions: memory.SuggestionAdapter{Store: store},
		Versions:    memory.VersionAdapter{Store: store},
		Parser:      p,
		Analyzer:    analyzer.NewRuleAnalyzer(),
		Suggester:   suggestion.NewRuleSuggester(),
		Generator:   generator.Default(),
		Exporter:    exporter.Default(),
		AI:          ai.NewProvider(cfg.AIProvider),
	})

	h := handler.New(svc, cfg.MaxUploadBytes)
	return handler.NewRouter(h, cfg.EnableCORS)
}
