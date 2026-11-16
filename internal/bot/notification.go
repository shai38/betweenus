package bot

import (
	"context"
	"fmt"
)

type planNotification struct {
	partnerID int64
	planTitle string
}

func planNotificationWorker(ctx context.Context, b messageSender, ch <-chan planNotification) {
	for {
		select {
		case <-ctx.Done():
			return
		case notification, ok := <-ch:
			if !ok {
				return
			}
			sendText := fmt.Sprintf(messagePartnerPlanAdded, notification.planTitle)
			sendMessage(ctx, b, notification.partnerID, sendText, nil)
		}
	}
}
