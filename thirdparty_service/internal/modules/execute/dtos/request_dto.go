package dtos

type RequestDto struct {
	From string `json:"from"`
	To   string `json:"to"`
	Body string `json:"body"`
}
