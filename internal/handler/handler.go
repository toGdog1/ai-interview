package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/cuixiangyu/ai-interview/internal/domain"
	"github.com/cuixiangyu/ai-interview/internal/service/resume"
	"github.com/cuixiangyu/ai-interview/internal/service/suggestion"
	"github.com/cuixiangyu/ai-interview/pkg/response"
)

// Handler exposes HTTP endpoints for the resume optimization API.
type Handler struct {
	svc        *resume.Service
	maxUpload  int64
}

// New creates an HTTP handler.
func New(svc *resume.Service, maxUpload int64) *Handler {
	if maxUpload <= 0 {
		maxUpload = 10 * 1024 * 1024
	}
	return &Handler{svc: svc, maxUpload: maxUpload}
}

type uploadJSONRequest struct {
	Text       string `json:"text"`
	TargetRole string `json:"targetRole"`
	TargetJD   string `json:"targetJd"`
	YearsOfExp int    `json:"yearsOfExp"`
	Format     string `json:"format"`
}

// Health returns service health.
func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "up"})
}

// UploadAndAnalyze handles JSON paste or multipart file upload.
func (h *Handler) UploadAndAnalyze(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ct := r.Header.Get("Content-Type")

	var req resume.UploadRequest

	switch {
	case strings.HasPrefix(ct, "multipart/form-data"):
		if err := r.ParseMultipartForm(h.maxUpload); err != nil {
			response.Error(w, http.StatusBadRequest, "invalid multipart form")
			return
		}
		req.TargetRole = r.FormValue("targetRole")
		req.TargetJD = r.FormValue("targetJd")
		req.Text = r.FormValue("text")
		if file, header, err := r.FormFile("file"); err == nil {
			defer file.Close()
			data, err := io.ReadAll(io.LimitReader(file, h.maxUpload+1))
			if err != nil {
				response.Error(w, http.StatusBadRequest, "failed to read file")
				return
			}
			if int64(len(data)) > h.maxUpload {
				response.Error(w, http.StatusRequestEntityTooLarge, domain.ErrFileTooLarge.Error())
				return
			}
			req.Content = data
			req.Format = formatFromFilename(header.Filename)
		}
	default:
		var body uploadJSONRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			response.Error(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		req.Text = body.Text
		req.TargetRole = body.TargetRole
		req.TargetJD = body.TargetJD
		req.YearsOfExp = body.YearsOfExp
		req.Format = domain.FileFormat(body.Format)
	}

	result, err := h.svc.UploadAndAnalyze(ctx, req)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, result)
}

// GetResume returns a resume by id.
func (h *Handler) GetResume(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/resumes/")
	id = strings.Split(id, "/")[0]
	res, err := h.svc.GetResume(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, res)
}

// UpdateStructured corrects parsed resume fields.
func (h *Handler) UpdateStructured(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/api/v1/resumes/", "/structured")
	var structured domain.StructuredResume
	if err := json.NewDecoder(r.Body).Decode(&structured); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	res, err := h.svc.UpdateStructured(r.Context(), id, structured)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, res)
}

// GetAnalysisByResume returns analysis for a resume.
func (h *Handler) GetAnalysisByResume(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/api/v1/resumes/", "/analysis")
	an, err := h.svc.GetAnalysisByResume(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, an)
}

// ListSuggestions returns suggestions for an analysis.
func (h *Handler) ListSuggestions(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/api/v1/analyses/", "/suggestions")
	list, err := h.svc.ListSuggestions(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	accepted, gain, _ := h.svc.SuggestionSummary(r.Context(), id)
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"suggestions":    list,
		"acceptedCount":  accepted,
		"estimatedGain":  gain,
		"acceptedLabel":  suggestion.FormatAcceptedCount(list),
	})
}

type suggestionStatusRequest struct {
	Status string `json:"status"`
	After  string `json:"after,omitempty"`
}

// UpdateSuggestion updates suggestion status.
func (h *Handler) UpdateSuggestion(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/suggestions/")
	var body suggestionStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	sug, err := h.svc.UpdateSuggestionStatus(r.Context(), id, domain.SuggestionStatus(body.Status), body.After)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, sug)
}

type generateRequest struct {
	ResumeID      string   `json:"resumeId"`
	Template      string   `json:"template"`
	SuggestionIDs []string `json:"suggestionIds"`
}

// Generate creates an optimized resume version.
func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	var body generateRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	ver, err := h.svc.Generate(r.Context(), resume.GenerateRequest{
		ResumeID:      body.ResumeID,
		Template:      domain.TemplateID(body.Template),
		SuggestionIDs: body.SuggestionIDs,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, ver)
}

// GetVersion returns a generated version.
func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/versions/")
	id = strings.Split(id, "/")[0]
	ver, err := h.svc.GetVersion(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, ver)
}

// Export downloads a version.
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	// /api/v1/versions/{id}/export?format=markdown|pdf
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/versions/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "export" {
		response.Error(w, http.StatusNotFound, "not found")
		return
	}
	format := domain.ExportFormat(r.URL.Query().Get("format"))
	if format == "" {
		format = domain.ExportMarkdown
	}
	filename, contentType, data, err := h.svc.Export(r.Context(), parts[0], format)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// Compare returns original vs optimized content.
func (h *Handler) Compare(w http.ResponseWriter, r *http.Request) {
	resumeID := r.URL.Query().Get("resumeId")
	versionID := r.URL.Query().Get("versionId")
	orig, opt, err := h.svc.Compare(r.Context(), resumeID, versionID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{
		"original":  orig,
		"optimized": opt,
	})
}

// DeleteResume deletes a resume.
func (h *Handler) DeleteResume(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/resumes/")
	id = strings.Split(id, "/")[0]
	if err := h.svc.DeleteResume(r.Context(), id); err != nil {
		writeDomainError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"deleted": id})
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrEmptyContent),
		errors.Is(err, domain.ErrUnsupportedFormat),
		errors.Is(err, domain.ErrInvalidStatus),
		errors.Is(err, domain.ErrParseFailed):
		response.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrFileTooLarge):
		response.Error(w, http.StatusRequestEntityTooLarge, err.Error())
	case errors.Is(err, domain.ErrExportUnsupported):
		response.Error(w, http.StatusNotImplemented, err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, err.Error())
	}
}

func formatFromFilename(name string) domain.FileFormat {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".pdf"):
		return domain.FormatPDF
	case strings.HasSuffix(lower, ".docx"):
		return domain.FormatDOCX
	case strings.HasSuffix(lower, ".md"), strings.HasSuffix(lower, ".markdown"):
		return domain.FormatMD
	default:
		return domain.FormatTXT
	}
}

func pathID(path, prefix, suffix string) string {
	s := strings.TrimPrefix(path, prefix)
	s = strings.TrimSuffix(s, suffix)
	return strings.Trim(s, "/")
}
