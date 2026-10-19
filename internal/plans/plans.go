package plans

type Plan struct {
	ID   int64
	Done bool
	NewPlan
}

type NewPlan struct {
	Title string
	Note  string
	Topic Topic
}
