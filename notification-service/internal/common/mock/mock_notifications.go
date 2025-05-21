package mockdata

type Recipient struct {
	UserID         string            `json:"userID,omitempty"`
	Email          string            `json:"email,omitempty"`
	Phone          string            `json:"phone,omitempty"`
	WhatsappNumber string            `json:"whatsapp_number,omitempty"`
	Data           map[string]string `json:"data"` // Data inside Recipient
}

type MockNotification struct {
	TemplateID string            `json:"templateID"`
	Meta       map[string]string `json:"meta"`
	Tags       []string          `json:"tags"`
	Recipients []Recipient       `json:"recipients"`
}

var Notifications = []MockNotification{
	{
		TemplateID: "otp_verification",
		Meta:       map[string]string{"source": "auth_service"},
		Tags:       []string{"otp", "security"},
		Recipients: []Recipient{
			{
				Phone: "+919988776655",
				Data: map[string]string{
					"otp":      "982134",
					"username": "Ritu Sharma",
				},
			},
		},
	},
	{
		TemplateID: "nps_feedback",
		Meta:       map[string]string{"source": "feedback_service"},
		Tags:       []string{"nps", "feedback"},
		Recipients: []Recipient{
			{
				Email: "feedback@example.com",
				Data: map[string]string{
					"username":  "Customer A",
					"orderID":   "ORD99223",
					"nps_score": "9",
					"feedback":  "Great service!",
				},
			},
		},
	},
}

