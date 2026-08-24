package notify

import "testing"

func TestEnvironmentFlagEnabled(t *testing.T) {
	for _, value := range []string{"1", "true", "TRUE", " yes ", "on"} {
		if !environmentFlagEnabled(value) {
			t.Fatalf("environmentFlagEnabled(%q) = false, want true", value)
		}
	}
	for _, value := range []string{"", "0", "false", "no", "off", "debug"} {
		if environmentFlagEnabled(value) {
			t.Fatalf("environmentFlagEnabled(%q) = true, want false", value)
		}
	}
}

func TestNoopNotifier(t *testing.T) {
	if err := (NoopNotifier{}).Notify(Notification{Title: "Tasklight"}); err != nil {
		t.Fatalf("Notify() error = %v, want nil", err)
	}
}
