package bot

import (
	"context"
	"testing"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/shai38/betweenus/internal/couples"
	"github.com/shai38/betweenus/internal/plans"
)

type fakeCallbackBot struct {
	edits   []*tgBot.EditMessageTextParams
	answers []*tgBot.AnswerCallbackQueryParams
}

func (f *fakeCallbackBot) EditMessageText(
	_ context.Context,
	params *tgBot.EditMessageTextParams,
) (*models.Message, error) {
	f.edits = append(f.edits, params)
	return nil, nil
}

func (f *fakeCallbackBot) AnswerCallbackQuery(
	_ context.Context,
	params *tgBot.AnswerCallbackQueryParams,
) (bool, error) {
	f.answers = append(f.answers, params)
	return true, nil
}

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

func TestHandleOpenPlan_ShowsCardWithAndWithoutNote(t *testing.T) {
	const (
		userID   int64 = 123
		coupleID int64 = 789
		planID   int64 = 1
	)
	for _, testCase := range []struct {
		name string
		note string
		want string
	}{
		{name: "with note", note: "В пятницу", want: "Вечер кино \n\nЗаметка:\nВ пятницу\n"},
		{name: "without note", want: "Вечер кино \n\n"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			callbackBot := &fakeCallbackBot{}
			rep := &fakeBotRepository{
				fakeRepository: &fakeRepository{
					addedPlan: plans.Plan{
						ID:       planID,
						NewPlan:  plans.NewPlan{Title: "Вечер кино", Note: testCase.note},
						CoupleID: coupleID,
					},
				},
				couple: couples.Couple{ID: coupleID, FirstUserID: userID, SecondUserID: 321},
			}
			query := testCallbackQuery(userID, planOpenPrefix+"1")

			handleOpenPlan(context.Background(), callbackBot, query, rep)

			if len(callbackBot.edits) != 1 {
				t.Fatalf("EditMessageText calls = %d, want 1", len(callbackBot.edits))
			}
			if got := callbackBot.edits[0].Text; got != testCase.want {
				t.Errorf("card text = %q, want %q", got, testCase.want)
			}
			if len(callbackBot.answers) != 1 {
				t.Errorf("AnswerCallbackQuery calls = %d, want 1", len(callbackBot.answers))
			}
		})
	}
}

func TestHandleSetPlanStatus_MarksDoneAndActive(t *testing.T) {
	const (
		userID   int64 = 123
		coupleID int64 = 789
		planID   int64 = 1
	)
	for _, testCase := range []struct {
		name         string
		initialDone  bool
		callbackData string
		wantDone     bool
		wantButton   string
	}{
		{
			name:         "mark done",
			callbackData: planSetDonePrefix + "1",
			wantDone:     true,
			wantButton:   "↩️ Вернуть в активные",
		},
		{
			name:         "return to active",
			initialDone:  true,
			callbackData: planSetActivePrefix + "1",
			wantDone:     false,
			wantButton:   "✅ Выполнить",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			callbackBot := &fakeCallbackBot{}
			rep := &fakeBotRepository{
				fakeRepository: &fakeRepository{
					addedPlan: plans.Plan{
						ID:       planID,
						Done:     testCase.initialDone,
						NewPlan:  plans.NewPlan{Title: "Вечер кино"},
						CoupleID: coupleID,
					},
				},
				couple: couples.Couple{ID: coupleID, FirstUserID: userID, SecondUserID: 321},
			}
			query := testCallbackQuery(userID, testCase.callbackData)

			handleSetPlanStatus(
				context.Background(),
				callbackBot,
				query,
				rep,
				testCase.wantDone,
			)

			if rep.addedPlan.Done != testCase.wantDone {
				t.Errorf("stored Done = %v, want %v", rep.addedPlan.Done, testCase.wantDone)
			}
			if len(callbackBot.edits) != 1 {
				t.Fatalf("EditMessageText calls = %d, want 1", len(callbackBot.edits))
			}
			keyboard, ok := callbackBot.edits[0].ReplyMarkup.(*models.InlineKeyboardMarkup)
			if !ok {
				t.Fatalf(
					"edited card keyboard type = %T, want *models.InlineKeyboardMarkup",
					callbackBot.edits[0].ReplyMarkup,
				)
			}
			if got := keyboard.InlineKeyboard[0][0].Text; got != testCase.wantButton {
				t.Errorf("status button = %q, want %q", got, testCase.wantButton)
			}
			if len(callbackBot.answers) != 1 {
				t.Errorf("AnswerCallbackQuery calls = %d, want 1", len(callbackBot.answers))
			}
		})
	}
}

func testCallbackQuery(userID int64, data string) *models.CallbackQuery {
	return &models.CallbackQuery{
		ID:   "callback-id",
		From: models.User{ID: userID},
		Message: models.MaybeInaccessibleMessage{
			Message: &models.Message{ID: 55},
		},
		Data: data,
	}
}
