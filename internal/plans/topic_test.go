package plans

import "testing"

func TestIsValidTopic(t *testing.T) {
	notValidVals := []Topic{-1, 100}

	for _, val := range notValidVals {
		if IsValidTopic(val) {
			t.Errorf("IsValidTopic(%v) = true, want false", val)
		}
	}

	validVals := []Topic{Empty, Places, AtHome, Travel, Other}
	for _, val := range validVals {
		if !IsValidTopic(val) {
			t.Errorf("IsValidTopic(%v) = false, want true", val)
		}
	}
}
