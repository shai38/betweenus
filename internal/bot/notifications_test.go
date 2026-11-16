package bot

import (
	"context"
	"errors"
	"testing"
	"time"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/shai38/betweenus/internal/couples"
)

type notificationSender struct {
	chatIDs []any
}

type planNotificationSend struct {
	chatID any
	text   string
}

type planNotificationSender struct {
	sends      chan planNotificationSend
	sendErrors []error
	calls      int
}

func (s *planNotificationSender) SendMessage(
	ctx context.Context,
	params *tgBot.SendMessageParams,
) (*models.Message, error) {
	s.sends <- planNotificationSend{
		chatID: params.ChatID,
		text:   params.Text,
	}

	call := s.calls
	s.calls++
	if call < len(s.sendErrors) {
		return nil, s.sendErrors[call]
	}
	return nil, nil
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

func TestPlanNotificationWorkerSendsNotification(t *testing.T) {
	const (
		partnerID int64 = 321
		planTitle       = "Watch a movie"
	)

	sender := &planNotificationSender{
		sends: make(chan planNotificationSend, 1),
	}
	notifications := make(chan planNotification, 1)
	workerDone := make(chan struct{})

	go func() {
		planNotificationWorker(context.Background(), sender, notifications)
		close(workerDone)
	}()

	notifications <- planNotification{
		partnerID: partnerID,
		planTitle: planTitle,
	}
	close(notifications)

	select {
	case got := <-sender.sends:
		wantText := "Твой партнёр добавил новый план: «Watch a movie». 💕"
		if got.chatID != partnerID || got.text != wantText {
			t.Errorf(
				"sent notification = %+v, want chatID %d and text %q",
				got,
				partnerID,
				wantText,
			)
		}
	case <-time.After(time.Second):
		t.Fatal("worker did not send the notification")
	}

	select {
	case <-workerDone:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after the channel was closed")
	}
}

func TestPlanNotificationWorkerContinuesAfterSendError(t *testing.T) {
	const (
		firstPartnerID  int64 = 321
		secondPartnerID int64 = 654
	)

	sender := &planNotificationSender{
		sends:      make(chan planNotificationSend, 2),
		sendErrors: []error{errors.New("telegram unavailable")},
	}
	notifications := make(chan planNotification, 2)
	workerDone := make(chan struct{})

	go func() {
		planNotificationWorker(context.Background(), sender, notifications)
		close(workerDone)
	}()

	notifications <- planNotification{partnerID: firstPartnerID, planTitle: "First plan"}
	notifications <- planNotification{partnerID: secondPartnerID, planTitle: "Second plan"}
	close(notifications)

	wantSends := []planNotificationSend{
		{
			chatID: firstPartnerID,
			text:   "Твой партнёр добавил новый план: «First plan». 💕",
		},
		{
			chatID: secondPartnerID,
			text:   "Твой партнёр добавил новый план: «Second plan». 💕",
		},
	}

	for i, want := range wantSends {
		select {
		case got := <-sender.sends:
			if got != want {
				t.Errorf("send %d = %+v, want %+v", i+1, got, want)
			}
		case <-time.After(time.Second):
			t.Fatalf("worker did not attempt send %d", i+1)
		}
	}

	select {
	case <-workerDone:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after the channel was closed")
	}
}

func TestPlanNotificationWorkerStopsWhenContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	sender := &planNotificationSender{
		sends: make(chan planNotificationSend, 1),
	}
	notifications := make(chan planNotification)
	workerDone := make(chan struct{})

	go func() {
		planNotificationWorker(ctx, sender, notifications)
		close(workerDone)
	}()

	cancel()

	select {
	case <-workerDone:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after the context was cancelled")
	}

	if len(sender.sends) != 0 {
		t.Errorf("sent notifications = %d, want 0", len(sender.sends))
	}
}
