package couples

import (
	"errors"
)

type InvitationService struct {
	rep InvitationServiceRepository
}

func NewInvitationService(rep InvitationServiceRepository) *InvitationService {
	return &InvitationService{rep: rep}
}

func (is *InvitationService) CreateInvitation(senderID int64) (Invitation, error) {
	code := GenerateInvitationCode()

	invitation := Invitation{
		Code:            code,
		CreatedByUserID: senderID,
	}

	return is.rep.CreateInvitationIfEligible(invitation)
}

func (is *InvitationService) AcceptInvitation(code string, receiverID int64) (int64, error) {
	return is.rep.AcceptInvitation(code, receiverID)
}

func (is *InvitationService) CanAcceptInvitation(userID int64) error {
	_, err := is.rep.GetCoupleByUserID(userID)

	if err == nil {
		err = ErrOnlyOneCouplePerUser
	} else if errors.Is(err, ErrCoupleNotFound) {
		err = nil
	}

	return err
}
