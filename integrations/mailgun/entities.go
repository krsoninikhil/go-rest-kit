package mailgun

type sendMessageRequest struct {
	From    string `url:"from"`
	To      string `url:"to"`
	Subject string `url:"subject"`
	Text    string `url:"text"`
}
