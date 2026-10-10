package domain

import "testing"

func TestScoreToGrade(t *testing.T) {
	cases := []struct {
		score float64
		want  Grade
	}{
		{95, GradeA},
		{90, GradeA},
		{89.9, GradeB},
		{75, GradeB},
		{60, GradeC},
		{59.9, GradeD},
		{0, GradeD},
	}
	for _, tc := range cases {
		if got := ScoreToGrade(tc.score); got != tc.want {
			t.Errorf("ScoreToGrade(%v) = %v, want %v", tc.score, got, tc.want)
		}
	}
}

func TestDimensionScoresWeightedOverall(t *testing.T) {
	d := DimensionScores{
		Completeness: 100,
		TechExpress:  100,
		ProjectQual:  100,
		JDMatch:      100,
		Readability:  100,
	}
	if got := d.WeightedOverall(); got != 100 {
		t.Fatalf("expected 100, got %v", got)
	}

	d = DimensionScores{
		Completeness: 50, // 10
		TechExpress:  80, // 20
		ProjectQual:  60, // 15
		JDMatch:      100, // 20
		Readability:  40, // 4
	}
	want := 10 + 20 + 15 + 20 + 4
	if got := d.WeightedOverall(); got != float64(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
