package couples

import "errors"

var ErrInvitationNotFound = errors.New("invitation not found")
var ErrCoupleNotFound = errors.New("couple not found")
var ErrSameUser = errors.New("a user cannot form a couple with themselves")
