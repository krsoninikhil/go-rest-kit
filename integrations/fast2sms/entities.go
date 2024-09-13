package fast2sms

type (
	Config struct {
		APIKey  string `validate:"required" log:"-"`
		BaseURL string `validate:"required"`
	}

	sendOTPRequest struct {
		Values  string `url:"variables_values"`
		Route   string `url:"route"`
		Numbers string `url:"numbers"`
	}
)

const smsRouteOTP = "otp"
