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

type sentMessage struct {
	chatID any
	text   string
	markup models.ReplyMarkup
}

type fakeMessageSender struct {
	messages []sentMessage
}

func (fm *fakeMessageSender) SendMessage(
	ctx context.Context,
	params *tgBot.SendMessageParams,
) (*models.Message, error) {

	fm.messages = append(
		fm.messages, sentMessage{
			chatID: params.ChatID,
			text:   params.Text,
			markup: params.ReplyMarkup,
		},
	)
	return nil, nil
}

type fakeRepository struct {
	addedPlan plans.Plan
	allCalls  int
	addCalls  int
}

func (fr *fakeRepository) Add(newPlan plans.NewPlan, coupleID int64) (int64, error) {
	fr.addCalls++
	fr.addedPlan = plans.Plan{
		ID:       1,
		NewPlan:  newPlan,
		CoupleID: coupleID,
	}

	return fr.addedPlan.ID, nil
}

func (fr *fakeRepository) All(coupleID int64) ([]plans.Plan, error) {
	fr.allCalls++
	if fr.addedPlan.ID != 0 && fr.addedPlan.CoupleID == coupleID {
		return []plans.Plan{fr.addedPlan}, nil
	}

	return []plans.Plan{}, nil
}

func (fr *fakeRepository) GetPlanByID(planID int64, coupleID int64) (plans.Plan, error) {
	if fr.addedPlan.ID != planID || fr.addedPlan.CoupleID != coupleID {
		return plans.Plan{}, nil
	}

	return fr.addedPlan, nil
}

type fakeBotRepository struct {
	*fakeRepository
	couples.CoupleRepository
	couples.InvitationRepository
	couple    couples.Couple
	coupleErr error
}

func (fr *fakeBotRepository) GetCoupleByUserID(userID int64) (couples.Couple, error) {
	if fr.coupleErr != nil {
		return couples.Couple{}, fr.coupleErr
	}
	return fr.couple, nil
}

func (fr *fakeRepository) SetDone(planID int64, coupleID int64, done bool) error {
	if fr.addedPlan.ID == planID && fr.addedPlan.CoupleID == coupleID {
		fr.addedPlan.Done = done
	} else {
		return fmt.Errorf(
			"%w: ID = %v, coupleID = %v",
			plans.ErrPlanNotFound,
			planID,
			coupleID,
		)
	}

	return nil
}

func (fr *fakeRepository) DeletePlan(planID int64) error {
	if fr.addedPlan.ID == planID {
		fr.addedPlan = plans.Plan{}
	} else {
		return fmt.Errorf("%w: ID = %v", plans.ErrPlanNotFound, planID)
	}

	return nil
}

func TestHandlePlanConversation_Success(t *testing.T) {
	// Arrange: подготовка
	rep := &fakeRepository{}
	sender := &fakeMessageSender{}
	conversationStore := NewConversationStore()

	const (
		userID    int64 = 123
		chatID    int64 = 456
		partnerID int64 = 321
		coupleID  int64 = 789
	)

	userConv := Conversation{
		State: StateWaitingForNote,
		NewPlan: plans.NewPlan{
			Title: "Watch a movie",
			Note:  "",
			Topic: plans.Movies,
		},
		CoupleID:  coupleID,
		PartnerID: partnerID,
	}

	conversationStore.Set(
		userID,
		userConv,
	)

	// Act: вызываем проверяемую функцию
	handlePlanConversation(
		context.Background(),
		sender,
		rep,
		conversationStore,
		userConv,
		"On Friday",
		userID,
		chatID,
	)

	// Assert: проверяем сохранённый план
	got := rep.addedPlan

	if got.Title != "Watch a movie" {
		t.Errorf("Title = %q, want %q", got.Title, "Watch a movie")
	}

	if got.Note != "On Friday" {
		t.Errorf("Note = %q, want %q", got.Note, "On Friday")
	}

	if got.Topic != plans.Movies {
		t.Errorf("Topic = %v, want %v", got.Topic, plans.Movies)
	}
	if got.CoupleID != coupleID {
		t.Errorf("CoupleID = %d, want %d", got.CoupleID, coupleID)
	}

	// Assert: проверяем отправленное сообщение
	if len(sender.messages) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(sender.messages))
	}

	if sender.messages[0].chatID != chatID || sender.messages[0].text != messagePlanAdded {
		t.Errorf(
			"first message = %+v, want chatID %d and text %q",
			sender.messages[0],
			chatID,
			messagePlanAdded,
		)
	}
	wantPartnerMessage := "Твой партнёр добавил новый план:" + userConv.NewPlan.Title + ". 💕"
	if sender.messages[1].chatID != partnerID || sender.messages[1].text != wantPartnerMessage {
		t.Errorf(
			"second message = %+v, want chatID %d and text %q",
			sender.messages[1],
			partnerID,
			wantPartnerMessage,
		)
	}

	// Assert: проверяем состояние диалога
	userConv, ok := conversationStore.Get(userID)
	if !ok {
		t.Fatalf("no conversation with userID = %v", userID)
	}
	if userConv.State != StateIdle {
		t.Errorf("State = %v, want %v", userConv.State, StateIdle)
	}

}

func TestHandlerPlansRequiresCouple(t *testing.T) {
	baseRep := &fakeRepository{}
	rep := &fakeBotRepository{
		fakeRepository: baseRep,
		coupleErr:      couples.ErrCoupleNotFound,
	}
	sender := &fakeMessageSender{}
	conversations := NewConversationStore()

	const (
		userID int64 = 123
		chatID
	)

	handler(
		context.Background(),
		sender,
		&models.Update{
			Message: &models.Message{
				From: &models.User{ID: userID},
				Chat: models.Chat{ID: chatID},
				Text: CommandPlans,
			},
		},
		rep,
		conversations,
	)

	if baseRep.allCalls != 0 {
		t.Errorf("All() calls = %d, want 0 for a user without a couple", baseRep.allCalls)
	}
	if len(sender.messages) != 1 || sender.messages[0].chatID != chatID {
		t.Fatalf("messages = %+v, want one response to chat %d", sender.messages, chatID)
	}
}

func TestHandlerAddPlanRequiresCouple(t *testing.T) {
	baseRep := &fakeRepository{}
	rep := &fakeBotRepository{
		fakeRepository: baseRep,
		coupleErr:      couples.ErrCoupleNotFound,
	}
	sender := &fakeMessageSender{}
	conversations := NewConversationStore()

	const (
		userID int64 = 123
		chatID
	)

	handler(
		context.Background(),
		sender,
		&models.Update{
			Message: &models.Message{
				From: &models.User{ID: userID},
				Chat: models.Chat{ID: chatID},
				Text: CommandAddPlan,
			},
		},
		rep,
		conversations,
	)

	if baseRep.addCalls != 0 {
		t.Errorf("Add() calls = %d, want 0 for a user without a couple", baseRep.addCalls)
	}
	if _, ok := conversations.Get(userID); ok {
		t.Error("conversation was created for a user without a couple")
	}
	if len(sender.messages) != 1 || sender.messages[0].chatID != chatID {
		t.Fatalf("messages = %+v, want one response to chat %d", sender.messages, chatID)
	}
}

func TestHandlerPlansUsesCouplePlansForBothUsers(t *testing.T) {
	const (
		coupleID     int64 = 789
		firstUserID  int64 = 123
		secondUserID int64 = 321
	)

	for _, userID := range []int64{firstUserID, secondUserID} {
		t.Run(fmt.Sprintf("user_%d", userID), func(t *testing.T) {
			baseRep := &fakeRepository{
				addedPlan: plans.Plan{
					ID:       1,
					Done:     true,
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

			if baseRep.allCalls != 1 {
				t.Errorf("All() calls = %d, want 1", baseRep.allCalls)
			}
			if len(sender.messages) != 1 {
				t.Fatalf("messages = %+v, want one plans response", sender.messages)
			}
			wantText, _ := buildPlansMessage([]plans.Plan{baseRep.addedPlan})
			if sender.messages[0].text != wantText {
				t.Errorf("message text = %q, want shared couple plan list %q", sender.messages[0].text, wantText)
			}
			keyboard, ok := sender.messages[0].markup.(*models.InlineKeyboardMarkup)
			if !ok {
				t.Fatalf("keyboard type = %T, want *models.InlineKeyboardMarkup", sender.messages[0].markup)
			}
			if got := keyboard.InlineKeyboard[0][0].Text; got != "✅ Вечер кино" {
				t.Errorf("shared plan status = %q, want completed plan label", got)
			}
		})
	}
}

func TestHandlerAddPlanStartsForBothUsersInCouple(t *testing.T) {
	const (
		coupleID     int64 = 789
		firstUserID  int64 = 123
		secondUserID int64 = 321
	)

	for _, testCase := range []struct {
		userID    int64
		partnerID int64
	}{
		{userID: firstUserID, partnerID: secondUserID},
		{userID: secondUserID, partnerID: firstUserID},
	} {
		t.Run(fmt.Sprintf("user_%d", testCase.userID), func(t *testing.T) {
			rep := &fakeBotRepository{
				fakeRepository: &fakeRepository{},
				couple: couples.Couple{
					ID:           coupleID,
					FirstUserID:  firstUserID,
					SecondUserID: secondUserID,
				},
			}
			sender := &fakeMessageSender{}
			conversations := NewConversationStore()

			handler(
				context.Background(),
				sender,
				&models.Update{
					Message: &models.Message{
						From: &models.User{ID: testCase.userID},
						Chat: models.Chat{ID: testCase.userID},
						Text: CommandAddPlan,
					},
				},
				rep,
				conversations,
			)

			conversation, ok := conversations.Get(testCase.userID)
			if !ok {
				t.Fatal("conversation was not created for a couple member")
			}
			if conversation.State != StateWaitingForTitle {
				t.Errorf("State = %v, want %v", conversation.State, StateWaitingForTitle)
			}
			if conversation.CoupleID != coupleID {
				t.Errorf("CoupleID = %d, want %d", conversation.CoupleID, coupleID)
			}
			if conversation.PartnerID != testCase.partnerID {
				t.Errorf("PartnerID = %d, want %d", conversation.PartnerID, testCase.partnerID)
			}

			handleConversation(
				context.Background(),
				sender,
				rep,
				conversations,
				"Сходить в кино",
				testCase.userID,
				testCase.userID,
			)
			handleConversation(
				context.Background(),
				sender,
				rep,
				conversations,
				"-",
				testCase.userID,
				testCase.userID,
			)

			if rep.addedPlan.Title != "Сходить в кино" || rep.addedPlan.CoupleID != coupleID {
				t.Errorf(
					"added plan = %+v, want title and couple ID %q/%d",
					rep.addedPlan,
					"Сходить в кино",
					coupleID,
				)
			}
		})
	}
}
