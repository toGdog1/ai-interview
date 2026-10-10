package analyzer

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cuixiangyu/ai-interview/internal/domain"
	"github.com/cuixiangyu/ai-interview/pkg/idgen"
)

// Analyzer scores a structured resume and produces an analysis report.
type Analyzer interface {
	Analyze(ctx context.Context, resume *domain.Resume) (*domain.Analysis, error)
}

// RuleAnalyzer implements PRD scoring with deterministic heuristics.
// Swap with an LLM-backed Analyzer via the same interface for richer results.
type RuleAnalyzer struct{}

// NewRuleAnalyzer creates a rule-based analyzer.
func NewRuleAnalyzer() *RuleAnalyzer {
	return &RuleAnalyzer{}
}

var quantRe = regexp.MustCompile(`(?i)(\d+%|\d+\s*(ms|qps|万|亿|人|倍|x)|p99|p95|latency|throughput)`)

// Analyze evaluates the resume against PRD dimensions.
func (a *RuleAnalyzer) Analyze(_ context.Context, resume *domain.Resume) (*domain.Analysis, error) {
	if resume == nil {
		return nil, domain.ErrInvalidInput
	}
	sr := resume.Structured
	dims := domain.DimensionScores{
		Completeness: scoreCompleteness(sr),
		TechExpress:  scoreTechExpress(sr),
		ProjectQual:  scoreProjectQuality(sr),
		JDMatch:      scoreJDMatch(sr, resume.TargetJD, resume.TargetRole),
		Readability:  scoreReadability(resume.RawContent, sr),
	}
	overall := dims.WeightedOverall()
	issues := collectIssues(sr, resume)
	highlights := collectHighlights(sr)
	summary := buildSummary(overall, issues)

	var jdReport *domain.JDMatchReport
	if strings.TrimSpace(resume.TargetJD) != "" || strings.TrimSpace(resume.TargetRole) != "" {
		jdReport = buildJDMatch(sr, resume.TargetJD, resume.TargetRole)
	}

	return &domain.Analysis{
		ID:              idgen.New("an"),
		ResumeID:        resume.ID,
		OverallScore:    round1(overall),
		Grade:           domain.ScoreToGrade(overall),
		Summary:         summary,
		DimensionScores: dims,
		Issues:          issues,
		Highlights:      highlights,
		JDMatch:         jdReport,
		CreatedAt:       time.Now().UTC(),
	}, nil
}

func scoreCompleteness(sr domain.StructuredResume) float64 {
	score := 40.0
	if sr.BasicInfo.Name != "" {
		score += 10
	}
	if sr.BasicInfo.Email != "" {
		score += 10
	}
	if sr.BasicInfo.Phone != "" {
		score += 5
	}
	if len(sr.Skills.Languages)+len(sr.Skills.Frameworks)+len(sr.Skills.Tools) > 0 {
		score += 15
	}
	if len(sr.WorkExperience) > 0 || len(sr.Projects) > 0 {
		score += 15
	}
	if len(sr.Education) > 0 {
		score += 5
	}
	return clamp(score)
}

func scoreTechExpress(sr domain.StructuredResume) float64 {
	skillCount := len(sr.Skills.Languages) + len(sr.Skills.Frameworks) + len(sr.Skills.Tools)
	score := 50.0
	if skillCount == 0 {
		return 30
	}
	if skillCount >= 3 {
		score += 15
	}
	if skillCount >= 8 {
		score += 10
	}
	if len(sr.Skills.Languages) > 0 && len(sr.Skills.Frameworks) > 0 {
		score += 10
	}
	// Penalize buzzword-only phrasing in highlights
	vague := 0
	for _, w := range sr.WorkExperience {
		for _, h := range w.Highlights {
			if isVague(h) {
				vague++
			}
		}
	}
	score -= float64(vague) * 5
	return clamp(score)
}

func scoreProjectQuality(sr domain.StructuredResume) float64 {
	items := 0
	quant := 0
	personal := 0
	for _, p := range sr.Projects {
		items++
		text := p.Description + " " + strings.Join(p.Achievements, " ")
		if quantRe.MatchString(text) {
			quant++
		}
		if strings.Contains(text, "负责") || strings.Contains(text, "主导") ||
			strings.Contains(strings.ToLower(text), "led") || strings.Contains(strings.ToLower(text), "owned") {
			personal++
		}
	}
	for _, w := range sr.WorkExperience {
		for _, h := range w.Highlights {
			items++
			if quantRe.MatchString(h) {
				quant++
			}
			if !isVague(h) {
				personal++
			}
		}
	}
	if items == 0 {
		return 35
	}
	score := 45.0 + float64(quant)*12 + float64(personal)*5
	if len(sr.Projects) > 0 {
		score += 8
	}
	return clamp(score)
}

func scoreJDMatch(sr domain.StructuredResume, jd, role string) float64 {
	if strings.TrimSpace(jd) == "" && strings.TrimSpace(role) == "" {
		return 75 // neutral when no JD provided
	}
	report := buildJDMatch(sr, jd, role)
	return report.MatchScore
}

func scoreReadability(raw string, sr domain.StructuredResume) float64 {
	score := 70.0
	lines := strings.Count(raw, "\n") + 1
	if lines > 120 {
		score -= 15
	} else if lines < 80 {
		score += 10
	}
	if sr.BasicInfo.Email != "" && sr.BasicInfo.Name != "" {
		score += 10
	}
	if len(sr.WorkExperience) > 0 || len(sr.Projects) > 0 {
		score += 5
	}
	return clamp(score)
}

func collectIssues(sr domain.StructuredResume, resume *domain.Resume) []domain.Issue {
	var issues []domain.Issue
	add := func(p domain.Priority, module, desc, impact, loc string) {
		issues = append(issues, domain.Issue{
			ID:          idgen.New("iss"),
			Priority:    p,
			Module:      module,
			Description: desc,
			Impact:      impact,
			Location:    loc,
		})
	}

	if sr.BasicInfo.Email == "" {
		add(domain.PriorityP0, "basic", "缺少邮箱联系方式", "HR 无法快速联系你，ATS 也可能标记为不完整", "基本信息")
	}
	if len(sr.Skills.Languages)+len(sr.Skills.Frameworks) == 0 {
		add(domain.PriorityP0, "skills", "未识别到清晰的技术栈分类", "面试官难以快速判断技术匹配度", "技能")
	}
	if len(sr.Projects) == 0 && len(sr.WorkExperience) == 0 {
		add(domain.PriorityP0, "projects", "缺少项目或工作经历", "无法证明工程能力与业务贡献", "项目/工作经历")
	}

	hasQuant := false
	for _, p := range sr.Projects {
		text := p.Description + strings.Join(p.Achievements, "")
		if quantRe.MatchString(text) {
			hasQuant = true
		}
	}
	for _, w := range sr.WorkExperience {
		for _, h := range w.Highlights {
			if quantRe.MatchString(h) {
				hasQuant = true
			}
			if isVague(h) {
				add(domain.PriorityP1, "work", fmt.Sprintf("经历描述偏笼统：「%s」", truncate(h, 40)),
					"缺少行动与结果，难以体现个人贡献", "工作经历")
			}
		}
	}
	if !hasQuant && (len(sr.Projects) > 0 || len(sr.WorkExperience) > 0) {
		add(domain.PriorityP0, "projects", "项目/工作描述缺少量化指标",
			"量化结果（QPS、延迟、用户量等）是技术简历差异化的关键", "项目经历")
	}

	hasGitHub := false
	for _, l := range sr.BasicInfo.Links {
		if strings.Contains(strings.ToLower(l), "github") {
			hasGitHub = true
			break
		}
	}
	if !hasGitHub {
		add(domain.PriorityP2, "basic", "未提供 GitHub / 技术博客链接",
			"开源与技术输出能增强可信度", "基本信息")
	}

	if resume.TargetJD != "" {
		report := buildJDMatch(sr, resume.TargetJD, resume.TargetRole)
		if len(report.MissingKeywords) > 0 {
			add(domain.PriorityP1, "jd",
				fmt.Sprintf("与目标 JD 仍有关键词缺口：%s", strings.Join(limit(report.MissingKeywords, 5), "、")),
				"关键词缺失会降低 ATS 与人工初筛通过率", "岗位匹配")
		}
	}
	return issues
}

func collectHighlights(sr domain.StructuredResume) []string {
	var hs []string
	skillN := len(sr.Skills.Languages) + len(sr.Skills.Frameworks) + len(sr.Skills.Tools)
	if skillN >= 5 {
		hs = append(hs, "技术栈覆盖较完整，便于快速匹配岗位")
	}
	for _, p := range sr.Projects {
		text := p.Description + strings.Join(p.Achievements, "")
		if quantRe.MatchString(text) {
			hs = append(hs, fmt.Sprintf("项目「%s」包含量化成果，说服力较强", p.Name))
			break
		}
	}
	for _, l := range sr.BasicInfo.Links {
		if strings.Contains(strings.ToLower(l), "github") {
			hs = append(hs, "已附带 GitHub 链接，利于展示工程实力")
			break
		}
	}
	if len(hs) == 0 {
		hs = append(hs, "简历结构可识别，具备进一步优化的基础")
	}
	return hs
}

func buildSummary(score float64, issues []domain.Issue) string {
	grade := domain.ScoreToGrade(score)
	p0 := 0
	for _, i := range issues {
		if i.Priority == domain.PriorityP0 {
			p0++
		}
	}
	switch grade {
	case domain.GradeA:
		return fmt.Sprintf("综合得分 %.0f（A），简历整体质量优秀，稍作微调即可投递。", score)
	case domain.GradeB:
		return fmt.Sprintf("综合得分 %.0f（B），整体良好；优先处理 %d 个高优先级问题可显著提升通过率。", score, p0)
	case domain.GradeC:
		return fmt.Sprintf("综合得分 %.0f（C），建议重点优化项目描述与技能表达，补齐量化成果。", score)
	default:
		return fmt.Sprintf("综合得分 %.0f（D），简历需要较大幅度修改，建议从联系方式完整性与项目量化写起。", score)
	}
}

func buildJDMatch(sr domain.StructuredResume, jd, role string) *domain.JDMatchReport {
	corpus := strings.ToLower(jd + " " + role)
	keywords := extractJDKeywords(corpus)
	resumeText := strings.ToLower(flattenResume(sr))

	var covered, missing []string
	for _, kw := range keywords {
		if strings.Contains(resumeText, strings.ToLower(kw)) {
			covered = append(covered, kw)
		} else {
			missing = append(missing, kw)
		}
	}
	score := 75.0
	if len(keywords) > 0 {
		score = float64(len(covered)) / float64(len(keywords)) * 100
	}
	var suggestions []string
	for _, m := range limit(missing, 5) {
		suggestions = append(suggestions, fmt.Sprintf("如有相关经验，建议在技能或项目中体现「%s」", m))
	}
	return &domain.JDMatchReport{
		CoveredKeywords: covered,
		MissingKeywords: missing,
		Suggestions:     suggestions,
		MatchScore:      round1(score),
	}
}

func extractJDKeywords(corpus string) []string {
	candidates := []string{
		"go", "golang", "java", "python", "redis", "mysql", "kafka", "docker",
		"kubernetes", "微服务", "分布式", "高并发", "grpc", "postgresql", "mongodb",
		"aws", "ci/cd", "prometheus", "elasticsearch", "react", "vue", "typescript",
	}
	var found []string
	for _, c := range candidates {
		if strings.Contains(corpus, c) {
			found = append(found, displayKeyword(c))
		}
	}
	return unique(found)
}

func displayKeyword(k string) string {
	switch k {
	case "golang":
		return "Go"
	case "ci/cd":
		return "CI/CD"
	default:
		return k
	}
}

func flattenResume(sr domain.StructuredResume) string {
	var b strings.Builder
	b.WriteString(sr.Summary)
	for _, l := range sr.Skills.Languages {
		b.WriteString(" " + l)
	}
	for _, l := range sr.Skills.Frameworks {
		b.WriteString(" " + l)
	}
	for _, l := range sr.Skills.Tools {
		b.WriteString(" " + l)
	}
	for _, w := range sr.WorkExperience {
		b.WriteString(" " + w.Company + " " + w.Title + " " + strings.Join(w.Highlights, " "))
	}
	for _, p := range sr.Projects {
		b.WriteString(" " + p.Name + " " + p.Description + " " + strings.Join(p.Achievements, " ") + " " + strings.Join(p.TechStack, " "))
	}
	return b.String()
}

func isVague(s string) bool {
	vaguePhrases := []string{"负责开发", "参与开发", "负责后端", "日常维护", "完成了", "熟悉使用"}
	for _, p := range vaguePhrases {
		if strings.Contains(s, p) && !quantRe.MatchString(s) {
			return true
		}
	}
	return false
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return round1(v)
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func limit(in []string, n int) []string {
	if len(in) <= n {
		return in
	}
	return in[:n]
}

func unique(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		k := strings.ToLower(s)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, s)
	}
	return out
}
