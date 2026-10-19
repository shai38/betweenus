package couples

type InvitationService struct {
	rep InvitationRepository
}

func NewInvitationService(rep InvitationRepository) *InvitationService {
	return &InvitationService{rep: rep}
}

func (is *InvitationService) CreateInvitation(senderID int64) (Invitation, error) {
	code := GenerateInvitationCode()
	invitation := Invitation{Code: code, CreatedByUserID: senderID}
	err := is.rep.CreateInvitation(invitation)
	if err != nil {
		return Invitation{}, err
	}

	return invitation, nil
}
