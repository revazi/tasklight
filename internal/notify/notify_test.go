package notify

import "testing"

func TestNoopNotifier(t *testing.T) {
	if err := (NoopNotifier{}).Notify(Notification{Title: "Tasklight"}); err != nil {
		t.Fatalf("Notify() error = %v, want nil", err)
	}
}
