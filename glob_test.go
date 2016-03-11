package glob

import "testing"

func TestGlobMatch_zeroTrailing(t *testing.T) {
	pat, err := NewPattern(`PLAN9*`)
	if err != nil {
		t.Fatal(err)
	}

	mustMatch := []string{"PLAN9", "PLAN9_foo", "PLAN9_"}
	noMatch := []string{"nope", "PLAN", "PLAN8", "PLAN8_foo"}

	for _, m := range mustMatch {
		if !pat.Matches(m) {
			t.Errorf("Expected %q to match", m)
		}
	}

	for _, m := range noMatch {
		if pat.Matches(m) {
			t.Errorf("Did not expect %q to match", m)
		}
	}
}
