package plans

type ID int64

type Plan struct {
	ID   ID
	Done bool
	NewPlan
}

type NewPlan struct {
	Title string
	Note  string
	Topic Topic
}
