package auth

import (
	"context"
	"math/rand"
	"time"

	"github.com/krsoninikhil/go-rest-kit/cache"
	"github.com/pkg/errors"
)

// dependencies
type (
	smsProvider interface {
		Send(phone, message string) error
	}
	cacheClient interface {
		Set(key string, value any, ttl time.Duration) error
		Get(key string) (any, error)
	}
)

type otpMetaData struct {
	OTP     string
	Attempt int
	SentAt  time.Time
}

type otpSvc struct {
	config      OTPConfig
	smsProvider smsProvider
	cache       cacheClient
}

func NewOTPSvc(config OTPConfig, smsProvider smsProvider, cache cacheClient) otpSvc {
	return otpSvc{
		config:      config,
		smsProvider: smsProvider,
		cache:       cache,
	}
}

func (s otpSvc) Send(ctx context.Context, phone string) (*OTPStatus, error) {
	attempt := 1
	lastOTPMeta, err := s.cache.Get(phone)
	if err != nil {
		if !errors.Is(err, cache.ErrKeyNotFound) {
			return nil, errors.Wrap(err, "unable to get last otp")
		}
	} else {
		lastOTP, ok := lastOTPMeta.(otpMetaData)
		if !ok {
			return nil, errors.Wrap(err, "invalid last otp")
		}

		if lastOTP.Attempt >= s.config.MaxAttempts {
			return nil, errors.Wrap(err, "max attempt reached")
		}

		if time.Since(lastOTP.SentAt) < s.config.retryAfter() {
			return nil, errors.Wrap(err, "retrying too soon")
		}
		attempt = lastOTP.Attempt + 1
	}

	otp := generateOTP(s.config.Length)
	if err := s.smsProvider.Send(phone, otp); err != nil {
		return nil, errors.Wrap(err, "unable to send otp")
	}

	otpMeta := otpMetaData{
		OTP:     otp,
		Attempt: attempt,
		SentAt:  time.Now(),
	}
	if err := s.cache.Set(phone, otpMeta, s.config.validity()); err != nil {
		return nil, errors.Wrap(err, "unable to set otp")
	}

	return &OTPStatus{
		RetryAfter:  s.config.RetryAfter,
		AttemptLeft: s.config.MaxAttempts - otpMeta.Attempt,
	}, nil
}

func (s otpSvc) Verify(ctx context.Context, phone, otp string) error {
	lastOTPMeta, err := s.cache.Get(phone)
	if err != nil {
		return errors.Wrap(err, "unable to get last otp")
	}

	lastOTP, ok := lastOTPMeta.(otpMetaData)
	if !ok {
		return errors.Wrap(err, "invalid last otp")
	}

	if time.Since(lastOTP.SentAt) >= s.config.validity() {
		return errors.Wrap(err, "otp expired")
	}

	if lastOTP.OTP != otp {
		return errors.Wrap(err, "invalid otp")
	}

	return nil
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