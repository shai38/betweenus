package couples

type InvitationRepository interface {
	CreateInvitation(Invitation) error
	GetInvitationByCode(string) (Invitation, error)
	DeleteInvitation(string) error
}

type CoupleRepository interface {
	CreateCouple(firstUserID int64, secondUserID int64) (int64, error)
	GetCoupleByID(coupleID int64) (Couple, error)
	GetCoupleByUserID(userID int64) (Couple, error)
}
