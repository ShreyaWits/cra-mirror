package mock

import "errors"

var (
	VerifiedAadhar = map[string]any{
		"123412341234": map[string]string{
			"name": "Rajesh Kumar",
			"dob":  "1985-04-12",
		},
		"234523452345": map[string]string{
			"name": "Priya Sharma",
			"dob":  "1992-07-30",
		},
		"345634563456": map[string]string{
			"name": "Amit Singh",
			"dob":  "1978-11-25",
		},
		"456745674567": map[string]string{
			"name": "Neha Verma",
			"dob":  "2000-03-15",
		},
		"567856785678": map[string]string{
			"name": "Mohammed Ali",
			"dob":  "1995-08-09",
		},
		"678967896789": map[string]string{
			"name": "Sita Rani",
			"dob":  "1980-12-01",
		},
		"789078907890": map[string]string{
			"name": "Vinod Patil",
			"dob":  "1970-06-20",
		},
		"890189018901": map[string]string{
			"name": "Anjali Mehta",
			"dob":  "1999-09-17",
		},
	}
)

func VerifyAadhar(aadhaar string) (bool, map[string]string, error) {
	if value, ok := VerifiedAadhar[aadhaar]; ok {
		aadharData, ok := value.(map[string]string)
		if !ok {
			return false, nil, errors.New("invalid data type")
		}
		return true, aadharData, nil
	}
	return false, nil, nil
}
