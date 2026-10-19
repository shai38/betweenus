package bot

import (
	"context"

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

type messageSender interface {
	SendMessage(ctx context.Context, params *tgBot.SendMessageParams) (*models.Message, error)
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
