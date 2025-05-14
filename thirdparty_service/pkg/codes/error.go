package codes

const (
	TS1001 string = "TS1001"
	TS1002 string = "TS1002"
	TS1003 string = "TS1003"
	TS1004 string = "TS1004"
	TS1005 string = "TS1005"
	TS1006 string = "TS1006"
	TS1007 string = "TS1007"
	TS1008 string = "TS1008"
	TS1009 string = "TS1009"
	TS1010 string = "TS1010"
	TS1011 string = "TS1011"
	TS1012 string = "TS1012"
)

func ErrorMessage(code string) string {
	codesMap := map[string]string{
		"TS1001": "Phone Number is invalid.",
		"TS1002": "Message required for sending sms.",
		"TS1003": "Country code required for sending sms.",
		"TS1004": "Subject required for sending email.",
		"TS1005": "Email required for sending email.",
		"TS1006": "Email content required for sending email.",
		"TS1007": "Error in sending email.",
		"TS1008": "Error in sending sms.",
		"TS1009": "Email content type required for sending email.",
		"TS1010": "Aadhar Number in not valid",
		"TS1011": "Pan Number is not valid",
	}
	return codesMap[code]
}
