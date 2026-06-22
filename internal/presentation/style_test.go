package presentation

import "testing"

func TestStylesHonorNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	got := Brand("◉ HEIMDALL")
	if got != "◉ HEIMDALL" {
		t.Fatalf("Brand() = %q, want unstyled text with NO_COLOR", got)
	}
}
