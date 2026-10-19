package couples

import (
	"crypto/rand"
	"encoding/hex"
)

type Invitation struct {
	Code            string
	CoupleID        int64
	CreatedByUserID int64
}

func GenerateInvitationCode() string {
	key := make([]byte, 16)
	rand.Read(key)

	return hex.EncodeToString(key)
}
