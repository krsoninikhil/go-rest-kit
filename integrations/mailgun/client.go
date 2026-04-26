package mailgun

import (
	"fmt"
	"log"
	"strings"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"
)

type Config struct {
	APIKey    string `log:"-"`
	Domain    string
	FromEmail string
	BaseURL   string
}

type Client struct {
	config Config
	sling  *sling.Sling
}

func NewClient(config Config) *Client {
	baseURL := strings.TrimSpace(config.BaseURL)
	if baseURL == "" {
		baseURL = "https://api.mailgun.net"
	}
	base := sling.New().Base(strings.TrimRight(baseURL, "/") + "/")
	slingClient := base.New().
		Set("Accept", "application/json").
		Set("Content-Type", "application/x-www-form-urlencoded").
		SetBasicAuth("api", config.APIKey)

	return &Client{config: config, sling: slingClient}
}

func (c *Client) SendEmail(to, subject, message string) error {
	req := sendMessageRequest{
		From:    c.config.FromEmail,
		To:      to,
		Subject: subject,
		Text:    message,
	}
	respError := map[string]any{}
	resp, err := c.sling.New().
		Post(fmt.Sprintf("v3/%s/messages", c.config.Domain)).
		BodyForm(&req).
		Receive(nil, &respError)
	if err != nil {
		log.Printf("mailgun: error sending email to=%s err=%v", to, err)
		return errors.Wrap(err, "error sending email")
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	log.Printf("mailgun: failed to send email to=%s status=%d resp=%v", to, resp.StatusCode, respError)
	return fmt.Errorf("failed to send email")
}
