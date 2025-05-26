//go:build !test
// build +test
package enums

//go:generate stringer -type=AIClassifier
type AIClassifier int

const (
	Llama  AIClassifier = iota
	Gemini AIClassifier = iota + 1
)