package llm

import "testing"

func TestEnforceThreeSentences(t *testing.T) {
	got := enforceThreeSentences("Раз. Два. Три. Четыре.")
	want := "Раз. Два. Три."
	if got != want {
		t.Fatalf("unexpected truncated text: got=%q want=%q", got, want)
	}
}

func TestEnforceThreeSentencesWithoutDots(t *testing.T) {
	input := "Короткий ответ без точки"
	want := "Короткий ответ без точки."
	if got := enforceThreeSentences(input); got != want {
		t.Fatalf("unexpected normalized text: got=%q want=%q", got, want)
	}
}

func TestEnforceThreeSentencesEmpty(t *testing.T) {
	if got := enforceThreeSentences("   "); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}
