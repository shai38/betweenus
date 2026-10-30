package bot

import (
	"testing"

	"github.com/go-telegram/bot/models"
	"github.com/shai38/betweenus/internal/plans"
)

func TestBuildPlansMessage_Empty(t *testing.T) {
	text, keyboard := buildPlansMessage(nil)

	if text != ZeroPlansText {
		t.Errorf("text = %q, want %q", text, ZeroPlansText)
	}
	if keyboard != nil {
		inlineKeyboard, ok := keyboard.(*models.InlineKeyboardMarkup)
		if !ok || inlineKeyboard != nil {
			t.Errorf("keyboard = %#v, want nil markup", keyboard)
		}
	}
}

func TestBuildPlansMessage_WithPlans(t *testing.T) {
	planList := []plans.Plan{
		{ID: 1, NewPlan: plans.NewPlan{Title: "Вечер кино"}},
		{ID: 2, Done: true, NewPlan: plans.NewPlan{Title: "Поездка"}},
	}

	text, keyboard := buildPlansMessage(planList)

	wantText := "💞 Наши планы\n\nНажми на план, чтобы посмотреть заметку и действия.\n"
	if text != wantText {
		t.Errorf("text = %q, want %q", text, wantText)
	}

	inlineKeyboard, ok := keyboard.(*models.InlineKeyboardMarkup)
	if !ok {
		t.Fatalf("keyboard type = %T, want *models.InlineKeyboardMarkup", keyboard)
	}
	if len(inlineKeyboard.InlineKeyboard) != 2 {
		t.Fatalf("keyboard rows = %d, want 2", len(inlineKeyboard.InlineKeyboard))
	}

	wantButtons := []models.InlineKeyboardButton{
		{Text: " Вечер кино", CallbackData: "plan:open:1"},
		{Text: "✅ Поездка", CallbackData: "plan:open:2"},
	}
	for index, want := range wantButtons {
		if len(inlineKeyboard.InlineKeyboard[index]) != 1 {
			t.Fatalf("row %d buttons = %d, want 1", index, len(inlineKeyboard.InlineKeyboard[index]))
		}
		got := inlineKeyboard.InlineKeyboard[index][0]
		if got != want {
			t.Errorf("button %d = %+v, want %+v", index, got, want)
		}
	}
}
