package fcm

import (
	"context"
	"log"

	"lockerin-backend/internal/config"
)

type Notifier struct {
	cfg *config.Config
}

func NewNotifier(cfg *config.Config) *Notifier {
	return &Notifier{cfg: cfg}
}

func (n *Notifier) SendPushNotification(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
	if len(tokens) == 0 {
		return nil
	}
	// In production, dispatch via Firebase Admin SDK or FCM HTTP v1 API
	log.Printf("[FCM] Sending push notification to %d device(s): Title='%s' Body='%s'", len(tokens), title, body)
	return nil
}
