package mock

import "errors"

var (
	VerifiedPAN = map[string]any{
		"ABCDE1234F": map[string]string{
			"name":     "Rajesh Kumar",
			"dob":      "1985-04-12",
			"pan_type": "Individual",
		},
		"PLMNB6789Q": map[string]string{
			"name":     "Priya Sharma",
			"dob":      "1992-07-30",
			"pan_type": "Individual",
		},
		"XYZAB1122Z": map[string]string{
			"name":     "Amit Singh",
			"dob":      "1978-11-25",
			"pan_type": "Individual",
		},
		"MNOPQ3344R": map[string]string{
			"name":     "Neha Verma",
			"dob":      "2000-03-15",
			"pan_type": "Individual",
		},
		"KLMNO4455S": map[string]string{
			"name":     "Mohammed Ali",
			"dob":      "1995-08-09",
			"pan_type": "Individual",
		},
	}
)

func VerifyPAN(pan string) (bool, map[string]string, error) {
	if value, ok := VerifiedPAN[pan]; ok {
		panData, ok := value.(map[string]string)
		if !ok {
			return false, nil, errors.New("invalid data type")
		}
		return true, panData, nil
	}
	return false, nil, nil
}
