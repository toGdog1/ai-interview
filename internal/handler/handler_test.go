package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cuixiangyu/ai-interview/internal/app"
	"github.com/cuixiangyu/ai-interview/internal/config"
)

func TestHealthAndAnalyzeFlow(t *testing.T) {
	router := app.New(config.Config{
		MaxUploadBytes: 10 * 1024 * 1024,
		EnableCORS:     true,
		AIProvider:     "rule",
	})

	// health
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("health=%d", rr.Code)
	}

	body := map[string]interface{}{
		"text": `赵六
zhaoliu@example.com
13700137000

专业技能
Go Python Redis

工作经历
公司A
工程师
2020-01 - 2023-01
- 负责开发

项目经历
系统B
Go 微服务
- QPS 提升 30%

教育背景
大学C
本科
2016-2010
`,
		"targetRole": "后端工程师",
		"targetJd":   "Go Redis",
	}
	raw, _ := json.Marshal(body)
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/resumes/analyze", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("analyze=%d body=%s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Data struct {
			Resume struct {
				ID string `json:"id"`
			} `json:"resume"`
			Analysis struct {
				ID           string  `json:"id"`
				OverallScore float64 `json:"overallScore"`
			} `json:"analysis"`
			Suggestions []struct {
				ID string `json:"id"`
			} `json:"suggestions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Resume.ID == "" || resp.Data.Analysis.ID == "" {
		t.Fatalf("bad response: %s", rr.Body.String())
	}

	// generate
	genBody, _ := json.Marshal(map[string]interface{}{
		"resumeId": resp.Data.Resume.ID,
		"template": "ats",
	})
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/versions/generate", bytes.NewReader(genBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("generate=%d body=%s", rr.Code, rr.Body.String())
	}

	var genResp struct {
		Data struct {
			ID         string `json:"id"`
			RenderedMd string `json:"renderedMd"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &genResp); err != nil {
		t.Fatal(err)
	}
	if genResp.Data.ID == "" || genResp.Data.RenderedMd == "" {
		t.Fatal("empty version")
	}

	// export
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/versions/"+genResp.Data.ID+"/export?format=markdown", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("export=%d", rr.Code)
	}
}

func TestAnalyzeEmptyBody(t *testing.T) {
	router := app.New(config.Config{MaxUploadBytes: 1024, AIProvider: "rule"})
	raw, _ := json.Marshal(map[string]string{"text": ""})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resumes/analyze", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
}
