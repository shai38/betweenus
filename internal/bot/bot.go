package bot

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func New(token string) (*bot.Bot, error) {
	return bot.New(token, bot.WithDefaultHandler(handler))
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	switch update.Message.Text {
	case "/start":
		sendText := "Привет! Это BetweenUs 💕\nМаленькие шаги к тому, чтобы стать ближе."

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   sendText,
		})
		if err != nil {
			log.Printf("failed to send message: %v", err)
		}

	case "/help":
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Я BetweenUs 💕\n\n/start — начать\n/help — помощь",
		})
		if err != nil {
			log.Printf("failed to send message: %v", err)
		}
	}
}
