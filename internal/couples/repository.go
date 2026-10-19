package couples

type InvitationRepository interface {
	CreateInvitation(Invitation) error
	GetInvitationByCode(string) (Invitation, error)
	DeleteInvitation(string) error
}
