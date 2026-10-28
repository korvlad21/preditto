package mailer

import (
	"context"
	"errors"
	"sync"
	"testing"

	"preditto/internal/config"
)

// Tests changing process-wide configuration must remain serial.
func TestGlobalSend(t *testing.T) {
	previous := defaultMailer.Swap(nil)
	t.Cleanup(func() { defaultMailer.Store(previous) })
	if err := Send(context.Background(), testMessage()); !errors.Is(err, ErrDisabled) {
		t.Fatal("unconfigured sender must be disabled", err)
	}
	fixture := &smtpFixture{mode: "tls"}
	local := newSMTPFixture(t, fixture)
	if err := Configure(local.cfg); err != nil {
		t.Fatal(err)
	}
	// Trust only this fixture certificate; production uses system roots.
	installed := defaultMailer.Load()
	installed.tlsConfig.RootCAs = local.tlsConfig.RootCAs
	if err := Configure(config.MailConfig{Enabled: true}); err == nil {
		t.Fatal("invalid configuration accepted")
	}
	if defaultMailer.Load() != installed {
		t.Fatal("failed configuration replaced the working sender")
	}
	if err := Send(context.Background(), testMessage()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-fixture.message:
	default:
		t.Fatal("global Send did not deliver to the configured SMTP fixture")
	}
	if err := Configure(config.MailConfig{}); err != nil {
		t.Fatal(err)
	}
	if err := Send(context.Background(), testMessage()); !errors.Is(err, ErrDisabled) {
		t.Fatal("disabled configuration must stop new sends", err)
	}
}

func TestConcurrentGlobalConfiguration(t *testing.T) {
	previous := defaultMailer.Swap(nil)
	t.Cleanup(func() { defaultMailer.Store(previous) })
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for j := 0; j < 20; j++ {
				if err := Configure(config.MailConfig{}); err != nil {
					t.Error(err)
				}
				if err := Send(context.Background(), testMessage()); !errors.Is(err, ErrDisabled) {
					t.Error("unexpected send result", err)
				}
			}
		}()
	}
	workers.Wait()
}
