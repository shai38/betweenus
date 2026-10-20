package bot

import (
	"context"
	"log"
	"strconv"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/shai38/betweenus/internal/plans"
)

func handler(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
	rep BotRepository,
	conversations *ConversationStore,
) {
	if update.Message == nil {
		return
	}
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	switch userText := update.Message.Text; userText {
	case CommandStart:
		sendText := "Привет! Это BetweenUs 💕\nМаленькие шаги к тому, чтобы стать ближе."
		sendMessage(ctx, b, chatID, sendText)

	case CommandHelp:
		sendText := "Я BetweenUs 💕\n\n" + formatHelp()
		sendMessage(ctx, b, chatID, sendText)

	case CommandPlans:
		allPlans, err := rep.All()
		if err != nil {
			log.Printf("failed to get plans: %v", err)

			sendErr := "Не удалось загрузить планы. Попробуй чуть позже 🙏"
			sendMessage(ctx, b, chatID, sendErr)

			return
		}

		sendMessage(ctx, b, chatID, formatPlans(allPlans))

	case CommandAddPlan:
		userConv := Conversation{
			State: StateWaitingForTitle,
			NewPlan: plans.NewPlan{
				Topic: plans.Other,
			},
		}
		conversations.Set(userID, userConv)
		sendText := "Как назовём ваш план? 💕"
		sendMessage(ctx, b, chatID, sendText)

	default:
		handleConversation(ctx, b, rep, conversations, userText, userID, chatID)

	}
}

func handleConversation(
	ctx context.Context,
	b messageSender,
	rep plans.Repository,
	conversations *ConversationStore,
	userText string,
	userID int64,
	chatID int64,
) {
	userConv, ok := conversations.Get(userID)
	if !ok {
		return
	}
	userConv, readyToSave := processConversationInput(userConv, userText)

	if !readyToSave {
		conversations.Set(userID, userConv)
		sendText := "Если нужно добавить заметку, введите текст заметки следующим сообщением, если нет, то введите -"
		sendMessage(ctx, b, chatID, sendText)
		return
	}

	_, err := rep.Add(userConv.NewPlan)
	if err != nil {
		log.Printf("failed to add plan: %v", err)

		sendErr := "Не удалось добавить план. Попробуй чуть позже 🙏"
		sendMessage(ctx, b, chatID, sendErr)
		return
	}
	sendMessage(ctx, b, chatID, messagePlanAdded)

	userConv.Clear()
	conversations.Set(userID, userConv)
}

func formatPlans(plansList []plans.Plan) string {
	if len(plansList) == 0 {
		return ZeroPlansText
	}

	text := "💞 Наши планы\n\n"
	counter := 1
	for _, plan := range plansList {
		text += strconv.Itoa(counter) + ". " + plan.Title + "\n"
		if plan.Note != "" {
			text += plan.Note + "\n"
		}
		text += "\n"
		counter++
	}

	return text
}

func sendMessage(ctx context.Context, b messageSender, chatID int64, text string) {
	_, err := b.SendMessage(ctx, &tgBot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		log.Printf("failed to send message: %v", err)
	}
}
