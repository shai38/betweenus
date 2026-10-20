package couples

import "errors"

type InvitationService struct {
	rep InvitationServiceRepository
}

func NewInvitationService(rep InvitationServiceRepository) *InvitationService {
	return &InvitationService{rep: rep}
}

func (is *InvitationService) CreateInvitation(senderID int64) (Invitation, error) {
	// Проверка на то, чтобы создавалось приглашение только для пользователя без пары
	_, err := is.rep.GetCoupleByUserID(senderID)
	if err == nil {
		return Invitation{}, ErrOnlyOneCouplePerUser
	}
	if !errors.Is(err, ErrCoupleNotFound) {
		return Invitation{}, err
	}

	code := GenerateInvitationCode()
	invitation := Invitation{Code: code, CreatedByUserID: senderID}
	err = is.rep.CreateInvitation(invitation)
	if err != nil {
		return Invitation{}, err
	}

	return invitation, nil
}
