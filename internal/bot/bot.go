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
	TgBot *tgBot.Bot
	Rep   plans.Repository
}

func (b *Bot) Start(ctx context.Context) {
	b.TgBot.Start(ctx)
}

func New(token string, rep plans.Repository) (*Bot, error) {
	TgBot, err := tgBot.New(
		token,
		tgBot.WithDefaultHandler(
			func(ctx context.Context, b *tgBot.Bot, update *models.Update) {
				handler(ctx, b, update, rep)
			},
		),
	)
	if err != nil {
		return nil, err
	}

	betweenUsBot := &Bot{TgBot: TgBot, Rep: rep}

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

func handler(ctx context.Context, b *tgBot.Bot, update *models.Update, rep plans.Repository) {
	if update.Message == nil {
		return
	}

	switch update.Message.Text {
	case "/start":
		sendText := "Привет! Это BetweenUs 💕\nМаленькие шаги к тому, чтобы стать ближе."
		sendMessage(ctx, b, update.Message.Chat.ID, sendText)

	case "/help":
		sendText := "Я BetweenUs 💕\n\n/start — начать\n/plans — список планов\n/help — помощь"
		sendMessage(ctx, b, update.Message.Chat.ID, sendText)

	case "/plans":
		allPlans, err := rep.All()
		if err != nil {
			log.Printf("failed to get plans: %v", err)

			sendErr := "Не удалось загрузить планы. Попробуй чуть позже 🙏"
			sendMessage(ctx, b, update.Message.Chat.ID, sendErr)

			return
		}
		if len(allPlans) == 0 {
			sendText := "Пока планов нет 💭"
			sendMessage(ctx, b, update.Message.Chat.ID, sendText)

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

			sendMessage(ctx, b, update.Message.Chat.ID, text)
		}
	}
}
