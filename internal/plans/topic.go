package plans

type Topic int

const (
	Places Topic = iota + 1
	Movies
	AtHome
	Travel
	Other
)

func IsValidTopic(topic Topic) bool {
	switch topic {
	case Places, Movies, AtHome, Travel, Other:
		return true
	default:
		return false
	}
}
