package bot

import (
	"context"
	"log"
	"strconv"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/shai38/betweenus/internal/plans"
)

type Bot struct {
	TgBot         *tgBot.Bot
	Rep           plans.Repository
	Conversations *ConversationStore
}

func (b *Bot) Start(ctx context.Context) {
	b.TgBot.Start(ctx)
}

func New(token string, rep plans.Repository) (*Bot, error) {
	conversationStore := NewConversationStore()

	TgBot, err := tgBot.New(
		token,
		tgBot.WithDefaultHandler(
			func(ctx context.Context, b *tgBot.Bot, update *models.Update) {
				handler(ctx, b, update, rep, conversationStore)
			},
		),
	)
	if err != nil {
		return nil, err
	}

	betweenUsBot := &Bot{
		TgBot:         TgBot,
		Rep:           rep,
		Conversations: conversationStore,
	}

	return betweenUsBot, nil
}

func sendMessage(ctx context.Context, b *tgBot.Bot, chatID int64, text string) {
	_, err := b.SendMessage(ctx, &tgBot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		log.Printf("failed to send message: %v", err)
	}
}

func handler(
	ctx context.Context,
	b *tgBot.Bot,
	update *models.Update,
	rep plans.Repository,
	conversations *ConversationStore,
) {
	if update.Message == nil {
		return
	}
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	switch userText := update.Message.Text; userText {
	case "/start":
		sendText := "Привет! Это BetweenUs 💕\nМаленькие шаги к тому, чтобы стать ближе."
		sendMessage(ctx, b, chatID, sendText)

	case "/help":
		sendText := "Я BetweenUs 💕\n\n/start — начать\n /addplan — добавить план \n/plans — список планов\n/help — помощь"
		sendMessage(ctx, b, chatID, sendText)

	case "/plans":
		allPlans, err := rep.All()
		if err != nil {
			log.Printf("failed to get plans: %v", err)

			sendErr := "Не удалось загрузить планы. Попробуй чуть позже 🙏"
			sendMessage(ctx, b, chatID, sendErr)

			return
		}
		if len(allPlans) == 0 {
			sendText := "Пока планов нет 💭"
			sendMessage(ctx, b, chatID, sendText)

		} else {
			text := "💞 Наши планы\n\n"
			counter := 1
			for _, plan := range allPlans {
				text += strconv.Itoa(counter) + ". " + plan.Title + "\n"
				if plan.Note != "" {
					text += plan.Note + "\n"
				}
				text += "\n"
				counter++
			}

			sendMessage(ctx, b, chatID, text)
		}

	case "/addplan":
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
	b *tgBot.Bot,
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

	switch userConv.State {
	case StateWaitingForTitle:
		userConv.NewPlan.Title = userText
		userConv.State = StateWaitingForNote
		conversations.Set(userID, userConv)

		sendText := "Если нужно добавить заметку, введите текст заметки следующим сообщением, если нет, то введите -"
		sendMessage(ctx, b, chatID, sendText)

	case StateWaitingForNote:
		if userText != "-" {
			userConv.NewPlan.Note = userText
		}
		_, err := rep.Add(userConv.NewPlan)
		if err != nil {
			log.Printf("failed to add plan: %v", err)

			sendErr := "Не удалось добавить план. Попробуй чуть позже 🙏"
			sendMessage(ctx, b, chatID, sendErr)
			return
		}
		sendText := "План успешно добавлен!"
		sendMessage(ctx, b, chatID, sendText)

		userConv.State = StateIdle
		userConv.NewPlan = plans.NewPlan{}
		conversations.Set(userID, userConv)
	}
}
