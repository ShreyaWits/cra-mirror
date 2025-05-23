package codes

const (
	TS0001 string = "TS0001"
)

func SuccessMessage(code string) string {
	codesMap := map[string]string{
		"TS0001": "Success",
	}
	return codesMap[code]
}
