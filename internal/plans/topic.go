package plans

type Topic int

const (
	Empty Topic = iota
	Places
	Movies
	AtHome
	Travel
	Other
)

func IsValidTopic(topic Topic) bool {
	switch topic {
	case Empty, Places, Movies, AtHome, Travel, Other:
		return true
	default:
		return false
	}
}
