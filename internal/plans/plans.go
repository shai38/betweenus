package plans

type Plan struct {
	ID    int64
	Title string
	Note  string
	Topic Topic
	Done  bool
}
