package email

import (
	"errors"
	"fmt"
	"log"

	"github.com/resend/resend-go/v2"

	"github.com/rj-davidson/greenrats/internal/config"
)

type Client struct {
	resend    *resend.Client
	fromEmail string
	enabled   bool
}

func New(cfg *config.Config) *Client {
	var resendClient *resend.Client
	if cfg.ResendAPIKey != "" {
		resendClient = resend.NewClient(cfg.ResendAPIKey)
	}

	return &Client{
		resend:    resendClient,
		fromEmail: cfg.FromEmail,
		enabled:   cfg.SendEmails && cfg.ResendAPIKey != "",
	}
}

func (c *Client) Send(to, subject, html string) error {
	if !c.enabled {
		log.Printf("[EMAIL DISABLED] To: %s, Subject: %s", to, subject)
		return nil
	}

	if c.resend == nil {
		return errors.New("resend client not initialized")
	}

	params := &resend.SendEmailRequest{
		From:    c.fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}

	_, err := c.resend.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("[EMAIL SENT] To: %s, Subject: %s", to, subject)
	return nil
}

func (c *Client) SendWelcome(to string, data WelcomeData) error {
	subject := "Welcome to GreenRats!"
	html := renderWelcome(data)
	return c.Send(to, subject, html)
}

func (c *Client) SendLeagueJoin(to string, data LeagueJoinData) error {
	var subject string
	if data.IsNewMember {
		subject = "You joined " + data.LeagueName + "!"
	} else {
		subject = data.NewMemberName + " joined " + data.LeagueName
	}
	html := renderLeagueJoin(data)
	return c.Send(to, subject, html)
}

func (c *Client) SendPickReminder(to string, data PickReminderData) error {
	subject := "Reminder: Make your pick for " + data.TournamentName
	html := renderPickReminder(data)
	return c.Send(to, subject, html)
}

func (c *Client) SendTournamentResults(to string, data *TournamentResultsData) error {
	subject := data.TournamentName + " Results - Your Pick: " + data.GolferName
	html := renderTournamentResults(data)
	return c.Send(to, subject, html)
}

func (c *Client) SendCommissionerAction(to string, data CommissionerActionData) error {
	subject := "[" + data.LeagueName + "] Commissioner Action"
	html := renderCommissionerAction(data)
	return c.Send(to, subject, html)
}
