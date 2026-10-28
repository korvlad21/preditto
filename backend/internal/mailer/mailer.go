// Package mailer sends individual messages over authenticated, encrypted SMTP.
package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"strings"

	mail "github.com/wneessen/go-mail"
	"preditto/internal/config"
)

var (
	ErrDisabled       = errors.New("email sending is disabled")
	ErrInvalidMessage = errors.New("invalid email message")
)

type Message struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

// Error keeps SMTP replies out of logs: servers may echo addresses or credentials.
// Unwrap preserves the original cause for errors.Is/errors.As and internal diagnostics.
type Error struct {
	Op   string
	Code int
	err  error
}

func (e *Error) Error() string {
	if errors.Is(e.err, context.DeadlineExceeded) {
		return "SMTP " + e.Op + " timed out"
	}
	if errors.Is(e.err, context.Canceled) {
		return "SMTP " + e.Op + " canceled"
	}
	if e.Code != 0 {
		return fmt.Sprintf("SMTP %s failed (code %d)", e.Op, e.Code)
	}
	return "SMTP " + e.Op + " failed"
}

func (e *Error) Unwrap() error { return e.err }

// Mailer is safe for concurrent Send calls; each call owns its SMTP connection.
type Mailer struct {
	cfg       config.MailConfig
	tlsConfig *tls.Config
}

// New validates settings without connecting to SMTP. A disabled mailer returns
// ErrDisabled on Send, rather than silently dropping messages.
func New(cfg config.MailConfig) (*Mailer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Mailer{cfg: cfg, tlsConfig: &tls.Config{
		ServerName: cfg.Host, MinVersion: tls.VersionTLS12,
	}}, nil
}

// Send returns nil once SMTP accepts the message; it does not confirm inbox delivery.
// Timeout covers the entire operation. There are no automatic retries: a lost
// final SMTP reply can leave delivery uncertain and retrying may create duplicates.
func (m *Mailer) Send(ctx context.Context, message Message) error {
	if !m.cfg.Enabled {
		return ErrDisabled
	}
	msg, err := m.message(message)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, m.cfg.Timeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return err
	}

	// The library's context primarily controls dialing. Closing the connection on
	// the outer context also interrupts greeting, AUTH, DATA and STARTTLS reads.
	var conn net.Conn
	var stop func() bool
	defer func() {
		if stop != nil {
			stop()
		}
		if conn != nil {
			_ = conn.Close()
		}
	}()
	dial := func(dialCtx context.Context, network, address string) (net.Conn, error) {
		dialer := &net.Dialer{}
		var err error
		if m.cfg.TLSMode == "tls" {
			tlsDialer := &tls.Dialer{NetDialer: dialer, Config: m.tlsConfig.Clone()}
			conn, err = tlsDialer.DialContext(dialCtx, network, address)
		} else {
			conn, err = dialer.DialContext(dialCtx, network, address)
		}
		if err != nil {
			return nil, err
		}
		connection := conn
		if secure, ok := connection.(*tls.Conn); ok {
			// Abort the underlying socket directly; TLS Close may try to write a
			// close-notify alert and wait on an unresponsive peer.
			connection = secure.NetConn()
		}
		stop = context.AfterFunc(ctx, func() { _ = connection.Close() })
		return conn, nil
	}
	authType := mail.SMTPAuthPlain
	if m.cfg.AuthType == "login" {
		authType = mail.SMTPAuthLogin
	}
	opts := []mail.Option{
		mail.WithPort(m.cfg.Port), mail.WithTimeout(m.cfg.Timeout),
		mail.WithTLSConfig(m.tlsConfig.Clone()), mail.WithTLSPolicy(mail.TLSMandatory),
		mail.WithSMTPAuth(authType), mail.WithUsername(m.cfg.Username), mail.WithPassword(m.cfg.Password),
		mail.WithDialContextFunc(dial),
	}
	if m.cfg.TLSMode == "tls" {
		opts = append(opts, mail.WithSSL())
	}
	client, err := mail.NewClient(m.cfg.Host, opts...)
	if err != nil {
		return smtpError(ctx, "configure", err)
	}
	if err := client.DialWithContext(ctx); err != nil {
		return smtpError(ctx, "connect/authenticate", err)
	}
	// The library also sends RSET after DATA acceptance. A failure at that point
	// must not turn an accepted message into a reported delivery failure.
	if err := client.Send(msg); err != nil && !msg.IsDelivered() {
		return smtpError(ctx, "send", err)
	}
	// QUIT failure does not undo acceptance of DATA and must not invite a retry.
	_ = client.Close()
	return nil
}

func (m *Mailer) message(input Message) (*mail.Msg, error) {
	if strings.TrimSpace(input.Subject) == "" || strings.TrimSpace(input.Text) == "" ||
		strings.ContainsAny(input.To+input.Subject, "\r\n\x00") {
		return nil, ErrInvalidMessage
	}
	msg := mail.NewMsg()
	if err := msg.FromFormat(m.cfg.FromName, m.cfg.FromAddress); err != nil {
		return nil, ErrInvalidMessage
	}
	if err := msg.To(input.To); err != nil {
		return nil, ErrInvalidMessage
	}
	msg.Subject(input.Subject)
	msg.SetDate()
	msg.SetMessageID()
	msg.SetBodyString(mail.TypeTextPlain, input.Text)
	if input.HTML != "" {
		msg.AddAlternativeString(mail.TypeTextHTML, input.HTML)
	}
	return msg, nil
}

func smtpError(ctx context.Context, op string, err error) error {
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	result := &Error{Op: op, err: err}
	var reply *textproto.Error
	if errors.As(err, &reply) {
		result.Code = reply.Code
	}
	var sendErr *mail.SendError
	if errors.As(err, &sendErr) {
		result.Code = sendErr.ErrorCode()
	}
	return result
}
