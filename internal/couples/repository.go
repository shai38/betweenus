package couples

type InvitationRepository interface {
	CreateInvitation(Invitation) error
	GetInvitationByCode(string) (Invitation, error)
	DeleteInvitation(string) error
}

type CoupleRepository interface {
	CreateCouple(int64, int64) (int64, error)
	GetCoupleByID(int64) (Couple, error)
}
