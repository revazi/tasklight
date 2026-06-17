package notify

// NoopNotifier discards notifications.
type NoopNotifier struct{}

func (NoopNotifier) Notify(Notification) error {
	return nil
}
