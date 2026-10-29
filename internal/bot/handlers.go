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

const (
	planOpenPrefix      = "plan:open:"
	planSetActivePrefix = "plan:active:"
	planSetDonePrefix   = "plan:done:"
	backToPlans         = "back_to_plans"
)

func handler(
	ctx context.Context,
	b messageSender,
	update *models.Update,
	rep BotRepository,
	conversations *ConversationStore,
) {
	if update.Message == nil {
		return
	}
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	var sendText string

	switch userText := update.Message.Text; userText {
	case CommandStart:
		handleStart(ctx, b, chatID, userID, rep)

	case CommandHelp:
		sendText = helpIntro + "\n\n" + formatHelp()
		sendMessage(ctx, b, chatID, sendText, nil)

	case CommandPlans:
		handlePlans(ctx, b, chatID, userID, rep)

	case CommandAddPlan:
		couple, err := rep.GetCoupleByUserID(userID)
		if err != nil {
			if errors.Is(err, couples.ErrCoupleNotFound) {
				sendText = "План можно будет добавить, когда ты создашь пару 💕 Нажми /start," +
					" чтобы отправить приглашение партнёру или ввести его код."
			} else {
				sendText = messageAddPlanLoadError
			}
			sendMessage(ctx, b, chatID, sendText, nil)
			return
		}

		partnerID, err := couple.PartnerID(userID)
		if err != nil {
			log.Printf("failed to get partner id: %v", err)

			sendMessage(ctx, b, chatID, messageAddPlanLoadError, nil)
			return
		}

		userConv := Conversation{
			State: StateWaitingForTitle,
			NewPlan: plans.NewPlan{
				Topic: plans.Other,
			},
			CoupleID:  couple.ID,
			PartnerID: partnerID,
		}
		conversations.Set(userID, userConv)
		sendText = "Как назовём ваш план? 💕"
		sendMessage(ctx, b, chatID, sendText, nil)

	default:
		handleConversation(ctx, b, rep, conversations, userText, userID, chatID)
	}
}

func handleStart(
	ctx context.Context,
	b messageSender,
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
			sendText = "Возникли временные трудности. Попробуй чуть позже. 🙏"
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
			sendText = "Один из вас уже состоит в паре. 💕"
		} else if errors.Is(err, couples.ErrInvitationNotFound) {
			sendText = "Не удалось принять приглашение. Проверь код и попробуй ещё раз. 🙏"
		} else {
			log.Printf("failed to accept invitation: %v", err)
			sendText = "Возникла техническая ошибка. Попробуй позже. 🙏"
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
	var sendText string

	if !readyToSave {
		conversations.Set(userID, userConv)
		sendText = "Если нужно добавить заметку, введите текст заметки следующим сообщением, " +
			"если нет, то введите -."
		sendMessage(ctx, b, chatID, sendText, nil)
		return
	}

	_, err := rep.Add(userConv.NewPlan, userConv.CoupleID)
	if err != nil {
		log.Printf("failed to add plan: %v", err)

		sendMessage(ctx, b, chatID, messageAddPlanLoadError, nil)
		return
	}
	sendMessage(ctx, b, chatID, messagePlanAdded, nil)

	// Отправление уведомления партнёру о создании плана
	sendText = "Твой партнёр добавил новый план:" + userConv.NewPlan.Title + ". 💕"
	sendMessage(ctx, b, userConv.PartnerID, sendText, nil)

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

func handlePlans(
	ctx context.Context,
	b messageSender,
	chatID int64,
	userID int64,
	rep BotRepository,
) {
	var sendText string

	couple, err := rep.GetCoupleByUserID(userID)
	if err != nil {
		if errors.Is(err, couples.ErrCoupleNotFound) {
			sendText = "Список планов появится, когда ты создашь пару 💕 Нажми /start," +
				" чтобы отправить приглашение партнёру или ввести его код."
		} else {
			sendText = messagePlansLoadError
		}
		sendMessage(ctx, b, chatID, sendText, nil)
		return
	}
	allPlans, err := rep.All(couple.ID)
	if err != nil {
		log.Printf("failed to get plans: %v", err)

		sendMessage(ctx, b, chatID, messagePlansLoadError, nil)
		return
	}

	var pk *models.InlineKeyboardMarkup
	if len(allPlans) == 0 {
		sendText = ZeroPlansText
		pk = nil
	} else {
		sendText = "💞 Наши планы\n\nНажми на план, чтобы посмотреть заметку и действия.\n"
		pk = plansKeyboard(allPlans)
	}

	sendMessage(ctx, b, chatID, sendText, pk)
}

func plansKeyboard(plansList []plans.Plan) *models.InlineKeyboardMarkup {
	keyboardRows := make([][]models.InlineKeyboardButton, len(plansList))
	for index, plan := range plansList {
		emoji := "⬜"

		if plan.Done {
			emoji = "✅"
		}

		button := models.InlineKeyboardButton{
			Text:         emoji + " " + plan.Title,
			CallbackData: "plan:open:" + strconv.FormatInt(plan.ID, 10),
		}

		keyboardRows[index] = []models.InlineKeyboardButton{button}
	}

	return &models.InlineKeyboardMarkup{InlineKeyboard: keyboardRows}
}

func planCardText(plan plans.Plan) string {
	var emoji string

	if plan.Done {
		emoji = "✅"
	}

	result := plan.Title + " " + emoji + "\n\n"
	if plan.Note != "" {
		result += "Заметка:\n" + plan.Note + "\n"
	}

	return result
}

func planCardKeyboard(plan plans.Plan) *models.InlineKeyboardMarkup {
	keyboardRows := make([][]models.InlineKeyboardButton, 1)

	var button models.InlineKeyboardButton

	if plan.Done {
		button = models.InlineKeyboardButton{
			Text:         "↩️ Вернуть в активные",
			CallbackData: planSetActivePrefix + strconv.FormatInt(plan.ID, 10),
		}
	} else {
		button = models.InlineKeyboardButton{
			Text:         "✅ Выполнить",
			CallbackData: planSetDonePrefix + strconv.FormatInt(plan.ID, 10),
		}
	}
	backbutton := models.InlineKeyboardButton{
		Text:         "← К списку",
		CallbackData: backToPlans,
	}
	keyboardRows[0] = []models.InlineKeyboardButton{button, backbutton}

	return &models.InlineKeyboardMarkup{InlineKeyboard: keyboardRows}
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
