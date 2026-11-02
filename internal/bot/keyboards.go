package bot

import (
	"strconv"

	"github.com/go-telegram/bot/models"
	"github.com/shai38/betweenus/internal/plans"
)

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

func plansKeyboard(plansList []plans.Plan) *models.InlineKeyboardMarkup {
	keyboardRows := make([][]models.InlineKeyboardButton, len(plansList))
	for index, plan := range plansList {
		emoji := ""

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

func planCardKeyboard(plan plans.Plan) *models.InlineKeyboardMarkup {
	keyboardRows := make([][]models.InlineKeyboardButton, 2)

	var setDoneButton models.InlineKeyboardButton

	if plan.Done {
		setDoneButton = models.InlineKeyboardButton{
			Text:         "↩️ В активные",
			CallbackData: planSetActivePrefix + strconv.FormatInt(plan.ID, 10),
		}
	} else {
		setDoneButton = models.InlineKeyboardButton{
			Text:         "✅ Выполнить",
			CallbackData: planSetDonePrefix + strconv.FormatInt(plan.ID, 10),
		}
	}
	backButton := models.InlineKeyboardButton{
		Text:         "← К списку",
		CallbackData: backToPlansCallbackData,
	}
	deleteButton := models.InlineKeyboardButton{
		Text:         "🗑 Удалить",
		CallbackData: planDeleteRequestPrefix + strconv.FormatInt(plan.ID, 10),
	}
	keyboardRows[0] = []models.InlineKeyboardButton{setDoneButton, deleteButton}
	keyboardRows[1] = []models.InlineKeyboardButton{backButton}

	return &models.InlineKeyboardMarkup{InlineKeyboard: keyboardRows}
}

func planCardDeleteKeyboard(plan plans.Plan) *models.InlineKeyboardMarkup {
	keyboardRows := make([][]models.InlineKeyboardButton, 1)

	deleteButton := models.InlineKeyboardButton{
		Text:         "Удалить",
		CallbackData: planDeleteConfirmPrefix + strconv.FormatInt(plan.ID, 10),
	}
	cancelButton := models.InlineKeyboardButton{
		Text:         "Отмена",
		CallbackData: planOpenPrefix + strconv.FormatInt(plan.ID, 10),
	}
	keyboardRows[0] = []models.InlineKeyboardButton{deleteButton, cancelButton}

	return &models.InlineKeyboardMarkup{InlineKeyboard: keyboardRows}
}
