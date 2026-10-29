package plans

type Repository interface {
	Add(plan NewPlan, coupleID int64) (int64, error)
	All(coupleID int64) ([]Plan, error)
	SetDone(planID int64, coupleID int64, done bool) error
	DeletePlan(planID int64) error
}
