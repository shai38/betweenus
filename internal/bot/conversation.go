package bot

func processConversationInput(
	conversation Conversation,
	userText string,
) (Conversation, bool) {
	var readyToSave bool

	switch conversation.State {
	case StateWaitingForTitle:
		conversation.NewPlan.Title = userText
		conversation.State = StateWaitingForNote
	case StateWaitingForNote:
		if userText != "-" {
			conversation.NewPlan.Note = userText
		}
		readyToSave = true
	}

	return conversation, readyToSave
}
