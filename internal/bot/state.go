package bot

import "github.com/shai38/betweenus/internal/plans"

type ConversationState int

const (
	StateIdle ConversationState = iota
	StateWaitingForTitle
	StateWaitingForNote
)

type Conversation struct {
	State   ConversationState
	NewPlan plans.NewPlan
}

type ConversationStore struct {
	conversations map[int64]Conversation
}

func NewConversationStore() *ConversationStore {
	return &ConversationStore{conversations: make(map[int64]Conversation)}
}

func (s *ConversationStore) Set(userID int64, conversation Conversation) {
	s.conversations[userID] = conversation
}

func (s *ConversationStore) Get(userID int64) (Conversation, bool) {
	conversation, ok := s.conversations[userID]
	return conversation, ok
}
