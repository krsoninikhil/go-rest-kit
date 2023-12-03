package twilio

type (
	sendMessageRequest struct {
		To   string `json:"To" url:"To"`
		From string `json:"From" url:"From"`
		Body string `json:"Body" url:"Body"`
	}
)
