package main

import (
	"context"
	"errors"
	"testing"

	"preditto/internal/mailer"
)

func TestRunRequiresExplicitRecipient(t *testing.T) {
	t.Setenv("MAIL_ENABLED", "false")
	for _, args := range [][]string{nil, {"unexpected"}, {"-to", "user@example.com", "unexpected"}} {
		if err := run(context.Background(), args); err == nil || errors.Is(err, mailer.ErrDisabled) {
			t.Fatal("expected usage error before attempting send")
		}
	}
	if err := run(context.Background(), []string{"-to", "user@example.com"}); !errors.Is(err, mailer.ErrDisabled) {
		t.Fatal("disabled sending should fail explicitly", err)
	}
}
