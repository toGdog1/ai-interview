package memory

import (
	"context"
	"testing"
	"time"

	"github.com/cuixiangyu/ai-interview/internal/domain"
)

func TestStoreResumeCRUD(t *testing.T) {
	s := NewStore()
	ctx := context.Background()
	r := &domain.Resume{ID: "r1", CreatedAt: time.Now()}
	if err := s.Save(ctx, r); err != nil {
		t.Fatal(err)
	}
	got, err := s.FindByID(ctx, "r1")
	if err != nil || got.ID != "r1" {
		t.Fatalf("got=%v err=%v", got, err)
	}
	r.TargetRole = "Go"
	if err := s.Update(ctx, r); err != nil {
		t.Fatal(err)
	}
	got, _ = s.FindByID(ctx, "r1")
	if got.TargetRole != "Go" {
		t.Fatal("update failed")
	}
	if err := s.Delete(ctx, "r1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.FindByID(ctx, "r1"); err != domain.ErrNotFound {
		t.Fatalf("want not found, got %v", err)
	}
}

func TestStoreAnalysisAndSuggestions(t *testing.T) {
	s := NewStore()
	ctx := context.Background()
	a := memoryAnalysisAdapter(s)
	sug := SuggestionAdapter{Store: s}

	an := &domain.Analysis{ID: "a1", ResumeID: "r1"}
	if err := a.Save(ctx, an); err != nil {
		t.Fatal(err)
	}
	got, err := a.FindByResumeID(ctx, "r1")
	if err != nil || got.ID != "a1" {
		t.Fatalf("got=%v err=%v", got, err)
	}

	list := []domain.Suggestion{
		{ID: "s1", AnalysisID: "a1", Status: domain.StatusPending},
		{ID: "s2", AnalysisID: "a1", Status: domain.StatusPending},
	}
	if err := sug.SaveBatch(ctx, list); err != nil {
		t.Fatal(err)
	}
	found, err := sug.FindByAnalysisID(ctx, "a1")
	if err != nil || len(found) != 2 {
		t.Fatalf("len=%d err=%v", len(found), err)
	}
	found[0].Status = domain.StatusAccepted
	if err := sug.Update(ctx, &found[0]); err != nil {
		t.Fatal(err)
	}
}

func memoryAnalysisAdapter(s *Store) AnalysisAdapter {
	return AnalysisAdapter{Store: s}
}

func TestStoreVersions(t *testing.T) {
	s := NewStore()
	v := VersionAdapter{Store: s}
	ctx := context.Background()
	ver := &domain.ResumeVersion{ID: "v1", ResumeID: "r1", RenderedMD: "hi"}
	if err := v.Save(ctx, ver); err != nil {
		t.Fatal(err)
	}
	list, err := v.FindByResumeID(ctx, "r1")
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%v err=%v", list, err)
	}
}
