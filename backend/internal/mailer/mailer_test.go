package mailer

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"io"
	"math/big"
	"mime"
	"mime/multipart"
	"net"
	"net/mail"
	"net/textproto"
	"strconv"
	"strings"
	"testing"
	"time"

	"preditto/internal/config"
)

type smtpFixture struct {
	mode       string
	reject     string
	stall      string
	noStartTLS bool
	dropAfter  bool
	stage      chan struct{}
	message    chan string
	commands   chan []string
}

// Local SMTP fixture uses a verified test certificate and accepts one connection.
// No external provider, real credentials or database is involved.
func newSMTPFixture(t *testing.T, f *smtpFixture) *Mailer {
	t.Helper()
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:    x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, pub, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	serverTLS := &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	if f.mode == "tls" {
		listener = tls.NewListener(listener, serverTLS)
	}
	f.stage = make(chan struct{})
	f.message = make(chan string, 1)
	f.commands = make(chan []string, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if secure, ok := conn.(*tls.Conn); ok {
			if err := secure.Handshake(); err != nil {
				return
			}
		}
		stall := func(stage string) bool {
			if f.stall != stage {
				return false
			}
			close(f.stage)
			_, _ = io.Copy(io.Discard, conn)
			return true
		}
		if stall("greeting") {
			return
		}
		wire := textproto.NewConn(conn)
		_ = wire.PrintfLine("220 localhost SMTP")
		var commands []string
		defer func() { f.commands <- commands }()
		for {
			line, err := wire.ReadLine()
			if err != nil {
				return
			}
			verb := strings.SplitN(line, " ", 2)[0]
			commands = append(commands, verb)
			if stall(verb) {
				return
			}
			if f.reject == verb {
				_ = wire.PrintfLine("550 rejected test-secret")
				continue
			}
			switch verb {
			case "EHLO":
				_ = wire.PrintfLine("250-localhost")
				if f.mode == "starttls" && !f.noStartTLS {
					_ = wire.PrintfLine("250-STARTTLS")
				}
				_ = wire.PrintfLine("250 AUTH PLAIN LOGIN")
			case "STARTTLS":
				_ = wire.PrintfLine("220 begin TLS")
				secure := tls.Server(conn, serverTLS)
				if err := secure.Handshake(); err != nil {
					return
				}
				conn = secure
				wire = textproto.NewConn(conn)
			case "AUTH":
				if strings.HasPrefix(line, "AUTH PLAIN ") {
					payload, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(line, "AUTH PLAIN "))
					if string(payload) != "\x00sender@example.com\x00test-secret" {
						_ = wire.PrintfLine("535 bad credentials")
						continue
					}
				} else if line == "AUTH LOGIN" {
					_ = wire.PrintfLine("334 VXNlcm5hbWU6")
					user, _ := wire.ReadLine()
					_ = wire.PrintfLine("334 UGFzc3dvcmQ6")
					pass, _ := wire.ReadLine()
					if user != base64.StdEncoding.EncodeToString([]byte("sender@example.com")) || pass != base64.StdEncoding.EncodeToString([]byte("test-secret")) {
						_ = wire.PrintfLine("535 bad credentials")
						continue
					}
				} else {
					_ = wire.PrintfLine("504 unsupported auth")
					continue
				}
				_ = wire.PrintfLine("235 authenticated")
			case "DATA":
				_ = wire.PrintfLine("354 send content")
				body, err := io.ReadAll(wire.DotReader())
				if err != nil {
					return
				}
				f.message <- string(body)
				if stall("data-reply") {
					return
				}
				_ = wire.PrintfLine("250 accepted")
				if f.dropAfter {
					return
				}
			case "QUIT":
				_ = wire.PrintfLine("221 bye")
				return
			default:
				_ = wire.PrintfLine("250 OK")
			}
		}
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		select {
		case <-done:
		case <-time.After(6 * time.Second):
			t.Error("SMTP fixture did not terminate")
		}
	})
	_, port, _ := net.SplitHostPort(listener.Addr().String())
	portNumber, _ := strconv.Atoi(port)
	sender, err := New(config.MailConfig{
		Enabled: true, Host: "127.0.0.1", Port: portNumber, Username: "sender@example.com", Password: "test-secret",
		FromAddress: "sender@example.com", FromName: "Preditto", TLSMode: f.mode, AuthType: "plain", Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(cert)
	sender.tlsConfig.RootCAs = roots
	return sender
}

func testMessage() Message {
	return Message{To: "user@example.com", Subject: "Проверка SMTP", Text: "Привет!\nКод: 123456", HTML: "<p>Привет!</p>"}
}

func TestSendTLSAndSTARTTLS(t *testing.T) {
	for _, mode := range []string{"tls", "starttls"} {
		for _, authType := range []string{"plain", "login"} {
			t.Run(mode+"/"+authType, func(t *testing.T) {
				f := &smtpFixture{mode: mode}
				sender := newSMTPFixture(t, f)
				sender.cfg.AuthType = authType
				if err := sender.Send(context.Background(), testMessage()); err != nil {
					t.Fatal(err)
				}
				msg, err := mail.ReadMessage(strings.NewReader(<-f.message))
				if err != nil {
					t.Fatal(err)
				}
				subject, err := new(mime.WordDecoder).DecodeHeader(msg.Header.Get("Subject"))
				if err != nil || subject != testMessage().Subject || msg.Header.Get("To") != "<user@example.com>" ||
					msg.Header.Get("Date") == "" || msg.Header.Get("Message-ID") == "" {
					t.Fatal("incorrect message headers")
				}
				mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
				if err != nil || mediaType != "multipart/alternative" {
					t.Fatal("expected multipart alternatives", err)
				}
				reader := multipart.NewReader(msg.Body, params["boundary"])
				for _, want := range []string{testMessage().Text, testMessage().HTML} {
					part, err := reader.NextPart()
					if err != nil {
						t.Fatal(err)
					}
					body, err := io.ReadAll(part)
					if err != nil || strings.TrimSuffix(strings.ReplaceAll(string(body), "\r\n", "\n"), "\n") != want {
						t.Fatalf("incorrect UTF-8 body: %q (%v)", body, err)
					}
				}
			})
		}
	}
}

func TestSendTextOnlyAndAcceptedConnectionLoss(t *testing.T) {
	f := &smtpFixture{mode: "tls", dropAfter: true}
	sender := newSMTPFixture(t, f)
	msg := testMessage()
	msg.HTML = ""
	if err := sender.Send(context.Background(), msg); err != nil {
		t.Fatal("post-acceptance disconnect must not report send failure", err)
	}
	parsed, err := mail.ReadMessage(strings.NewReader(<-f.message))
	if err != nil || !strings.HasPrefix(parsed.Header.Get("Content-Type"), "text/plain") {
		t.Fatal("expected plain text MIME", err)
	}
}

func TestSendRejectsSMTPFailureWithoutLeakingReply(t *testing.T) {
	for _, verb := range []string{"AUTH", "RCPT", "DATA"} {
		t.Run(verb, func(t *testing.T) {
			sender := newSMTPFixture(t, &smtpFixture{mode: "tls", reject: verb})
			err := sender.Send(context.Background(), testMessage())
			var sendErr *Error
			if !errors.As(err, &sendErr) || sendErr.Code != 550 || strings.Contains(err.Error(), "test-secret") {
				t.Fatal("expected sanitized SMTP error with status code", err)
			}
		})
	}
}

func TestSendRequiresVerifiedTLS(t *testing.T) {
	for _, tc := range []string{"untrusted", "hostname", "no-starttls"} {
		t.Run(tc, func(t *testing.T) {
			f := &smtpFixture{mode: "tls"}
			if tc == "no-starttls" {
				f.mode, f.noStartTLS = "starttls", true
			}
			sender := newSMTPFixture(t, f)
			if tc == "untrusted" {
				sender.tlsConfig.RootCAs = x509.NewCertPool()
			} else if tc == "hostname" {
				sender.tlsConfig.ServerName = "wrong.example.com"
			}
			if err := sender.Send(context.Background(), testMessage()); err == nil {
				t.Fatal("unsafe TLS accepted")
			}
			if tc == "no-starttls" {
				for _, command := range <-f.commands {
					if command == "AUTH" || command == "MAIL" {
						t.Fatal("credentials or message sent without TLS")
					}
				}
			}
		})
	}
}

func TestSendTimeoutAndCancellation(t *testing.T) {
	for _, stage := range []string{"greeting", "AUTH", "data-reply"} {
		t.Run(stage, func(t *testing.T) {
			f := &smtpFixture{mode: "tls", stall: stage}
			sender := newSMTPFixture(t, f)
			sender.cfg.Timeout = 200 * time.Millisecond
			started := time.Now()
			if err := sender.Send(context.Background(), testMessage()); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatal("expected overall timeout", err)
			}
			if time.Since(started) > time.Second {
				t.Fatal("SMTP operation exceeded timeout bound")
			}
		})
	}
	f := &smtpFixture{mode: "tls", stall: "AUTH"}
	sender := newSMTPFixture(t, f)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { <-f.stage; cancel() }()
	if err := sender.Send(ctx, testMessage()); !errors.Is(err, context.Canceled) {
		t.Fatal("expected cancellation during SMTP AUTH", err)
	}
}

func TestSendDisabledInvalidAndCanceled(t *testing.T) {
	sender, err := New(config.MailConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if err := sender.Send(context.Background(), testMessage()); !errors.Is(err, ErrDisabled) {
		t.Fatal(err)
	}
	sender.cfg.Enabled = true
	sender.cfg.FromAddress = "sender@example.com"
	for _, mutate := range []func(*Message){
		func(m *Message) { m.To = "bad" },
		func(m *Message) { m.To = "a@example.com, b@example.com" },
		func(m *Message) { m.Subject = "subject\r\nBcc: victim@example.com" },
		func(m *Message) { m.Text = "" },
	} {
		msg := testMessage()
		mutate(&msg)
		if err := sender.Send(context.Background(), msg); !errors.Is(err, ErrInvalidMessage) {
			t.Fatal("expected message rejection before network access", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sender.Send(ctx, testMessage()); !errors.Is(err, context.Canceled) {
		t.Fatal("expected pre-canceled request to avoid network access", err)
	}
}
