package mailer

import (
	"context"
	"sync/atomic"

	"preditto/internal/config"
)

var defaultMailer atomic.Pointer[Mailer]

// Configure installs the process-wide sender without opening a connection.
// Call it at startup before serving requests. Invalid settings leave the current
// sender unchanged. Concurrent sends keep using the sender they started with.
func Configure(cfg config.MailConfig) error {
	sender, err := New(cfg)
	if err != nil {
		return err
	}
	defaultMailer.Store(sender)
	return nil
}

// Send uses the process-wide sender. Until Configure is called, it returns
// ErrDisabled. Each call owns its connection and respects the caller's context.
func Send(ctx context.Context, message Message) error {
	sender := defaultMailer.Load()
	if sender == nil {
		return ErrDisabled
	}
	return sender.Send(ctx, message)
}
