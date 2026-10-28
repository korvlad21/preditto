// Command mailtest sends one explicit test email without starting the API or database.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"preditto/internal/config"
	"preditto/internal/mailer"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
	log.Print("SMTP server accepted the test message; inbox delivery is not guaranteed")
}

func run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("mailtest", flag.ContinueOnError)
	to := flags.String("to", "", "recipient email address (required)")
	subject := flags.String("subject", "Preditto SMTP test", "message subject")
	text := flags.String("text", "This is a test email from Preditto.", "plain text body")
	html := flags.String("html", "", "optional HTML alternative")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *to == "" || flags.NArg() != 0 {
		return fmt.Errorf("usage: mailtest -to recipient@example.com [-subject ...] [-text ...] [-html ...]")
	}
	cfg, err := config.LoadMail()
	if err != nil {
		return err
	}
	sender, err := mailer.New(cfg)
	if err != nil {
		return err
	}
	return sender.Send(ctx, mailer.Message{To: *to, Subject: *subject, Text: *text, HTML: *html})
}
