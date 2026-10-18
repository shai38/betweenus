package bot

import (
	"testing"

	"github.com/shai38/betweenus/internal/plans"
)

func checkConversationState(
	t *testing.T,
	store *ConversationStore,
	userID int64,
	want ConversationState,
) {
	t.Helper()

	got, ok := store.Get(userID)
	if !ok {
		t.Fatalf("Get(%d) found no conversation", userID)
	}
	if got.State != want {
		t.Fatalf("Get(%d).State = %v, want %v", userID, got.State, want)
	}
}

func TestConversationStore(t *testing.T) {
	firstUserID := int64(101)

	convStore := NewConversationStore()

	firstConv, ok := convStore.Get(firstUserID)
	if ok {
		t.Fatalf("new conversation store has state: %v", firstConv.State)
	}
	firstState := StateWaitingForTitle

	convStore.Set(
		firstUserID,
		Conversation{State: firstState, NewPlan: plans.NewPlan{}},
	)

	checkConversationState(t, convStore, firstUserID, firstState)

	secondUserID := int64(202)
	secondState := StateWaitingForNote

	convStore.Set(
		secondUserID,
		Conversation{State: secondState, NewPlan: plans.NewPlan{}},
	)

	checkConversationState(t, convStore, firstUserID, firstState)
	checkConversationState(t, convStore, secondUserID, secondState)
}

func TestClear(t *testing.T) {
	conversation := Conversation{
		State:   StateWaitingForTitle,
		NewPlan: plans.NewPlan{Title: "Watch SpiderMan", Note: "In 3D", Topic: plans.Movies},
	}
	conversation.Clear()

	if conversation.State != StateIdle {
		t.Errorf("state was not set to default after clear method, got %v, want %v", conversation.State, StateIdle)
	}
	if conversation.NewPlan.Title != "" {
		t.Errorf("title of the new plan was not deleted after clear method, got %v", conversation.NewPlan.Title)
	}
	if conversation.NewPlan.Note != "" {
		t.Errorf("note of the new plan was not deleted after clear method, got %v", conversation.NewPlan.Note)
	}
	if conversation.NewPlan.Topic != plans.Empty {
		t.Errorf("topic of the new plan was not deleted after clear method, got %v", conversation.NewPlan.Topic)
	}
}
