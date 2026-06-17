package notify

// Notifier sends a user-visible notification.
type Notifier interface {
	Notify(notification Notification) error
}

// Notification describes a user-visible notification.
type Notification struct {
	Title        string
	Subtitle     string
	Message      string
	Sound        bool
	ActivateApp  string
	ClickCommand string
	IconPath     string
}
