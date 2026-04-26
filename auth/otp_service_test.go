package auth

import (
	"context"
	"regexp"
	"testing"

	"github.com/krsoninikhil/go-rest-kit/cache"
)

type fakeSMSProvider struct {
	lastPhone string
	lastBody  string
}

func (f *fakeSMSProvider) SendSMS(phone, message string) error {
	f.lastPhone = phone
	f.lastBody = message
	return nil
}

type fakeEmailProvider struct {
	lastTo      string
	lastSubject string
	lastBody    string
}

func (f *fakeEmailProvider) SendEmail(to, subject, message string) error {
	f.lastTo = to
	f.lastSubject = subject
	f.lastBody = message
	return nil
}

func Test_generateOTP(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"length 4", 4},
		{"length 6", 6},
		{"length 8", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateOTP(tt.length)
			if len(got) != tt.length {
				t.Fatalf("unexpected otp length: got=%d want=%d", len(got), tt.length)
			}
		})
	}
}

func TestOTPSvc_SMS_DefaultChannelCompatibility(t *testing.T) {
	sms := &fakeSMSProvider{}
	svc := NewOTPSvc(otpConfig{
		ValiditySeconds:   600,
		MaxAttempts:       5,
		RetryAfterSeconds: 30,
		Length:            6,
	}, sms, cache.NewInMemory())

	res, err := svc.Send(context.Background(), "+12345678901", OTPChannelSMS)
	if err != nil {
		t.Fatalf("send otp failed: %v", err)
	}
	if res.RetryAfter != 30 || res.AttemptLeft != 4 {
		t.Fatalf("unexpected otp status: %+v", res)
	}
	if sms.lastBody == "" {
		t.Fatalf("expected sms otp body to be set")
	}

	err = svc.Verify(context.Background(), "+12345678901", sms.lastBody, OTPChannelSMS)
	if err != nil {
		t.Fatalf("verify otp failed: %v", err)
	}
}

func TestOTPSvc_EmailChannel(t *testing.T) {
	sms := &fakeSMSProvider{}
	email := &fakeEmailProvider{}
	svc := NewOTPSvc(otpConfig{
		ValiditySeconds:   600,
		MaxAttempts:       5,
		RetryAfterSeconds: 30,
		Length:            6,
	}, sms, cache.NewInMemory()).WithEmailProvider(email)

	res, err := svc.Send(context.Background(), "reader@example.com", OTPChannelEmail)
	if err != nil {
		t.Fatalf("send email otp failed: %v", err)
	}
	if res.RetryAfter != 30 {
		t.Fatalf("unexpected retry_after: %d", res.RetryAfter)
	}
	if email.lastTo != "reader@example.com" {
		t.Fatalf("unexpected email target: %s", email.lastTo)
	}
	if email.lastSubject != "Your verification code" {
		t.Fatalf("unexpected email subject: %s", email.lastSubject)
	}
	if email.lastBody == "" {
		t.Fatalf("expected otp email body")
	}

	re := regexp.MustCompile(`\d{6}`)
	otp := re.FindString(email.lastBody)
	if len(otp) != 6 {
		t.Fatalf("otp not found in email body: %s", email.lastBody)
	}
	err = svc.Verify(context.Background(), "reader@example.com", otp, OTPChannelEmail)
	if err != nil {
		t.Fatalf("verify email otp failed: %v", err)
	}
}

func TestOTPSvc_ChannelDefaultsToSMS(t *testing.T) {
	sms := &fakeSMSProvider{}
	svc := NewOTPSvc(otpConfig{
		ValiditySeconds:   600,
		MaxAttempts:       5,
		RetryAfterSeconds: 30,
		Length:            6,
	}, sms, cache.NewInMemory())

	_, err := svc.Send(context.Background(), "+12345678901", "")
	if err != nil {
		t.Fatalf("send otp with default channel failed: %v", err)
	}
	if sms.lastBody == "" {
		t.Fatalf("expected sms body for default channel")
	}
}

func TestOTPSvc_TestEmailSkipsSending(t *testing.T) {
	sms := &fakeSMSProvider{}
	email := &fakeEmailProvider{}
	svc := NewOTPSvc(otpConfig{
		ValiditySeconds:   600,
		MaxAttempts:       5,
		RetryAfterSeconds: 30,
		Length:            6,
		TestEmail:         "engineering@faithlabs.io",
	}, sms, cache.NewInMemory()).WithEmailProvider(email)

	_, err := svc.Send(context.Background(), "engineering@faithlabs.io", OTPChannelEmail)
	if err != nil {
		t.Fatalf("send test email otp failed: %v", err)
	}
	if email.lastBody != "" {
		t.Fatalf("expected email provider send to be skipped for test email")
	}
	if err := svc.Verify(context.Background(), "engineering@faithlabs.io", "000000", OTPChannelEmail); err != nil {
		t.Fatalf("verify test email otp failed: %v", err)
	}
}
