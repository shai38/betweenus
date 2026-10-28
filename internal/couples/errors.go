package couples

import "errors"

var ErrInvitationNotFound = errors.New("invitation not found")
var ErrCoupleNotFound = errors.New("couple not found")
var ErrSameUser = errors.New("a user cannot form a couple with themselves")
var ErrOnlyOneCouplePerUser = errors.New("a user can form only one couple")
var ErrUserNotInCouple = errors.New("user is not a member of the couple")
