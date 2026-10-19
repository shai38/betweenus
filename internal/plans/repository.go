package plans

type Repository interface {
	Add(plan NewPlan) (int64, error)
	All() ([]Plan, error)
	MarkDone(planID int64) error
	Delete(planID int64) error
}
