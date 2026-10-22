package bot

import (
	"context"
	"errors"
	"log"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/shai38/betweenus/internal/couples"
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
func (b *Bot) SetMyCommands(ctx context.Context, params *tgBot.SetMyCommandsParams) (bool, error) {
	return b.TgBot.SetMyCommands(ctx, params)
}

type messageSender interface {
	SendMessage(ctx context.Context, params *tgBot.SendMessageParams) (*models.Message, error)
}

func New(token string, rep BotRepository) (*Bot, error) {
	conversationStore := NewConversationStore()
	invService := couples.NewInvitationService(rep)

	TgBot, err := tgBot.New(
		token,
		tgBot.WithDefaultHandler(
			func(ctx context.Context, b *tgBot.Bot, update *models.Update) {
				handler(ctx, b, update, rep, conversationStore)
			},
		),
		tgBot.WithCallbackQueryDataHandler(
			"create_invitation",
			tgBot.MatchTypeExact,
			func(ctx context.Context, b *tgBot.Bot, update *models.Update) {
				answerCallbackQuery(ctx, b, update.CallbackQuery)
				var sendText string

				inv, err := invService.CreateInvitation(update.CallbackQuery.From.ID)
				if err != nil {
					if errors.Is(err, couples.ErrOnlyOneCouplePerUser) {
						sendText = "Твоя пара уже найдена."
					} else {
						log.Printf("failed to create invitation: %v", err)
						sendText = "Не удалось сформировать приглашение. Попробуй чуть позже 🙏"
					}
				} else {
					sendText = "Твой код приглашения: " + inv.Code + "\nОтправь его своему партнёру 💌"
				}
				sendMessage(ctx, b, update.CallbackQuery.Message.Message.Chat.ID, sendText, nil)
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

func answerCallbackQuery(ctx context.Context, b *tgBot.Bot, CallbackQuery *models.CallbackQuery) {
	if CallbackQuery != nil {
		_, err := b.AnswerCallbackQuery(
			ctx,
			&tgBot.AnswerCallbackQueryParams{
				CallbackQueryID: CallbackQuery.ID,
			},
		)
		if err != nil {
			log.Printf("failed to answer callback query: %v", err)
		}
	}
}
