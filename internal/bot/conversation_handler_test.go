package bot

import (
	"context"
	"fmt"

	"testing"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/shai38/betweenus/internal/plans"
)

type fakeMessageSender struct {
	messages []string
}

func (fm *fakeMessageSender) SendMessage(
	ctx context.Context,
	params *tgBot.SendMessageParams,
) (*models.Message, error) {

	fm.messages = append(fm.messages, params.Text)
	return nil, nil
}

type fakeRepository struct {
	addedPlan plans.Plan
}

func (fr *fakeRepository) Add(newPlan plans.NewPlan) (int64, error) {
	fr.addedPlan = plans.Plan{ID: 1, NewPlan: newPlan}

	return fr.addedPlan.ID, nil
}

func (fr *fakeRepository) All() ([]plans.Plan, error) {
	if fr.addedPlan.ID != 0 {
		return []plans.Plan{fr.addedPlan}, nil
	}

	return []plans.Plan{}, nil
}

func (fr *fakeRepository) MarkDone(planID int64) error {
	if fr.addedPlan.ID == planID {
		fr.addedPlan.Done = true
	} else {
		return fmt.Errorf("%w: ID = %v", plans.ErrPlanNotFound, planID)
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

func TestHandleConversation_Success(t *testing.T) {
	// Arrange: подготовка
	rep := &fakeRepository{}
	sender := &fakeMessageSender{}
	conversationStore := NewConversationStore()

	const userID int64 = 123
	const chatID int64 = 456

	conversationStore.Set(
		userID,
		Conversation{
			State: StateWaitingForNote,
			NewPlan: plans.NewPlan{
				Title: "Watch a movie",
				Note:  "",
				Topic: plans.Movies,
			},
		},
	)

	// Act: вызываем проверяемую функцию
	handleConversation(
		context.Background(),
		sender,
		rep,
		conversationStore,
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

	// Assert: проверяем отправленное сообщение
	if len(sender.messages) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(sender.messages))
	}

	wantMessage := messagePlanAdded
	if sender.messages[0] != wantMessage {
		t.Errorf("message = %q, want %q", sender.messages[0], wantMessage)
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
