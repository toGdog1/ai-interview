package parser

import (
	"context"
	"regexp"
	"strings"

	"github.com/cuixiangyu/ai-interview/internal/domain"
)

var (
	emailRe = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	phoneRe = regexp.MustCompile(`(?:(?:\+?86[-\s]?)?1[3-9]\d{9})|(?:\d{3,4}[-\s]?\d{7,8})`)
	urlRe   = regexp.MustCompile(`https?://[^\s]+|github\.com/[^\s]+|linkedin\.com/[^\s]+`)

	sectionHeaders = map[string]string{
		"基本信息": "basic", "个人信息": "basic", "联系方式": "basic",
		"个人简介": "summary", "自我评价": "summary", "简介": "summary", "summary": "summary",
		"技能": "skills", "专业技能": "skills", "技术栈": "skills", "skills": "skills",
		"工作经历": "work", "工作经验": "work", "实习经历": "work", "experience": "work",
		"项目经历": "projects", "项目经验": "projects", "个人项目": "projects", "projects": "projects",
		"教育背景": "education", "教育经历": "education", "学历": "education", "education": "education",
		"证书": "certs", "获奖": "certs", "certifications": "certs",
	}

	langKeywords = []string{
		"go", "golang", "java", "python", "javascript", "typescript", "c++", "c#", "rust",
		"kotlin", "swift", "php", "ruby", "scala", "sql",
	}
	fwKeywords = []string{
		"spring", "spring boot", "gin", "echo", "django", "flask", "fastapi", "react",
		"vue", "angular", "next.js", "nestjs", "grpc", "hibernate", "mybatis",
	}
	toolKeywords = []string{
		"docker", "kubernetes", "k8s", "git", "jenkins", "ci/cd", "redis", "mysql",
		"postgresql", "mongodb", "kafka", "rabbitmq", "nginx", "aws", "aliyun", "prometheus",
	}
)

// RuleStructureBuilder parses resume text with heuristics (no LLM required).
// Suitable for MVP; can be replaced by an LLM-backed StructureBuilder.
type RuleStructureBuilder struct{}

// NewRuleStructureBuilder creates a rule-based structure builder.
func NewRuleStructureBuilder() *RuleStructureBuilder {
	return &RuleStructureBuilder{}
}

// Build converts raw resume text into a StructuredResume.
func (b *RuleStructureBuilder) Build(_ context.Context, text string) (domain.StructuredResume, error) {
	text = NormalizeWhitespace(text)
	sections := splitSections(text)

	var sr domain.StructuredResume
	sr.BasicInfo = extractBasicInfo(text, sections["basic"])
	sr.Summary = firstNonEmpty(sections["summary"])
	sr.Skills = extractSkills(firstNonEmpty(sections["skills"], text))
	sr.WorkExperience = extractWork(sections["work"])
	sr.Projects = extractProjects(sections["projects"])
	sr.Education = extractEducation(sections["education"])
	sr.Certifications = extractLines(sections["certs"])

	// Fallback: if no name found, take first non-empty line that isn't a section header.
	if sr.BasicInfo.Name == "" {
		sr.BasicInfo.Name = guessName(text)
	}
	return sr, nil
}

func splitSections(text string) map[string]string {
	lines := strings.Split(text, "\n")
	result := make(map[string]string)
	current := "basic"
	var buf []string

	flush := func() {
		if len(buf) > 0 {
			result[current] = strings.TrimSpace(strings.Join(buf, "\n"))
			buf = nil
		}
	}

	for _, line := range lines {
		key := matchSection(line)
		if key != "" {
			flush()
			current = key
			continue
		}
		buf = append(buf, line)
	}
	flush()
	return result
}

func matchSection(line string) string {
	clean := strings.TrimSpace(line)
	clean = strings.TrimLeft(clean, "#*- ")
	clean = strings.TrimRight(clean, ":：")
	lower := strings.ToLower(clean)
	if key, ok := sectionHeaders[lower]; ok {
		return key
	}
	if key, ok := sectionHeaders[clean]; ok {
		return key
	}
	return ""
}

func extractBasicInfo(full, section string) domain.BasicInfo {
	src := section
	if src == "" {
		src = full
	}
	info := domain.BasicInfo{}
	if m := emailRe.FindString(src); m != "" {
		info.Email = m
	}
	if m := phoneRe.FindString(src); m != "" {
		info.Phone = m
	}
	info.Links = uniqueStrings(urlRe.FindAllString(src, -1))

	for _, loc := range []string{"北京", "上海", "广州", "深圳", "杭州", "成都", "南京", "武汉", "西安"} {
		if strings.Contains(src, loc) {
			info.Location = loc
			break
		}
	}
	return info
}

func guessName(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || matchSection(line) != "" {
			continue
		}
		if emailRe.MatchString(line) || phoneRe.MatchString(line) {
			continue
		}
		// Chinese name: 2-4 chars; or short Latin name
		runes := []rune(line)
		if len(runes) >= 2 && len(runes) <= 4 && !strings.ContainsAny(line, "@/\\") {
			return line
		}
		if len(runes) <= 40 && !strings.Contains(line, "：") && !strings.Contains(line, ":") {
			return line
		}
	}
	return ""
}

func extractSkills(text string) domain.Skills {
	lower := strings.ToLower(text)
	var s domain.Skills
	for _, k := range langKeywords {
		if strings.Contains(lower, k) {
			s.Languages = append(s.Languages, displaySkill(k))
		}
	}
	for _, k := range fwKeywords {
		if strings.Contains(lower, k) {
			s.Frameworks = append(s.Frameworks, displaySkill(k))
		}
	}
	for _, k := range toolKeywords {
		if strings.Contains(lower, k) {
			s.Tools = append(s.Tools, displaySkill(k))
		}
	}
	s.Languages = uniqueStrings(s.Languages)
	s.Frameworks = uniqueStrings(s.Frameworks)
	s.Tools = uniqueStrings(s.Tools)
	return s
}

func displaySkill(k string) string {
	switch k {
	case "golang":
		return "Go"
	case "javascript":
		return "JavaScript"
	case "typescript":
		return "TypeScript"
	case "spring boot":
		return "Spring Boot"
	case "next.js":
		return "Next.js"
	case "ci/cd":
		return "CI/CD"
	case "k8s":
		return "Kubernetes"
	case "postgresql":
		return "PostgreSQL"
	default:
		if len(k) <= 3 {
			return strings.ToUpper(k)
		}
		return strings.Title(k)
	}
}

func extractWork(section string) []domain.WorkExperience {
	if strings.TrimSpace(section) == "" {
		return nil
	}
	blocks := splitBlocks(section)
	var out []domain.WorkExperience
	for _, block := range blocks {
		lines := nonEmptyLines(block)
		if len(lines) == 0 {
			continue
		}
		we := domain.WorkExperience{Company: lines[0]}
		if len(lines) > 1 {
			we.Title = lines[1]
		}
		for _, l := range lines {
			if looksLikeDuration(l) {
				we.Duration = l
				break
			}
		}
		for _, l := range lines[1:] {
			if strings.HasPrefix(l, "-") || strings.HasPrefix(l, "•") || strings.HasPrefix(l, "*") {
				we.Highlights = append(we.Highlights, strings.TrimLeft(l, "-•* "))
			}
		}
		if len(we.Highlights) == 0 && len(lines) > 2 {
			we.Highlights = lines[2:]
		}
		out = append(out, we)
	}
	return out
}

func extractProjects(section string) []domain.Project {
	if strings.TrimSpace(section) == "" {
		return nil
	}
	blocks := splitBlocks(section)
	var out []domain.Project
	for _, block := range blocks {
		lines := nonEmptyLines(block)
		if len(lines) == 0 {
			continue
		}
		p := domain.Project{Name: lines[0]}
		var desc []string
		for _, l := range lines[1:] {
			trim := strings.TrimLeft(l, "-•* ")
			if strings.HasPrefix(l, "-") || strings.HasPrefix(l, "•") || strings.HasPrefix(l, "*") {
				p.Achievements = append(p.Achievements, trim)
			} else {
				desc = append(desc, trim)
			}
		}
		p.Description = strings.Join(desc, " ")
		// tech stack from description
		skills := extractSkills(block)
		p.TechStack = append(p.TechStack, skills.Languages...)
		p.TechStack = append(p.TechStack, skills.Frameworks...)
		p.TechStack = uniqueStrings(p.TechStack)
		out = append(out, p)
	}
	return out
}

func extractEducation(section string) []domain.Education {
	if strings.TrimSpace(section) == "" {
		return nil
	}
	var out []domain.Education
	for _, block := range splitBlocks(section) {
		lines := nonEmptyLines(block)
		if len(lines) == 0 {
			continue
		}
		ed := domain.Education{School: lines[0]}
		if len(lines) > 1 {
			ed.Degree = lines[1]
		}
		for _, l := range lines {
			if looksLikeDuration(l) {
				ed.Duration = l
			}
			if strings.Contains(l, "专业") || strings.Contains(strings.ToLower(l), "major") {
				ed.Major = l
			}
		}
		out = append(out, ed)
	}
	return out
}

func extractLines(section string) []string {
	return nonEmptyLines(section)
}

func splitBlocks(section string) []string {
	// Split on blank lines
	parts := regexp.MustCompile(`\n\s*\n`).Split(section, -1)
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 && strings.TrimSpace(section) != "" {
		return []string{strings.TrimSpace(section)}
	}
	return out
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}

func looksLikeDuration(s string) bool {
	return regexp.MustCompile(`\d{4}`).MatchString(s) &&
		(strings.Contains(s, "-") || strings.Contains(s, "至") || strings.Contains(s, "~") || strings.Contains(s, "—"))
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		key := strings.ToLower(s)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}
	return out
}
