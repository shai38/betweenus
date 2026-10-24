package bot

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/shai38/betweenus/internal/couples"
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
		handleStart(ctx, b, chatID, userID, rep)

	case CommandHelp:
		sendText := helpIntro + "\n\n" + formatHelp()

		sendMessage(ctx, b, chatID, sendText, nil)
	case CommandPlans:
		allPlans, err := rep.All()
		if err != nil {
			log.Printf("failed to get plans: %v", err)

			sendErr := "Не удалось загрузить планы. Попробуй чуть позже 🙏"
			sendMessage(ctx, b, chatID, sendErr, nil)

			return
		}

		sendMessage(ctx, b, chatID, formatPlans(allPlans), nil)

	case CommandAddPlan:
		userConv := Conversation{
			State: StateWaitingForTitle,
			NewPlan: plans.NewPlan{
				Topic: plans.Other,
			},
		}
		conversations.Set(userID, userConv)
		sendText := "Как назовём ваш план? 💕"
		sendMessage(ctx, b, chatID, sendText, nil)

	default:
		handleConversation(ctx, b, rep, conversations, userText, userID, chatID)
	}
}

func handleStart(
	ctx context.Context,
	b *tgBot.Bot,
	chatID int64,
	userID int64,
	rep BotRepository,
) {
	var sendText string
	var keyboard models.ReplyMarkup

	_, err := rep.GetCoupleByUserID(userID)

	if err == nil {
		sendText = startReturningUserText
	} else {
		if errors.Is(err, couples.ErrCoupleNotFound) {
			sendText = startNewUserText
			keyboard = startKeyboard()
		} else {
			log.Printf("failed to get couple by user: %v", err)
			sendText = "Возникли временные трудности. Попробуй чуть позже 🙏 "
		}
	}

	sendMessage(ctx, b, chatID, sendText, keyboard)
}

func handleConversation(
	ctx context.Context,
	b messageSender,
	rep BotRepository,
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
	case StateWaitingForTitle, StateWaitingForNote:
		handlePlanConversation(
			ctx,
			b,
			rep,
			conversations,
			userConv,
			userText,
			userID,
			chatID,
		)

	case StateWaitingForInvitationCode:
		handleInvitationCode(
			ctx,
			b,
			rep,
			conversations,
			userText,
			userID,
			chatID,
		)
	}
}

func handleInvitationCode(
	ctx context.Context,
	b messageSender,
	rep BotRepository,
	conversations *ConversationStore,
	userText string,
	userID int64,
	chatID int64,
) {
	code := strings.TrimSpace(userText)
	if code == "" {
		sendMessage(ctx, b, chatID, "Отправь код приглашения 💌", nil)
		return
	}

	invService := couples.NewInvitationService(rep)

	coupleID, err := invService.AcceptInvitation(code, userID)
	if err != nil {
		var sendText string
		if errors.Is(err, couples.ErrOnlyOneCouplePerUser) {
			sendText = "Один из вас уже состоит в паре 💕"
		} else if errors.Is(err, couples.ErrInvitationNotFound) {
			sendText = "Не удалось принять приглашение. Проверь код и попробуй ещё раз 🙏"
		} else {
			log.Printf("failed to accept invitation: %v", err)
			sendText = "Возникла техническая ошибка. Попробуй позже 🙏"
		}

		sendMessage(ctx, b, chatID, sendText, nil)

		return
	}

	conversation := Conversation{}
	conversations.Set(userID, conversation)

	notifyCoupleCreated(ctx, b, rep, coupleID)
}

// notifyCoupleCreated отправляет уведомление о создании пары обоим её участникам
func notifyCoupleCreated(
	ctx context.Context,
	b messageSender,
	rep couples.CoupleRepository,
	coupleID int64,
) {
	couple, err := rep.GetCoupleByID(coupleID)

	if err != nil {
		log.Printf("failed to notify users about successful couple creation: %v", err)
		return
	}

	sendMessage(ctx, b, couple.FirstUserID, coupleCreatedNotification, nil)
	sendMessage(ctx, b, couple.SecondUserID, coupleCreatedNotification, nil)
}

func handlePlanConversation(
	ctx context.Context,
	b messageSender,
	rep plans.Repository,
	conversations *ConversationStore,
	userConv Conversation,
	userText string,
	userID int64,
	chatID int64,
) {
	userConv, readyToSave := processConversationPlanInput(userConv, userText)

	if !readyToSave {
		conversations.Set(userID, userConv)
		sendText := "Если нужно добавить заметку, введите текст заметки следующим сообщением, " +
			"если нет, то введите -"
		sendMessage(ctx, b, chatID, sendText, nil)
		return
	}

	_, err := rep.Add(userConv.NewPlan)
	if err != nil {
		log.Printf("failed to add plan: %v", err)

		sendErr := "Не удалось добавить план. Попробуй чуть позже 🙏"
		sendMessage(ctx, b, chatID, sendErr, nil)
		return
	}
	sendMessage(ctx, b, chatID, messagePlanAdded, nil)

	userConv.Clear()
	conversations.Set(userID, userConv)
}

// startKeyboard создаёт клавиатуру
func startKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         "Создать приглашение",
					CallbackData: "create_invitation",
				},
			},
			{
				{
					Text:         "У меня есть код",
					CallbackData: "enter_invitation_code",
				},
			},
		},
	}
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

func sendMessage(
	ctx context.Context,
	b messageSender,
	chatID int64,
	text string,
	keyboard models.ReplyMarkup,
) {
	_, err := b.SendMessage(
		ctx,
		&tgBot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ReplyMarkup: keyboard,
		},
	)
	if err != nil {
		log.Printf("failed to send message: %v", err)
	}
}

func sendMarkdownMessage(
	ctx context.Context,
	b messageSender,
	chatID int64,
	text string,
	keyboard models.ReplyMarkup,
) {
	_, err := b.SendMessage(
		ctx,
		&tgBot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ParseMode:   models.ParseModeMarkdownV1,
			ReplyMarkup: keyboard,
		},
	)
	if err != nil {
		log.Printf("failed to send message: %v", err)
	}
}
