package hanotify

import "smartdisplay-core/internal/logger"

const (
	GuestAccessRequested = "GuestAccessRequested"
	GuestAccessApproved  = "GuestAccessApproved"
	GuestAccessDenied    = "GuestAccessDenied"
	AlarmTriggered       = "AlarmTriggered"
	AlarmRearmed         = "AlarmRearmed"
)

type Notifier interface {
	Notify(ntype string, payload map[string]interface{}) error
}

type StubNotifier struct{}

func (s *StubNotifier) Notify(ntype string, payload map[string]interface{}) error {
	logger.Info("notify: " + ntype)
	return nil
}
