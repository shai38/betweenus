package bot

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
	text := "Доступные команды:\n"

	for _, command := range commands {
		text += command.Name + " — " + command.Description + "\n"
	}

	return text
}
