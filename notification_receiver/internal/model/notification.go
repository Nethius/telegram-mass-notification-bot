package model

type Notification struct {
	Sender       string   `json:"sender"`
	RecipientsId []string `json:"recipients"`
	Message      string   `json:"message"`
}
