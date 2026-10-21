package couples

type InvitationServiceRepository interface {
	InvitationRepository
	CoupleRepository
}

type InvitationRepository interface {
	CreateInvitation(Invitation) error
	GetInvitationByCode(code string) (Invitation, error)
	GetInvitationByCreatorUserID(userID int64) (Invitation, error)
	DeleteInvitation(code string) error
}

type CoupleRepository interface {
	CreateCouple(firstUserID int64, secondUserID int64) (int64, error)
	GetCoupleByID(coupleID int64) (Couple, error)
	GetCoupleByUserID(userID int64) (Couple, error)
}
