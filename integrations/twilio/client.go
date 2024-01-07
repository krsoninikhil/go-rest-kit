package twilio

import (
	"fmt"
	"log"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"
)

type Config struct {
	AccountSID string
	AuthToken  string `log:"-"`
	FromNumber string
}

type Client struct {
	config Config
	sling  *sling.Sling
}

func NewClient(config Config) *Client {
	base := sling.New().Base("https://api.twilio.com/2010-04-01/Accounts/" + config.AccountSID + "/")
	sling := base.New().
		Set("Accept", "application/json").
		Set("Content-Type", "application/x-www-form-urlencoded").
		SetBasicAuth(config.AccountSID, config.AuthToken)
	return &Client{config: config, sling: sling}
}

func (c *Client) SendSMS(toNumber, message string) error {
	req := sendMessageRequest{
		To:   toNumber,
		From: c.config.FromNumber,
		Body: message,
	}
	respError := map[string]any{}
	resp, err := c.sling.New().Post("Messages.json").BodyForm(&req).Receive(nil, &respError)
	if err != nil {
		log.Printf("error sending message to=%s err=%v", toNumber, err)
		return errors.Wrap(err, "error sending message")
	}

	if resp.StatusCode == 201 {
		return nil
	} else {
		log.Printf("failed to send SMS, phone=%s twilio status=%d resp=%v",
			toNumber, resp.StatusCode, respError)
		return fmt.Errorf("failed to send SMS")
	}
}
