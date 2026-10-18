package bot

import (
	"testing"

	"github.com/shai38/betweenus/internal/plans"
)

func TestProcessConversationInput(t *testing.T) {
	conversation := Conversation{
		State:   StateWaitingForTitle,
		NewPlan: plans.NewPlan{},
	}
	textTitle := "Watch drama movie"

	conversation, _ = processConversationInput(conversation, textTitle)
	if conversation.State != StateWaitingForNote {
		t.Fatalf("conversation didn't change state, got %v, want %v", conversation.State, StateWaitingForNote)
	}

	textNote := "Description"
	conversation, readyToSave := processConversationInput(conversation, textNote)
	if !readyToSave {
		t.Fatal("readyToSave is false after filling note")
	}
	if conversation.NewPlan.Note != textNote {
		t.Fatalf("note was not filled, got %v, want %v", conversation.NewPlan.Note, textNote)
	}

	conversation, readyToSave = processConversationInput(conversation, "-")
	if !readyToSave {
		t.Fatal("readyToSave is false after filling note")
	}
	if conversation.NewPlan.Note == "-" && conversation.NewPlan.Note != textNote {
		t.Fatalf("note was filled: %v", textNote)
	}

}
