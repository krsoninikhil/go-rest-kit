package fast2sms

import (
	"fmt"
	"log"
	"strings"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"
)

type client struct {
	config Config
	sling  *sling.Sling
}

func NewClient(config Config) *client {
	slingClient := sling.New().
		Base(config.BaseURL).
		Set("Content-Type", "application/x-www-form-urlencoded").
		Set("Accept", "application/json").
		Set("Authorization", config.APIKey)

	return &client{
		config: config,
		sling:  slingClient,
	}
}

func (c *client) SendSMS(toNumber, otp string) error {
	toNumber, _ = strings.CutPrefix(toNumber, "+91")
	req := sendOTPRequest{
		Values:  otp,
		Route:   smsRouteOTP,
		Numbers: toNumber,
	}
	respError := map[string]any{}
	resp, err := c.sling.New().Post("/dev/bulkV2").BodyForm(&req).Receive(nil, &respError)
	if err != nil {
		log.Printf("error connecting OTP provided for %s: %v", toNumber, err)
		return errors.Wrap(err, "error sending message")
	}
	fmt.Println("fast2sms response: ", resp.Status, respError)

	if resp.StatusCode != 200 {
		log.Printf("error sending OTP to %s: %v", toNumber, respError)
		return errors.New("error sending OTP")
	}
	return nil
}
