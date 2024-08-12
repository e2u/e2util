package e2context

import (
	"context"

	"github.com/sirupsen/logrus"
)

func CheckAndCancelContext(ctx context.Context, cancel context.CancelFunc) {
	// Check if the context is still active
	select {
	case <-ctx.Done():
		logrus.Error("Context is already canceled")
		return
	default:
		logrus.Error("Context is active")
	}

	// Cancel the context
	cancel()
	logrus.Error("Context has been canceled")

	// Verify that the context is now canceled
	select {
	case <-ctx.Done():
		logrus.Error("Confirmed: context is now canceled")
	default:
		logrus.Error("Unexpected: context is still active")
	}
}
