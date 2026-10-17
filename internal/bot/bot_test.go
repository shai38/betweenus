package bot

import (
	"testing"

	"github.com/shai38/betweenus/internal/plans"
)

func TestFormatPlansEmpty(t *testing.T) {
	resultText := formatPlans([]plans.Plan{})
	if resultText != ZeroPlansText {
		t.Errorf("formatPlans text error, got %v, want %v", resultText, ZeroPlansText)
	}
}
