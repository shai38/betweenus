package bot

import (
	"github.com/shai38/betweenus/internal/couples"
	"github.com/shai38/betweenus/internal/plans"
)

type BotRepository interface {
	plans.Repository
	couples.CoupleRepository
}
