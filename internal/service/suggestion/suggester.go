package suggestion

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cuixiangyu/ai-interview/internal/domain"
	"github.com/cuixiangyu/ai-interview/pkg/idgen"
)

// Suggester produces actionable optimization suggestions from an analysis.
type Suggester interface {
	Suggest(ctx context.Context, resume *domain.Resume, analysis *domain.Analysis) ([]domain.Suggestion, error)
}

// RuleSuggester maps analysis issues to concrete Before/After suggestions.
type RuleSuggester struct{}

// NewRuleSuggester creates a rule-based suggester.
func NewRuleSuggester() *RuleSuggester {
	return &RuleSuggester{}
}

// Suggest builds suggestions from analysis issues and resume content.
func (s *RuleSuggester) Suggest(_ context.Context, resume *domain.Resume, analysis *domain.Analysis) ([]domain.Suggestion, error) {
	if resume == nil || analysis == nil {
		return nil, domain.ErrInvalidInput
	}

	var out []domain.Suggestion
	now := time.Now().UTC()

	for _, issue := range analysis.Issues {
		sug := domain.Suggestion{
			ID:         idgen.New("sug"),
			AnalysisID: analysis.ID,
			Module:     issue.Module,
			Priority:   issue.Priority,
			Problem:    issue.Description,
			Reason:     issue.Impact,
			Status:     domain.StatusPending,
			CreatedAt:  now,
		}
		fillAdvice(&sug, issue, resume)
		out = append(out, sug)
	}

	// Always add a structure tip for skills when skills exist but ungrouped messaging.
	if len(resume.Structured.Skills.Languages) > 0 {
		out = append(out, domain.Suggestion{
			ID:         idgen.New("sug"),
			AnalysisID: analysis.ID,
			Module:     "skills",
			Type:       domain.TypeSkill,
			Priority:   domain.PriorityP2,
			Problem:    "技能列表可按熟练度/类别进一步整理",
			Reason:     "分类清晰的技能区能帮助面试官 10 秒内判断匹配度",
			Advice:     "按「语言 / 框架 / 中间件与工具」分组，避免「熟悉 XXX」堆砌",
			Before:     strings.Join(resume.Structured.Skills.Languages, ", "),
			After:      formatSkills(resume.Structured.Skills),
			Status:     domain.StatusPending,
			CreatedAt:  now,
		})
	}

	return out, nil
}

func fillAdvice(sug *domain.Suggestion, issue domain.Issue, resume *domain.Resume) {
	switch {
	case strings.Contains(issue.Description, "邮箱"):
		sug.Type = domain.TypeSupplement
		sug.Advice = "在简历顶部补充常用邮箱，建议使用专业域名或主流邮箱"
		sug.Before = "（缺失）"
		sug.After = "邮箱：your.name@example.com"
	case strings.Contains(issue.Description, "量化"):
		sug.Type = domain.TypeRewrite
		sug.Advice = "为关键项目补充可验证的指标：性能、规模、效率或质量"
		sug.Before = firstVagueHighlight(resume)
		sug.After = "主导订单服务重构，将 P99 延迟从 800ms 降至 120ms，支撑日活 10 万+"
		if sug.Before == "" {
			sug.Before = "负责后端开发与维护"
		}
	case strings.Contains(issue.Description, "笼统"):
		sug.Type = domain.TypeRewrite
		sug.Advice = "使用 STAR 结构重写：背景 → 你的行动 → 可量化结果"
		sug.Before = extractQuoted(issue.Description)
		if sug.Before == "" {
			sug.Before = "负责后端开发"
		}
		sug.After = "负责支付回调链路稳定性建设，引入重试与幂等，将失败率从 2.1% 降至 0.3%"
	case strings.Contains(issue.Description, "GitHub"):
		sug.Type = domain.TypeSupplement
		sug.Advice = "补充 GitHub / 技术博客链接，突出有代表性的仓库"
		sug.Before = "（缺失）"
		sug.After = "GitHub：https://github.com/yourname"
	case strings.Contains(issue.Description, "技术栈"):
		sug.Type = domain.TypeSkill
		sug.Advice = "单独列出语言、框架、中间件，并与目标岗位对齐排序"
		sug.Before = "技能散落在正文中"
		sug.After = formatSkills(resume.Structured.Skills)
	case strings.Contains(issue.Description, "关键词"):
		sug.Type = domain.TypeSupplement
		sug.Advice = "在真实经验前提下，将 JD 高频关键词自然写入技能与项目描述（禁止捏造）"
		sug.Before = "未覆盖部分 JD 关键词"
		if analysisJD := resume.TargetJD; analysisJD != "" {
			sug.After = "在项目描述中体现真实使用过的相关技术与场景"
		} else {
			sug.After = "对照 JD 补齐真实具备的技能关键词"
		}
	case strings.Contains(issue.Description, "项目或工作"):
		sug.Type = domain.TypeSupplement
		sug.Advice = "至少补充 1~2 个能体现工程能力的项目，包含技术栈与个人贡献"
		sug.Before = "（缺失）"
		sug.After = "项目名称 | 技术栈 | 你的职责 | 量化成果"
	default:
		sug.Type = domain.TypeStructure
		sug.Advice = issue.Description
		sug.Before = issue.Location
		sug.After = "按建议调整对应模块"
	}
}

func firstVagueHighlight(resume *domain.Resume) string {
	phrases := []string{"负责开发", "参与开发", "负责后端"}
	for _, w := range resume.Structured.WorkExperience {
		for _, h := range w.Highlights {
			for _, p := range phrases {
				if strings.Contains(h, p) {
					return h
				}
			}
		}
	}
	for _, p := range resume.Structured.Projects {
		if p.Description != "" {
			return p.Description
		}
	}
	return ""
}

func extractQuoted(s string) string {
	start := strings.Index(s, "「")
	end := strings.Index(s, "」")
	if start >= 0 && end > start {
		return s[start+len("「") : end]
	}
	return ""
}

func formatSkills(sk domain.Skills) string {
	var parts []string
	if len(sk.Languages) > 0 {
		parts = append(parts, "语言："+strings.Join(sk.Languages, ", "))
	}
	if len(sk.Frameworks) > 0 {
		parts = append(parts, "框架："+strings.Join(sk.Frameworks, ", "))
	}
	if len(sk.Tools) > 0 {
		parts = append(parts, "工具："+strings.Join(sk.Tools, ", "))
	}
	if len(parts) == 0 {
		return "语言：Go, Java\n框架：Gin, Spring Boot\n工具：Docker, Redis, MySQL"
	}
	return strings.Join(parts, "\n")
}

// EstimateScoreGain estimates score improvement from accepted suggestions.
func EstimateScoreGain(suggestions []domain.Suggestion) float64 {
	var gain float64
	for _, s := range suggestions {
		if s.Status != domain.StatusAccepted && s.Status != domain.StatusEdited {
			continue
		}
		switch s.Priority {
		case domain.PriorityP0:
			gain += 4
		case domain.PriorityP1:
			gain += 2.5
		default:
			gain += 1
		}
	}
	if gain > 25 {
		return 25
	}
	return gain
}

// FormatAcceptedCount helper for API responses.
func FormatAcceptedCount(suggestions []domain.Suggestion) string {
	n := 0
	for _, s := range suggestions {
		if s.Status == domain.StatusAccepted || s.Status == domain.StatusEdited {
			n++
		}
	}
	return fmt.Sprintf("%d", n)
}
