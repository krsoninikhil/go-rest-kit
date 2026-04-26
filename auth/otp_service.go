package auth

import (
	"context"
	"fmt"
	"math/rand"
	"net/mail"
	"strings"
	"time"

	"github.com/krsoninikhil/go-rest-kit/apperrors"
	"github.com/krsoninikhil/go-rest-kit/cache"
	"github.com/pkg/errors"
)

// dependencies
type (
	smsProvider interface {
		SendSMS(phone, message string) error
	}
	emailProvider interface {
		SendEmail(to, subject, message string) error
	}
	cacheClient interface {
		Set(key string, value any, ttl time.Duration) error
		Get(key string) (any, error)
	}
)

const (
	OTPChannelSMS   = "sms"
	OTPChannelEmail = "email"
	otpEmailSubject = "Your verification code"
)

type otpMetaData struct {
	OTP     string
	Attempt int
	SentAt  time.Time
}

type otpSvc struct {
	config        otpConfig
	cache         cacheClient
	smsProvider   smsProvider
	emailProvider emailProvider
}

func NewOTPSvc(config otpConfig, smsProvider smsProvider, cache cacheClient) otpSvc {
	return otpSvc{
		config:      config,
		cache:       cache,
		smsProvider: smsProvider,
	}
}

func (s otpSvc) WithEmailProvider(provider emailProvider) otpSvc {
	s.emailProvider = provider
	return s
}

func (s otpSvc) Send(ctx context.Context, target, channel string) (*OTPStatus, error) {
	channel = normalizeOTPChannel(channel)
	target = strings.TrimSpace(target)
	attempt := 1
	cacheKey := buildOTPKey(channel, target)

	lastOTPMeta, err := s.cache.Get(cacheKey)
	if err != nil {
		if !errors.Is(err, cache.ErrKeyNotFound) {
			return nil, errors.Wrap(err, "unable to get last otp")
		}
	} else {
		lastOTP, ok := lastOTPMeta.(otpMetaData)
		if !ok {
			return nil, apperrors.NewServerError(errors.New("invalid last otp"))
		}

		if lastOTP.Attempt >= s.config.MaxAttempts {
			return nil, apperrors.NewInvalidParamsError("otp", errors.New("max attempt reached"))
		}

		if time.Since(lastOTP.SentAt) < s.config.retryAfter() {
			return nil, apperrors.NewInvalidParamsError("otp", errors.New("retry too soon"))
		}
		attempt = lastOTP.Attempt + 1
	}

	otp := generateOTP(s.config.Length)
	if channel == OTPChannelSMS && s.config.TestPhone == target {
		otp = testOTP
	} else if channel == OTPChannelEmail && strings.EqualFold(s.config.TestEmail, target) {
		otp = testOTP
	} else if err := s.sendOTPByChannel(channel, target, otp); err != nil {
		return nil, err
	}

	otpMeta := otpMetaData{
		OTP:     otp,
		Attempt: attempt,
		SentAt:  time.Now(),
	}
	if err := s.cache.Set(cacheKey, otpMeta, s.config.validity()); err != nil {
		return nil, errors.Wrap(err, "unable to set otp")
	}

	return &OTPStatus{
		RetryAfter:  s.config.RetryAfterSeconds,
		AttemptLeft: s.config.MaxAttempts - otpMeta.Attempt,
	}, nil
}

func (s otpSvc) Verify(ctx context.Context, target, otp, channel string) error {
	channel = normalizeOTPChannel(channel)
	target = strings.TrimSpace(target)
	lastOTPMeta, err := s.cache.Get(buildOTPKey(channel, target))
	if err != nil {
		if errors.Is(err, cache.ErrKeyNotFound) {
			return apperrors.NewInvalidParamsError("otp", errors.New("otp not sent or expired"))
		}
		return errors.Wrap(err, "unable to get last otp")
	}

	lastOTP, ok := lastOTPMeta.(otpMetaData)
	if !ok {
		return apperrors.NewServerError(errors.New("invalid last otp"))
	}

	if time.Since(lastOTP.SentAt) >= s.config.validity() {
		return apperrors.NewInvalidParamsError("otp", errors.New("otp expired"))
	}

	if lastOTP.OTP != otp {
		return apperrors.NewInvalidParamsError("otp", errors.New("incorrect otp"))
	}

	return nil
}

func (s otpSvc) sendOTPByChannel(channel, target, otp string) error {
	if target == "" {
		return apperrors.NewInvalidParamsError("target", errors.New("target is required"))
	}

	switch channel {
	case OTPChannelSMS:
		if s.smsProvider == nil {
			return apperrors.NewServerError(errors.New("sms provider not configured"))
		}
		if err := s.smsProvider.SendSMS(target, otp); err != nil {
			return errors.Wrap(err, "unable to send otp")
		}
		return nil
	case OTPChannelEmail:
		if s.emailProvider == nil {
			return apperrors.NewServerError(errors.New("email provider not configured"))
		}
		if err := s.emailProvider.SendEmail(target, otpEmailSubject, otpMessage(otp)); err != nil {
			return errors.Wrap(err, "unable to send otp")
		}
		return nil
	default:
		return apperrors.NewInvalidParamsError("channel", fmt.Errorf("unsupported channel: %s", channel))
	}
}

func (r SendOTPRequest) resolveOTPInputs() (string, string, error) {
	target := r.OTPDestination()
	channel := r.OTPChannel()
	if target == "" {
		return "", "", apperrors.NewInvalidParamsError("target", errors.New("phone or target is required"))
	}
	switch channel {
	case OTPChannelSMS:
		if err := validatePhone(target); err != nil {
			return "", "", apperrors.NewInvalidParamsError("phone", err)
		}
	case OTPChannelEmail:
		if _, err := mail.ParseAddress(target); err != nil {
			return "", "", apperrors.NewInvalidParamsError("email", errors.New("invalid email address"))
		}
	default:
		return "", "", apperrors.NewInvalidParamsError("channel", fmt.Errorf("unsupported channel: %s", channel))
	}
	return target, channel, nil
}

func normalizeOTPChannel(channel string) string {
	c := strings.ToLower(strings.TrimSpace(channel))
	if c == "" {
		return OTPChannelSMS
	}
	return c
}

func buildOTPKey(channel, target string) string {
	return "otp:" + channel + ":" + strings.ToLower(strings.TrimSpace(target))
}

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func generateOTP(length int) string {
	digits := "0123456789"
	otp := make([]byte, length)
	for i := range otp {
		otp[i] = digits[r.Intn(len(digits))]
	}

	return string(otp)
}

func otpMessage(otp string) string {
	return "Your OTP code is " + otp + ". It expires shortly."
}

func validatePhone(phone string) error {
	if len(phone) < 9 || len(phone) > 15 {
		return errors.New("invalid phone number")
	}
	if phone[0] != '+' {
		return errors.New("country code is required in phone")
	}
	return nil
}
