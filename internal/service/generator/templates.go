package generator

import (
	"fmt"
	"strings"

	"github.com/cuixiangyu/ai-interview/internal/domain"
)

// ATSRenderer produces a plain, ATS-friendly markdown resume.
type ATSRenderer struct{}

func (ATSRenderer) ID() domain.TemplateID { return domain.TemplateATS }

func (ATSRenderer) Render(sr domain.StructuredResume) string {
	var b strings.Builder
	name := sr.BasicInfo.Name
	if name == "" {
		name = "姓名"
	}
	b.WriteString(name + "\n")
	b.WriteString(strings.Repeat("=", len([]rune(name))) + "\n\n")

	contact := joinNonEmpty(" | ", sr.BasicInfo.Phone, sr.BasicInfo.Email, sr.BasicInfo.Location)
	if contact != "" {
		b.WriteString(contact + "\n")
	}
	if len(sr.BasicInfo.Links) > 0 {
		b.WriteString(strings.Join(sr.BasicInfo.Links, " | ") + "\n")
	}
	b.WriteString("\n")

	if sr.Summary != "" {
		b.WriteString("SUMMARY\n-------\n")
		b.WriteString(sr.Summary + "\n\n")
	}

	writeSkillsATS(&b, sr)
	writeWorkATS(&b, sr)
	writeProjectsATS(&b, sr)
	writeEducationATS(&b, sr)
	return strings.TrimSpace(b.String()) + "\n"
}

// TechRenderer produces a developer-oriented markdown resume.
type TechRenderer struct{}

func (TechRenderer) ID() domain.TemplateID { return domain.TemplateTech }

func (TechRenderer) Render(sr domain.StructuredResume) string {
	var b strings.Builder
	name := sr.BasicInfo.Name
	if name == "" {
		name = "姓名"
	}
	b.WriteString("# " + name + "\n\n")

	contact := joinNonEmpty(" · ", sr.BasicInfo.Phone, sr.BasicInfo.Email, sr.BasicInfo.Location)
	if contact != "" {
		b.WriteString(contact + "\n\n")
	}
	for _, l := range sr.BasicInfo.Links {
		b.WriteString("- " + l + "\n")
	}
	if len(sr.BasicInfo.Links) > 0 {
		b.WriteString("\n")
	}

	if sr.Summary != "" {
		b.WriteString("## 简介\n\n" + sr.Summary + "\n\n")
	}

	b.WriteString("## 技术栈\n\n")
	if len(sr.Skills.Languages) > 0 {
		b.WriteString("- **语言**: " + strings.Join(sr.Skills.Languages, ", ") + "\n")
	}
	if len(sr.Skills.Frameworks) > 0 {
		b.WriteString("- **框架**: " + strings.Join(sr.Skills.Frameworks, ", ") + "\n")
	}
	if len(sr.Skills.Tools) > 0 {
		b.WriteString("- **工具**: " + strings.Join(sr.Skills.Tools, ", ") + "\n")
	}
	b.WriteString("\n")

	if len(sr.WorkExperience) > 0 {
		b.WriteString("## 工作经历\n\n")
		for _, w := range sr.WorkExperience {
			b.WriteString(fmt.Sprintf("### %s · %s\n\n", w.Company, w.Title))
			if w.Duration != "" {
				b.WriteString(w.Duration + "\n\n")
			}
			for _, h := range w.Highlights {
				b.WriteString("- " + h + "\n")
			}
			b.WriteString("\n")
		}
	}

	if len(sr.Projects) > 0 {
		b.WriteString("## 项目经历\n\n")
		for _, p := range sr.Projects {
			b.WriteString("### " + p.Name + "\n\n")
			if len(p.TechStack) > 0 {
				b.WriteString("`" + strings.Join(p.TechStack, "` · `") + "`\n\n")
			}
			if p.Description != "" {
				b.WriteString(p.Description + "\n\n")
			}
			for _, a := range p.Achievements {
				b.WriteString("- " + a + "\n")
			}
			b.WriteString("\n")
		}
	}

	if len(sr.Education) > 0 {
		b.WriteString("## 教育背景\n\n")
		for _, e := range sr.Education {
			line := e.School
			if e.Degree != "" {
				line += " · " + e.Degree
			}
			if e.Major != "" {
				line += " · " + e.Major
			}
			b.WriteString("- " + line)
			if e.Duration != "" {
				b.WriteString("（" + e.Duration + "）")
			}
			b.WriteString("\n")
		}
	}
	return strings.TrimSpace(b.String()) + "\n"
}

func writeSkillsATS(b *strings.Builder, sr domain.StructuredResume) {
	skills := joinNonEmpty(", ",
		strings.Join(sr.Skills.Languages, ", "),
		strings.Join(sr.Skills.Frameworks, ", "),
		strings.Join(sr.Skills.Tools, ", "),
	)
	if skills == "" {
		return
	}
	b.WriteString("SKILLS\n------\n")
	if len(sr.Skills.Languages) > 0 {
		b.WriteString("Languages: " + strings.Join(sr.Skills.Languages, ", ") + "\n")
	}
	if len(sr.Skills.Frameworks) > 0 {
		b.WriteString("Frameworks: " + strings.Join(sr.Skills.Frameworks, ", ") + "\n")
	}
	if len(sr.Skills.Tools) > 0 {
		b.WriteString("Tools: " + strings.Join(sr.Skills.Tools, ", ") + "\n")
	}
	b.WriteString("\n")
}

func writeWorkATS(b *strings.Builder, sr domain.StructuredResume) {
	if len(sr.WorkExperience) == 0 {
		return
	}
	b.WriteString("EXPERIENCE\n----------\n")
	for _, w := range sr.WorkExperience {
		b.WriteString(fmt.Sprintf("%s | %s", w.Company, w.Title))
		if w.Duration != "" {
			b.WriteString(" | " + w.Duration)
		}
		b.WriteString("\n")
		for _, h := range w.Highlights {
			b.WriteString("- " + h + "\n")
		}
		b.WriteString("\n")
	}
}

func writeProjectsATS(b *strings.Builder, sr domain.StructuredResume) {
	if len(sr.Projects) == 0 {
		return
	}
	b.WriteString("PROJECTS\n--------\n")
	for _, p := range sr.Projects {
		b.WriteString(p.Name + "\n")
		if len(p.TechStack) > 0 {
			b.WriteString("Tech: " + strings.Join(p.TechStack, ", ") + "\n")
		}
		if p.Description != "" {
			b.WriteString(p.Description + "\n")
		}
		for _, a := range p.Achievements {
			b.WriteString("- " + a + "\n")
		}
		b.WriteString("\n")
	}
}

func writeEducationATS(b *strings.Builder, sr domain.StructuredResume) {
	if len(sr.Education) == 0 {
		return
	}
	b.WriteString("EDUCATION\n---------\n")
	for _, e := range sr.Education {
		b.WriteString(e.School)
		if e.Degree != "" {
			b.WriteString(" | " + e.Degree)
		}
		if e.Major != "" {
			b.WriteString(" | " + e.Major)
		}
		if e.Duration != "" {
			b.WriteString(" | " + e.Duration)
		}
		b.WriteString("\n")
	}
}

func joinNonEmpty(sep string, parts ...string) string {
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, sep)
}
