package plans

type Repository interface {
	Add(plan NewPlan) (ID, error)
	All() ([]Plan, error)
	MarkDone(planID ID) error
	Delete(planID ID) error
}
