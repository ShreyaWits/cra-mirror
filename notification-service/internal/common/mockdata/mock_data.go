package mockdata

type MockRecipient struct {
	UserID         string
	Email          string
	Phone          string
	WhatsappNumber string
	Data           map[string]string
}

type MockNotification struct {
	TemplateID string
	Meta       map[string]string
	Tags       []string
	Recipients []MockRecipient
}

var Notifications = []MockNotification{
	{
		TemplateID: "test_template",
		Meta: map[string]string{
			"source":   "test",
			"priority": "high",
		},
		Tags: []string{"test"},
		Recipients: []MockRecipient{
			{
				UserID: "user123",
				Email:  "test@example.com",
				Data: map[string]string{
					"name": "Test User",
				},
			},
		},
	},
}
