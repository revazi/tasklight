//go:build !darwin && !linux
// +build !darwin,!linux

package notify

func DefaultNotifier() Notifier {
	return NoopNotifier{}
}
