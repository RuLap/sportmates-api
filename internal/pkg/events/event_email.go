package events

type Event interface {
	GetType() string
}

type EmailEvent struct {
	To       string                 `json:"to"`
	Template string                 `json:"template"`
	Subject  string                 `json:"subject"`
	Data     map[string]interface{} `json:"data"`
}

func (e EmailEvent) GetType() string {
	return "email"
}
