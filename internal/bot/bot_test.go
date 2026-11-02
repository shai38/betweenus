package bot

import (
	"context"
	"fmt"
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

	if text != zeroPlansText {
		t.Errorf("text = %q, want %q", text, zeroPlansText)
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
			wantButton:   "↩️ В активные",
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

func TestCancelPlanDeletion_DoesNotDeletePlan(t *testing.T) {
	const (
		userID   int64 = 123
		coupleID int64 = 789
		planID   int64 = 1
	)
	callbackBot := &fakeCallbackBot{}
	rep := &fakeBotRepository{
		fakeRepository: &fakeRepository{
			addedPlan: plans.Plan{
				ID:       planID,
				NewPlan:  plans.NewPlan{Title: "Вечер кино"},
				CoupleID: coupleID,
			},
		},
		couple: couples.Couple{ID: coupleID, FirstUserID: userID, SecondUserID: 321},
	}
	keyboard := planCardDeleteKeyboard(rep.addedPlan)
	cancelCallbackData := keyboard.InlineKeyboard[0][1].CallbackData

	handleOpenPlan(
		context.Background(),
		callbackBot,
		testCallbackQuery(userID, cancelCallbackData),
		rep,
	)

	if rep.deleteCalls != 0 {
		t.Errorf("DeletePlan() calls = %d, want 0 after cancellation", rep.deleteCalls)
	}
	if rep.addedPlan.ID != planID {
		t.Errorf("plan ID after cancellation = %d, want %d", rep.addedPlan.ID, planID)
	}
	if len(callbackBot.edits) != 1 || callbackBot.edits[0].Text != planCardText(rep.addedPlan) {
		t.Errorf("cancel did not return to the plan card: edits = %+v", callbackBot.edits)
	}
}

func TestConfirmPlanDeletion_RemovesPlanFromBothUsersLists(t *testing.T) {
	const (
		coupleID     int64 = 789
		firstUserID  int64 = 123
		secondUserID int64 = 321
		planID       int64 = 1
	)
	callbackBot := &fakeCallbackBot{}
	baseRep := &fakeRepository{
		addedPlan: plans.Plan{
			ID:       planID,
			NewPlan:  plans.NewPlan{Title: "Вечер кино"},
			CoupleID: coupleID,
		},
	}
	rep := &fakeBotRepository{
		fakeRepository: baseRep,
		couple: couples.Couple{
			ID:           coupleID,
			FirstUserID:  firstUserID,
			SecondUserID: secondUserID,
		},
	}

	handleConfirmPlanDeletion(
		context.Background(),
		callbackBot,
		testCallbackQuery(firstUserID, planDeleteConfirmPrefix+"1"),
		rep,
	)

	if baseRep.deleteCalls != 1 {
		t.Fatalf("DeletePlan() calls = %d, want 1", baseRep.deleteCalls)
	}
	if baseRep.addedPlan.ID != 0 {
		t.Fatalf("plan after confirmation = %+v, want deleted", baseRep.addedPlan)
	}
	if len(callbackBot.answers) != 1 || callbackBot.answers[0].Text != messageDeletePlanSuccess {
		t.Errorf("confirmation response = %+v, want deletion success message", callbackBot.answers)
	}

	for _, userID := range []int64{firstUserID, secondUserID} {
		t.Run(fmt.Sprintf("user_%d", userID), func(t *testing.T) {
			sender := &fakeMessageSender{}
			handler(
				context.Background(),
				sender,
				&models.Update{
					Message: &models.Message{
						From: &models.User{ID: userID},
						Chat: models.Chat{ID: userID},
						Text: CommandPlans,
					},
				},
				rep,
				NewConversationStore(),
			)

			if len(sender.messages) != 1 {
				t.Fatalf("messages = %+v, want one plans response", sender.messages)
			}
			if sender.messages[0].text != zeroPlansText {
				t.Errorf(
					"plans text = %q, want empty-list text %q",
					sender.messages[0].text,
					zeroPlansText,
				)
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
