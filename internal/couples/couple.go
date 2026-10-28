package couples

type Couple struct {
	ID           int64
	FirstUserID  int64
	SecondUserID int64
}

func (c Couple) PartnerID(userID int64) (int64, error) {
	switch userID {
	case c.FirstUserID:
		return c.SecondUserID, nil
	case c.SecondUserID:
		return c.FirstUserID, nil
	}

	return 0, ErrUserNotInCouple
}
