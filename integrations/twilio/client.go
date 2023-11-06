package twilio

import "github.com/pkg/errors"

package twilio

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dghubble/sling"
)

type Config struct {
	AccountSid string 
	AuthToken  string `log:"-"`
	FromNumber string
}

type Twilio struct {
	config Config
	sling *sling.Sling
}

func NewTwilio(config Config) *Twilio {
	base := sling.New().Base("https://api.twilio.com/2010-04-01/Accounts/" + config.AccountSid + "/")
	sling := base.New().
		Set("Accept", "application/json").
		Set("Content-Type", "application/x-www-form-urlencoded").
		SetBasicAuth(c.accountSid, c.authToken)
	return &Twilio{config: config, sling: sling}
}

func (c *TwilioClient) SendSMS(toNumber, message string) error {
	msgData := c.sling.New().Post("Messages.json").
		BodyForm(strings.NewReader(fmt.Sprintf("To=%s&From=%s&Body=%s", toNumber, c.fromNumber, message)))

	resp, err := msgData.ReceiveSuccess(nil)
	if err != nil {
		return errors.Wrap(err, "error sending message")
	}

	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		return nil
	} else {
		return fmt.Errorf("Failed to send SMS: %s", resp.Status())
	}
}
