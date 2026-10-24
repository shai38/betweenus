package bot

import (
	"context"
	"errors"
	"testing"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/shai38/betweenus/internal/couples"
)

type notificationSender struct {
	chatIDs []any
}

func (s *notificationSender) SendMessage(
	ctx context.Context,
	params *tgBot.SendMessageParams,
) (*models.Message, error) {
	s.chatIDs = append(s.chatIDs, params.ChatID)
	return nil, nil
}

type notificationCoupleRepository struct {
	couple couples.Couple
	err    error
}

func (r *notificationCoupleRepository) CreateCouple(int64, int64) (int64, error) {
	return r.couple.ID, r.err
}

func (r *notificationCoupleRepository) GetCoupleByID(int64) (couples.Couple, error) {
	return r.couple, r.err
}

func (r *notificationCoupleRepository) GetCoupleByUserID(int64) (couples.Couple, error) {
	return couples.Couple{}, nil
}

func TestNotifyCoupleCreated_SendsNotificationToBothUsers(t *testing.T) {
	sender := &notificationSender{}
	repo := &notificationCoupleRepository{
		couple: couples.Couple{FirstUserID: 101, SecondUserID: 202},
	}

	notifyCoupleCreated(context.Background(), sender, repo, 1)

	if len(sender.chatIDs) != 2 {
		t.Fatalf("number of notifications = %d, want 2", len(sender.chatIDs))
	}
	if sender.chatIDs[0] != int64(101) || sender.chatIDs[1] != int64(202) {
		t.Errorf("notification chat IDs = %v, want [101 202]", sender.chatIDs)
	}
}

func TestNotifyCoupleCreated_DoesNotSendIfCoupleLookupFails(t *testing.T) {
	sender := &notificationSender{}
	repo := &notificationCoupleRepository{err: errors.New("lookup failed")}

	notifyCoupleCreated(context.Background(), sender, repo, 1)

	if len(sender.chatIDs) != 0 {
		t.Errorf("number of notifications = %d, want 0", len(sender.chatIDs))
	}
}
