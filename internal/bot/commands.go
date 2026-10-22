package bot

import (
	"context"
	"strings"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type CommandInfo struct {
	Name        string
	Description string
}

const (
	CommandStart   = "/start"
	CommandHelp    = "/help"
	CommandPlans   = "/plans"
	CommandAddPlan = "/addplan"
)

var commands = []CommandInfo{
	{
		Name:        CommandStart,
		Description: "начать",
	},
	{
		Name:        CommandPlans,
		Description: "список планов",
	},
	{
		Name:        CommandAddPlan,
		Description: "добавить план",
	},
	{
		Name:        CommandHelp,
		Description: "помощь",
	},
}

func formatHelp() string {
	var text string

	for _, command := range commands {
		text += command.Name + " — " + command.Description + "\n"
	}

	return strings.TrimSpace(text)
}

func SetBotCommands(ctx context.Context, b *Bot) error {
	botCommands := make([]models.BotCommand, 0, len(commands))

	for _, command := range commands {
		botCommands = append(
			botCommands,
			models.BotCommand{
				Command:     strings.TrimPrefix(command.Name, "/"),
				Description: command.Description,
			},
		)
	}

	_, err := b.SetMyCommands(
		ctx,
		&tgBot.SetMyCommandsParams{Commands: botCommands},
	)

	return err
}
