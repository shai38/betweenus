package plans

type Repository interface {
	Add(plan NewPlan, coupleID int64) (int64, error)
	All(coupleID int64) ([]Plan, error)
	MarkDone(planID int64) error
	DeletePlan(planID int64) error
}
