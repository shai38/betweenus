package plans

type Plan struct {
	ID   int64
	Done bool
	NewPlan
	CoupleID int64
}

type NewPlan struct {
	Title string
	Note  string
	Topic Topic
}
