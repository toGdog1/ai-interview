package exporter

import (
	"context"
	"strings"
	"testing"

	"github.com/cuixiangyu/ai-interview/internal/domain"
)

func TestMarkdownExport(t *testing.T) {
	reg := Default()
	ver := &domain.ResumeVersion{
		Content:    domain.StructuredResume{BasicInfo: domain.BasicInfo{Name: "张三"}},
		RenderedMD: "# 张三\n\nGo Engineer\n",
	}
	name, ct, data, err := reg.Export(context.Background(), domain.ExportMarkdown, ver)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(name, ".md") {
		t.Fatalf("filename=%s", name)
	}
	if !strings.Contains(ct, "markdown") {
		t.Fatalf("ct=%s", ct)
	}
	if string(data) != ver.RenderedMD {
		t.Fatal("content mismatch")
	}
}

func TestPDFExport(t *testing.T) {
	reg := Default()
	ver := &domain.ResumeVersion{
		Content:    domain.StructuredResume{BasicInfo: domain.BasicInfo{Name: "TestUser"}},
		RenderedMD: "TestUser\nGo Redis Docker\n",
	}
	name, ct, data, err := reg.Export(context.Background(), domain.ExportPDF, ver)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(name, ".pdf") {
		t.Fatalf("filename=%s", name)
	}
	if ct != "application/pdf" {
		t.Fatalf("ct=%s", ct)
	}
	if !strings.HasPrefix(string(data), "%PDF") {
		t.Fatal("not a pdf")
	}
}

func TestUnsupportedFormat(t *testing.T) {
	_, _, _, err := Default().Export(context.Background(), domain.ExportDOCX, &domain.ResumeVersion{})
	if err == nil {
		t.Fatal("expected error")
	}
}
