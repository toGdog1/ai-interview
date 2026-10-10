package parser

import (
	"context"
	"strings"
	"testing"

	"github.com/cuixiangyu/ai-interview/internal/domain"
)

const sampleResume = `张三
13800138000
zhangsan@example.com
北京
https://github.com/zhangsan

专业技能
Go Java Python
Gin Spring Boot Redis MySQL Docker Kubernetes

工作经历
字节跳动
后端工程师
2021-07 - 2024-03
- 负责订单服务开发，将 P99 延迟从 800ms 降至 120ms
- 主导支付回调幂等改造

项目经历
AI 简历优化器
基于 Go 实现的简历解析与分析服务
- 支持多格式解析与评分，QPS 达 200+

教育背景
某某大学
本科 计算机科学
2017-09 - 2021-06
`

func TestRuleStructureBuilder(t *testing.T) {
	b := NewRuleStructureBuilder()
	sr, err := b.Build(context.Background(), sampleResume)
	if err != nil {
		t.Fatal(err)
	}
	if sr.BasicInfo.Name != "张三" {
		t.Errorf("name = %q", sr.BasicInfo.Name)
	}
	if sr.BasicInfo.Email != "zhangsan@example.com" {
		t.Errorf("email = %q", sr.BasicInfo.Email)
	}
	if sr.BasicInfo.Phone == "" {
		t.Error("expected phone")
	}
	if len(sr.Skills.Languages) == 0 {
		t.Error("expected languages")
	}
	if len(sr.WorkExperience) == 0 {
		t.Error("expected work experience")
	}
	if len(sr.Projects) == 0 {
		t.Error("expected projects")
	}
	if len(sr.Education) == 0 {
		t.Error("expected education")
	}
	foundGitHub := false
	for _, l := range sr.BasicInfo.Links {
		if strings.Contains(l, "github") {
			foundGitHub = true
		}
	}
	if !foundGitHub {
		t.Error("expected github link")
	}
}

func TestParserPasteText(t *testing.T) {
	p := New(NewRuleStructureBuilder(), 1024, DefaultExtractors()...)
	raw, sr, err := p.Parse(context.Background(), ParseInput{Text: sampleResume})
	if err != nil {
		t.Fatal(err)
	}
	if raw == "" || sr.BasicInfo.Name == "" {
		t.Fatalf("unexpected parse result: raw empty=%v name=%q", raw == "", sr.BasicInfo.Name)
	}
}

func TestParserEmpty(t *testing.T) {
	p := New(NewRuleStructureBuilder(), 1024, DefaultExtractors()...)
	_, _, err := p.Parse(context.Background(), ParseInput{})
	if err != domain.ErrEmptyContent {
		t.Fatalf("want ErrEmptyContent, got %v", err)
	}
}

func TestParserTooLarge(t *testing.T) {
	p := New(NewRuleStructureBuilder(), 10, DefaultExtractors()...)
	_, _, err := p.Parse(context.Background(), ParseInput{
		Content: []byte("01234567890"),
		Format:  domain.FormatTXT,
	})
	if err != domain.ErrFileTooLarge {
		t.Fatalf("want ErrFileTooLarge, got %v", err)
	}
}

func TestDetectFormat(t *testing.T) {
	if DetectFormat([]byte("%PDF-1.4")) != domain.FormatPDF {
		t.Error("pdf")
	}
	if DetectFormat([]byte("# Title\n## Skills")) != domain.FormatMD {
		t.Error("md")
	}
	if DetectFormat([]byte("plain text resume")) != domain.FormatTXT {
		t.Error("txt")
	}
}

func TestPlainTextExtractorInvalidUTF8(t *testing.T) {
	_, err := PlainTextExtractor{}.Extract(context.Background(), []byte{0xff, 0xfe, 0xfd})
	if err == nil {
		t.Fatal("expected error")
	}
}
